package forwarder

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	k "github.com/dhth/kplay/internal/kafka"
	t "github.com/dhth/kplay/internal/types"

	"github.com/twmb/franz-go/pkg/kgo"
)

const (
	fetchBatchSize                = 100
	shutDownTimeoutSeconds        = 20
	httpServerShutDownTimeoutSecs = 3
)

type Forwarder struct {
	clients  []*kgo.Client
	config   t.Config
	s3Client *s3.Client
}

func New(clients []*kgo.Client, s3Client *s3.Client, config t.Config) Forwarder {
	forwarder := Forwarder{
		clients:  clients,
		config:   config,
		s3Client: s3Client,
	}

	return forwarder
}

func (f *Forwarder) Execute(ctx context.Context) error {
	forwarderCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	sigChan := make(chan os.Signal, 2)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigChan)

	uploadWorkChan := make(chan uploadWork, 10)

	forwarderShutDownChan := make(chan struct{})
	serverShutDownChan := make(chan struct{})

	go func(shutDownChan chan struct{}) {
		f.pollRecords(forwarderCtx, uploadWorkChan)
		shutDownChan <- struct{}{}
	}(forwarderShutDownChan)

	for range 10 {
		go f.startUploadWorker(ctx, uploadWorkChan)
	}

	go func(shutDownChan chan struct{}) {
		startServer(forwarderCtx)
		shutDownChan <- struct{}{}
	}(serverShutDownChan)

	componentsRunning := 2

	select {
	case <-sigChan:
		slog.Info("Received shutdown signal; stopping forwarder and http server")
		cancel()

		timeout := time.After(shutDownTimeoutSeconds * time.Second)
		for componentsRunning > 0 {
			select {
			case <-forwarderShutDownChan:
				componentsRunning--
			case <-serverShutDownChan:
				componentsRunning--
				// on a second signal
			case <-sigChan:
				return nil
				// timeout after first signal
			case <-timeout:
				slog.Error("Couldn't shut down gracefully; exiting")
				return t.ErrCouldntShutDownGracefully
			}
		}
		return nil
	case <-serverShutDownChan:
		cancel()
		<-forwarderShutDownChan
		return nil
	}
}

func startServer(ctx context.Context) {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintf(w, "HEALTHY")
	}))

	addr := "127.0.0.1:8343"

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	serverErrChan := make(chan error)

	go func(errChan chan<- error) {
		slog.Info("starting http server", "address", addr)
		err := server.ListenAndServe()
		if !errors.Is(err, http.ErrServerClosed) {
			errChan <- err
		}
	}(serverErrChan)

	select {
	case <-ctx.Done():
		shutDownCtx, shutDownRelease := context.WithTimeout(context.TODO(), httpServerShutDownTimeoutSecs*time.Second)
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
		slog.Error("couldn't start http server", "error", err)
		return
	}
}

type uploadWork struct {
	clientIndex int
	record      *kgo.Record
}

func (f *Forwarder) pollRecords(ctx context.Context, uploadChan chan<- uploadWork) {
	client := f.clients[0]
	topics := client.GetConsumeTopics()

	slog.Info("starting record poller", "topics", topics)

	pendingRecords := make([]*kgo.Record, 0)

	for {
		select {
		case <-ctx.Done():
			slog.Info("shutting down record poller", "topics", topics)
			return
		default:
			if len(pendingRecords) > 0 {
				var remainingRecords []*kgo.Record
				for _, record := range pendingRecords {
					select {
					case <-ctx.Done():
						slog.Info("shutting down record poller", "topics", topics)
						return
					case uploadChan <- uploadWork{record: record}:
					default:
						remainingRecords = append(remainingRecords, record)
					}
				}
				pendingRecords = remainingRecords
			} else {
				fetchCtx, fetchCancel := context.WithTimeout(ctx, 5*time.Second)
				records, err := k.FetchRecords(fetchCtx, client, fetchBatchSize)
				fetchCancel()

				if err != nil {
					time.Sleep(3 * time.Second)
					slog.Error("couldn't fetch records from Kafka", "error", err)
					continue
				}

				if len(records) == 0 {
					slog.Info("kafka returned no records")
					time.Sleep(3 * time.Second)
					continue
				}

				slog.Info("fetched kafka records", "num_records", len(records))

				for _, record := range records {
					select {
					case <-ctx.Done():
						slog.Info("shutting down record poller", "topics", topics)
						return
					case uploadChan <- uploadWork{record: record}:
					default:
						pendingRecords = append(pendingRecords, record)
					}
				}
			}

		}
	}
}

func (f *Forwarder) startUploadWorker(ctx context.Context, workChan <-chan uploadWork) {
	slog.Info("starting s3 upload worker")

	for {
		select {
		case <-ctx.Done():
			slog.Info("shutting down s3 upload worker")
			return
		case work := <-workChan:
			record := work.record
			objectKey := fmt.Sprintf("%s/partition-%d/offset-%d.txt", f.config.Topic, record.Partition, record.Offset)

			toBreak := false
			var err error
			for i := range 5 {
				msg := t.GetMessageFromRecord(*record, f.config, true)
				bodyReader := strings.NewReader(msg.GetDetails())
				uploadCtx, uploadCancel := context.WithTimeout(ctx, 5*time.Second)
				err = upload(uploadCtx, f.s3Client, bodyReader, "kafka-forwarder-aa18b37f-5d40-4265-82fb-54ec346a4683", objectKey)
				uploadCancel()

				if err != nil {
					if errors.Is(err, context.Canceled) {
						slog.Info("uploading to s3 cancelled", "object_key", objectKey, "attempt_num", i+1)
					} else {
						slog.Error("uploading to s3 failed", "object_key", objectKey, "attempt_num", i+1, "error", err)
					}
				}

				select {
				case <-ctx.Done():
					return
				default:
					if err == nil {
						slog.Info("uploaded to s3", "object_key", objectKey, "attempt_num", i+1)
						toBreak = true
					}
				}
				if toBreak {
					break
				}
			}
		default:
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func upload(ctx context.Context, client *s3.Client, body io.Reader, bucketName, objectKey string) error {
	_, err := client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucketName),
		Key:         aws.String(objectKey),
		Body:        body,
		ContentType: aws.String("text/plain"),
	})

	return err
}
