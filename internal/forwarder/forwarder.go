package forwarder

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	k "github.com/dhth/kplay/internal/kafka"
	t "github.com/dhth/kplay/internal/types"

	"github.com/twmb/franz-go/pkg/kgo"
)

const (
	serverShutDownTimeoutMillis = 3000
	reportUploadTimeOutMillis   = 10 * 1000
	numUploadRetryAttempts      = 5
	reportFileTimestampFormat   = "2006-01-02T15-04-05Z"
)

var errServerShutDownUnexpectedly = errors.New("server shut down unexpectedly")

type Forwarder struct {
	kafkaClients []*kgo.Client
	configs      []t.Config
	destination  Destination
	behaviours   Behaviours
}

type uploadWork struct {
	msg      t.Message
	fileName string
}

type uploadResult struct {
	work        uploadWork
	err         error
	numAttempts int
}

func New(kafkaClients []*kgo.Client, configs []t.Config, destination Destination, behaviours Behaviours) Forwarder {
	forwarder := Forwarder{
		kafkaClients: kafkaClients,
		configs:      configs,
		destination:  destination,
		behaviours:   behaviours,
	}

	return forwarder
}

func (f *Forwarder) Execute(ctx context.Context) error {
	forwarderCtx, forwarderCancel := context.WithCancel(ctx)
	defer forwarderCancel()

	sigChan := make(chan os.Signal, 2)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigChan)

	forwarderShutDownChan := make(chan struct{})

	// uninitialized so we don't select over it if the server is not to be run
	// https://www.dolthub.com/blog/2024-10-25-go-nil-channels-pattern/
	var serverShutDownChan chan struct{}

	serverCtx, serverCancel := context.WithCancel(ctx)
	defer serverCancel()

	if f.behaviours.RunServer {
		serverShutDownChan = make(chan struct{})

		go func(shutDownChan chan struct{}) {
			startServer(serverCtx, f.behaviours.ServerHost, f.behaviours.ServerPort, serverShutDownTimeoutMillis)
			shutDownChan <- struct{}{}
		}(serverShutDownChan)
	}

	go func(shutDownChan chan struct{}) {
		f.start(forwarderCtx)
		shutDownChan <- struct{}{}
	}(forwarderShutDownChan)

	select {
	case <-sigChan:
		slog.Info("received shutdown signal; stopping forwarder")
		forwarderCancel()

		timeout := time.After(time.Duration(f.behaviours.ForwarderShutdownTimeoutMillis) * time.Millisecond)

		select {
		case <-forwarderShutDownChan:
		case <-sigChan:
			slog.Error("got a second shutdown signal; exiting right away")
			return nil
		case <-timeout:
			slog.Error("couldn't shut down forwarder gracefully; exiting")
			return t.ErrCouldntShutDownGracefully
		}

		if f.behaviours.RunServer {
			serverCancel()
			select {
			case <-serverShutDownChan:
			case <-sigChan:
				slog.Error("got a second shutdown signal; exiting right away")
				return nil
			case <-timeout:
				slog.Error("couldn't shut down http server gracefully; exiting")
				return t.ErrCouldntShutDownGracefully
			}
		}

		slog.Info("all components stopped gracefully; bye 👋")
		return nil
	case <-serverShutDownChan:
		slog.Error("server shut down unexpectedly; stopping forwarder as well")
		forwarderCancel()

		timeout := time.After(time.Duration(f.behaviours.ForwarderShutdownTimeoutMillis) * time.Millisecond)

		select {
		case <-forwarderShutDownChan:
			slog.Info("all components stopped")
			return errServerShutDownUnexpectedly
		case <-sigChan:
			slog.Error("got a shutdown signal; exiting right away")
			return nil
		case <-timeout:
			slog.Error("couldn't shut down forwarder gracefully; exiting")
			return t.ErrCouldntShutDownGracefully
		}
	}
}

