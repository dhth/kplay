package forwarder

import (
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
)

type Forwarder struct {
	kafkaClients []*kgo.Client
	configs      []t.Config
	destination  Destination
	behaviours   Behaviours
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

	uploadWorkChan := make(chan uploadWork, 10)
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
		f.start(forwarderCtx, uploadWorkChan)
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
			slog.Error("got a second shutdown signal; shutting down right away")
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
				slog.Error("got a second shutdown signal; shutting down right away")
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
		<-forwarderShutDownChan
		slog.Error("all components stopped")
		return nil
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
		shutDownCtx, shutDownRelease := context.WithTimeout(context.TODO(), time.Duration(shutdownTimeoutMillis)*time.Millisecond)
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
	msg      t.Message
	fileName string
}

func (f *Forwarder) start(ctx context.Context, uploadWorkChan chan uploadWork) {
	slog.Info("starting forwarder")

	pendingWork := make([]uploadWork, 0)

	uploadCtx, uploadCancel := context.WithCancel(context.TODO())
	var wg sync.WaitGroup

	slog.Info("starting upload workers")
	for range f.behaviours.NumUploadWorkers {
		wg.Add(1)
		go f.startUploadWorker(uploadCtx, uploadWorkChan, &wg)
	}

	clientIndex := 0
	for {
		select {
		case <-ctx.Done():
			if len(pendingWork) > 0 {
				slog.Info("finishing uploads for pending records", "num_pending", len(pendingWork))
				for _, work := range pendingWork {
					uploadWorkChan <- work
				}
			}

			slog.Info("waiting for upload workers to finish")
			uploadCancel()
			wg.Wait()
			slog.Info("all upload workers shut down")
			slog.Info("forwarder shut down")
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
			}

			if len(pendingWork) == 0 {
				client := f.kafkaClients[clientIndex]
				fetchCtx, fetchCancel := context.WithTimeout(ctx, time.Duration(f.behaviours.PollFetchTimeoutMillis)*time.Millisecond)
				records, err := k.FetchRecords(fetchCtx, client, uint(f.behaviours.FetchBatchSize))
				fetchCancel()

				if err != nil {
					slog.Error("couldn't fetch records from Kafka", "profile", f.configs[clientIndex].Name, "error", err)
				} else if len(records) > 0 {
					for _, record := range records {
						slog.Info("processing kafka record", "key", string(record.Key), "topic", record.Topic, "partition", record.Partition, "offset", record.Offset, "value_bytes", len(record.Value))
						msg := t.GetMessageFromRecord(*record, f.configs[clientIndex], true)
						work := uploadWork{
							msg:      msg,
							fileName: fmt.Sprintf("%s/partition-%d/offset-%d.txt", record.Topic, record.Partition, record.Offset),
						}
						pendingWork = append(pendingWork, work)
					}
				}

				time.Sleep(time.Duration(f.behaviours.PollSleepMillis) * time.Millisecond)
			}

			clientIndex++
			if clientIndex >= len(f.kafkaClients) {
				clientIndex = 0
			}
		}
	}
}

func (f *Forwarder) startUploadWorker(ctx context.Context, workChan <-chan uploadWork, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case work := <-workChan:
			var err error
			objectKey := f.destination.getDestinationFilePath(work.fileName)

			for i := range 5 {
				bodyReader := strings.NewReader(work.msg.GetDetails())
				uploadCtx, uploadCancel := context.WithTimeout(context.TODO(), time.Duration(f.behaviours.UploadTimeoutMillis)*time.Millisecond)
				err = f.destination.upload(uploadCtx, bodyReader, work.fileName)
				uploadCancel()

				if err != nil {
					slog.Error("uploading to s3 failed", "object_key", objectKey, "attempt_num", i+1, "error", err)
				}

				if err == nil {
					break
				}
			}
		default:
			time.Sleep(time.Duration(f.behaviours.UploadWorkerSleepMillis) * time.Millisecond)
		}
	}
}
