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
	fetchBatchSize                = 50
	shutDownTimeoutSeconds        = 20
	httpServerShutDownTimeoutSecs = 3
)

type Forwarder struct {
	clients  []*kgo.Client
	configs  []t.Config
	s3Client *s3.Client
}

func New(clients []*kgo.Client, s3Client *s3.Client, configs []t.Config) Forwarder {
	forwarder := Forwarder{
		clients:  clients,
		configs:  configs,
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
				slog.Error("couldn't shut down gracefully; exiting")
				return t.ErrCouldntShutDownGracefully
			}
		}
		slog.Info("all components stopped gracefully; bye 👋")
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
	msg       t.Message
	objectKey string
}

func (f *Forwarder) pollRecords(ctx context.Context, uploadChan chan uploadWork) {
	slog.Info("starting record poller")

	pendingWork := make([]uploadWork, 0)

	uploadCtx, uploadCancel := context.WithCancel(context.TODO())

	for range 50 {
		go f.startUploadWorker(uploadCtx, uploadChan)
	}

	clientIndex := 0
	for {
		select {
		case <-ctx.Done():
			slog.Info("finishing uploads for pending records")
			for _, work := range pendingWork {
				uploadChan <- work
			}

			uploadCancel()
			slog.Info("shut down record poller")
			return
		default:
			if len(pendingWork) > 0 {
				var remainingWork []uploadWork
				for _, work := range pendingWork {
					select {
					case uploadChan <- work:
					default:
						remainingWork = append(remainingWork, work)
					}
				}
				pendingWork = remainingWork
			}

			if len(pendingWork) == 0 {
				client := f.clients[clientIndex]
				fetchCtx, fetchCancel := context.WithTimeout(ctx, 5*time.Second)
				records, err := k.FetchRecords(fetchCtx, client, fetchBatchSize)
				fetchCancel()

				if err != nil {
					slog.Error("couldn't fetch records from Kafka", "profile", f.configs[clientIndex].Name, "error", err)
				} else if len(records) > 0 {
					slog.Info("fetched kafka records", "profile", f.configs[clientIndex].Name, "num_records", len(records))

					for _, record := range records {
						msg := t.GetMessageFromRecord(*record, f.configs[clientIndex], true)
						work := uploadWork{
							msg:       msg,
							objectKey: fmt.Sprintf("%s/partition-%d/offset-%d.txt", record.Topic, record.Partition, record.Offset),
						}
						// select {
						// case uploadChan <- work:
						// default:
						pendingWork = append(pendingWork, work)
						// }
					}
				}
			}

			clientIndex++
			if clientIndex >= len(f.clients) {
				clientIndex = 0
			}

			time.Sleep(100 * time.Millisecond)
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

			var err error
			for i := range 5 {
				bodyReader := strings.NewReader(work.msg.GetDetails())
				uploadCtx, uploadCancel := context.WithTimeout(context.TODO(), 5*time.Second)
				err = upload(uploadCtx, f.s3Client, bodyReader, "kafka-forwarder-aa18b37f-5d40-4265-82fb-54ec346a4683", work.objectKey)
				uploadCancel()

				if err != nil {
					if errors.Is(err, context.Canceled) {
						slog.Info("uploading to s3 cancelled", "object_key", work.objectKey, "attempt_num", i+1)
					} else {
						slog.Error("uploading to s3 failed", "object_key", work.objectKey, "attempt_num", i+1, "error", err)
					}
				}

				if err == nil {
					slog.Info("uploaded to s3", "object_key", work.objectKey, "attempt_num", i+1)
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