func startServer(ctx context.Context, host string, port uint16, shutdownTimeoutMillis uint16) {
	serverErrChan := make(chan error)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintf(w, "HEALTHY")
	}))

	addr := fmt.Sprintf("%s:%d", host, port)

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	slog.Info("starting http server", "address", addr)

	go func(errChan chan<- error) {
		err := server.ListenAndServe()
		if !errors.Is(err, http.ErrServerClosed) {
			errChan <- err
		}
	}(serverErrChan)

	select {
	case <-ctx.Done():
		shutDownCtx, shutDownRelease := context.WithTimeout(context.WithoutCancel(ctx), time.Duration(shutdownTimeoutMillis)*time.Millisecond)
		defer shutDownRelease()
		err := server.Shutdown(shutDownCtx)
		if err != nil {
			slog.Error("couldn't shut down http server; trying forceful shutdown", "error", err)

			err := server.Close()
			if err != nil {
				slog.Error("forceful shutdown of http server failed", "error", err)
				return
			}
		}
		slog.Info("http server shut down")
	case err := <-serverErrChan:
		slog.Error("http server errored out", "error", err)
		return
	}
}

func (f *Forwarder) start(ctx context.Context) {
	slog.Info("starting forwarder")
	uploadWorkChan := make(chan uploadWork, f.behaviours.NumUploadWorkers)

	var uploadResultChan chan uploadResult

	if f.behaviours.UploadReports {
		uploadResultChan = make(chan uploadResult, f.behaviours.NumUploadWorkers*3)
	}
	var reportDoneChan chan uint64

	uploadCtx, cancelUploadCtx := context.WithCancel(context.WithoutCancel(ctx))
	reportCtx, cancelReportCtx := context.WithCancel(context.WithoutCancel(ctx))

	pendingWork := make([]uploadWork, 0)
	var numRecordsProcessed uint64

	var uploadWg sync.WaitGroup
	slog.Info("starting upload workers", "num", f.behaviours.NumUploadWorkers)
	for range f.behaviours.NumUploadWorkers {
		uploadWg.Add(1)
		go f.startUploadWorker(uploadCtx, uploadWorkChan, uploadResultChan, &uploadWg)
	}

	if f.behaviours.UploadReports {
		reportDoneChan = make(chan uint64)
		go f.startReporterWorker(reportCtx, uploadResultChan, reportDoneChan)
	}

	clientIndex := 0
	for {
		select {
		case <-ctx.Done():
			if len(pendingWork) > 0 {
				slog.Info("sending pending records to upload workers", "num_pending", len(pendingWork))
				for _, work := range pendingWork {
					uploadWorkChan <- work
				}
			}

			slog.Info("waiting for upload workers to finish")
			cancelUploadCtx()
			uploadWg.Wait()
			slog.Info("all upload workers shut down")

			cancelReportCtx()
			if f.behaviours.UploadReports {
				numReportRowsWritten := <-reportDoneChan
				slog.Info("reporter worker shut down", "num_report_rows_written", numReportRowsWritten)
			}

			slog.Info("forwarder shut down", "num_records_processed", numRecordsProcessed)
			return
		default:
			if len(pendingWork) > 0 {
				var remainingWork []uploadWork
				for _, work := range pendingWork {
					select {
					case uploadWorkChan <- work:
					default:
						remainingWork = append(remainingWork, work)
					}
				}
				pendingWork = remainingWork

				time.Sleep(time.Duration(f.behaviours.PollSleepMillis) * time.Millisecond)
			}

			if len(pendingWork) == 0 {
				client := f.kafkaClients[clientIndex]
				fetchCtx, fetchCancel := context.WithTimeout(ctx, time.Duration(f.behaviours.PollFetchTimeoutMillis)*time.Millisecond)
				records, err := k.FetchRecords(fetchCtx, client, uint(f.behaviours.FetchBatchSize))
				fetchCancel()

				if err != nil {
					if !errors.Is(err, context.Canceled) {
						slog.Error("couldn't fetch records from Kafka", "profile", f.configs[clientIndex].Name, "error", err)
					}
				} else if len(records) > 0 {
					for _, record := range records {
						slog.Info("processing record",
							"key", string(record.Key),
							"topic", record.Topic,
							"offset", record.Offset,
							"partition", record.Partition,
							"value_bytes", len(record.Value),
						)
						msg := t.GetMessageFromRecord(*record, f.configs[clientIndex], true)
						work := uploadWork{
							msg:      msg,
							fileName: fmt.Sprintf("%s/partition-%d/offset-%d.txt", record.Topic, record.Partition, record.Offset),
						}
						pendingWork = append(pendingWork, work)
						numRecordsProcessed++
					}
				}
			}

			clientIndex++
			if clientIndex >= len(f.kafkaClients) {
				clientIndex = 0
			}
		}
	}
}

func (f *Forwarder) startUploadWorker(
	ctx context.Context,
	workChan <-chan uploadWork,
	resultChan chan<- uploadResult,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			for {
				select {
				case work := <-workChan:
					f.processUpload(context.WithoutCancel(ctx), work, resultChan)
				default:
					return
				}
			}
		case work := <-workChan:
			f.processUpload(context.WithoutCancel(ctx), work, resultChan)
		}
	}
}

func (f *Forwarder) processUpload(
	ctx context.Context,
	work uploadWork,
	resultChan chan<- uploadResult,
) {
	objectKey := f.destination.getDestinationFilePath(work.fileName)
	result := uploadResult{
		work: work,
	}

	for i := range numUploadRetryAttempts {
		result.numAttempts = i
		bodyReader := strings.NewReader(work.msg.GetDetails())
		uploadCtx, uploadCancel := context.WithTimeout(context.WithoutCancel(ctx), time.Duration(f.behaviours.UploadTimeoutMillis)*time.Millisecond)
		result.err = f.destination.upload(uploadCtx, bodyReader, work.fileName, "text/plain")
		uploadCancel()

		if result.err == nil {
			if result.numAttempts > 0 {
				slog.Info("uploading to s3 succeeded after failures", "object_key", objectKey, "attempt_num", i+1)
			}
			break
		}
		slog.Error("uploading to s3 failed", "object_key", objectKey, "attempt_num", i+1, "error", result.err)
	}

	if resultChan != nil {
		resultChan <- result
	}
}

func (f *Forwarder) startReporterWorker(
	ctx context.Context,
	resultChan <-chan uploadResult,
	doneChan chan<- uint64,
) {
	slog.Info("starting reporter worker")
	var reportWg sync.WaitGroup
	writerStore := make(map[string]*reportWriter)
	var numReportRowsWritten uint64

	handleResult := func(result uploadResult) (*reportWriter, string) {
		topic := result.work.msg.Topic
		writer, ok := writerStore[topic]

		if !ok {
			writer = newReportWriter()
			writerStore[topic] = writer
		}

		err := writer.writeRow(result)
		if err != nil {
			slog.Error("report writer couldn't write row",
				"error", err,
				"file_name", result.work.fileName,
			)
		} else {
			numReportRowsWritten++
		}

		return writer, topic
	}

	uploadResult := func(writer *reportWriter, topic string) {
		reportBytes, err := writer.getContent()
		if err != nil {
			slog.Error("couldn't get bytes from report writer",
				"error", err,
			)
		} else if len(reportBytes) > 0 {
			reportFileObjectKey := fmt.Sprintf("%s/reports/report-%s.csv", topic, writer.startTime.Format(reportFileTimestampFormat))
			reportWg.Add(1)
			go f.uploadReport(ctx, reportBytes, reportFileObjectKey, &reportWg)
		}
		writer.reset()
	}

	for {
		select {
		case <-ctx.Done():
			slog.Info("reporter worker starting shutdown process")
			// the result channel might have entries that haven't been processed yet, drain it first
		drainLoop:
			for {
				select {
				case result := <-resultChan:
					// NOTE: this might lead to a report batch that is bigger than the user's preference
					handleResult(result)
				default:
					break drainLoop
				}
			}

			for topic, writer := range writerStore {
				if writer.numMsgs > 0 {
					uploadResult(writer, topic)
				}
			}

			reportWg.Wait()
			doneChan <- numReportRowsWritten
			return
		case result := <-resultChan:
			writer, topic := handleResult(result)

			if writer.numMsgs >= f.behaviours.ReportBatchSize {
				uploadResult(writer, topic)
			}
		}
	}
}

func (f *Forwarder) uploadReport(ctx context.Context, content []byte, objectKey string, wg *sync.WaitGroup) {
	defer wg.Done()

	for i := range numUploadRetryAttempts {
		reader := bytes.NewReader(content)
		uploadCtx, uploadCancel := context.WithTimeout(context.WithoutCancel(ctx), reportUploadTimeOutMillis*time.Millisecond)
		err := f.destination.upload(uploadCtx, reader, objectKey, "text/csv")
		uploadCancel()

		if err == nil {
			return
		}

		slog.Error("uploading report to s3 failed", "object_key", objectKey, "attempt_num", i+1, "error", err)
	}
}
