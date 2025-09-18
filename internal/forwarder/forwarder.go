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
	"sync"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	k "github.com/dhth/kplay/internal/kafka"
	t "github.com/dhth/kplay/internal/types"

	"github.com/twmb/franz-go/pkg/kgo"
)

const (
	fetchBatchSize                 = 50
	numS3Workers                   = 50
	forwarderShutDownTimeoutMillis = 20000
	serverShutDownTimeoutMillis    = 3000
	numPollSleepMillis             = 5000
	uploadWorkerSleepMillis        = 1000
	pollFetchTimeoutMillis         = 5000
	uploadTimeoutMillis            = 5000
)

type Forwarder struct {
	kafkaClients []*kgo.Client
	configs      []t.Config
	s3Client     *s3.Client
	behaviours   Behaviours
}

func New(kafkaClients []*kgo.Client, s3Client *s3.Client, configs []t.Config, behaviours Behaviours) Forwarder {
	forwarder := Forwarder{
		kafkaClients: kafkaClients,
		configs:      configs,
		s3Client:     s3Client,
		behaviours:   behaviours,
	}

	return forwarder
}

func (f *Forwarder) Execute(ctx context.Context) error {
	forwarderCtx, forwarderCancel := context.WithCancel(ctx)
	serverCtx, serverCancel := context.WithCancel(ctx)
	defer forwarderCancel()
	defer serverCancel()

	sigChan := make(chan os.Signal, 2)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigChan)

	uploadWorkChan := make(chan uploadWork, 10)

	forwarderShutDownChan := make(chan struct{})
	serverShutDownChan := make(chan struct{})

	go func(shutDownChan chan struct{}) {
		startServer(serverCtx, f.behaviours.Host, f.behaviours.Port)
		shutDownChan <- struct{}{}
	}(serverShutDownChan)

	go func(shutDownChan chan struct{}) {
		f.start(forwarderCtx, uploadWorkChan)
		shutDownChan <- struct{}{}
	}(forwarderShutDownChan)

	select {
	case <-sigChan:
		slog.Info("received shutdown signal; stopping forwarder")
		forwarderCancel()

		timeout := time.After(forwarderShutDownTimeoutMillis * time.Millisecond)

		select {
		case <-forwarderShutDownChan:
		case <-sigChan:
			slog.Error("got a second shutdown signal; shutting down right away")
			return nil
		case <-timeout:
			slog.Error("couldn't shut down forwarder gracefully; exiting")
			return t.ErrCouldntShutDownGracefully
		}

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

func startServer(ctx context.Context, host string, port uint) {
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
		shutDownCtx, shutDownRelease := context.WithTimeout(context.TODO(), serverShutDownTimeoutMillis*time.Millisecond)
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

func (f *Forwarder) start(ctx context.Context, uploadWorkChan chan uploadWork) {
	slog.Info("starting forwarder")

	pendingWork := make([]uploadWork, 0)

	uploadCtx, uploadCancel := context.WithCancel(context.TODO())
	var wg sync.WaitGroup

	slog.Info("starting s3 upload workers", "num", numS3Workers)
	for range numS3Workers {
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
			slog.Info("all S3 workers shut down")
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
				fetchCtx, fetchCancel := context.WithTimeout(ctx, pollFetchTimeoutMillis*time.Millisecond)
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
						pendingWork = append(pendingWork, work)
					}
				}

				time.Sleep(numPollSleepMillis * time.Millisecond)
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
			for i := range 5 {
				bodyReader := strings.NewReader(work.msg.GetDetails())
				uploadCtx, uploadCancel := context.WithTimeout(context.TODO(), uploadTimeoutMillis*time.Millisecond)
				err = upload(uploadCtx, f.s3Client, bodyReader, f.behaviours.BucketName, work.objectKey)
				uploadCancel()

				if err != nil {
					slog.Error("uploading to s3 failed", "object_key", work.objectKey, "attempt_num", i+1, "error", err)
				}

				if err == nil {
					slog.Info("uploaded to s3", "object_key", work.objectKey, "attempt_num", i+1)
					break
				}
			}
		default:
			time.Sleep(uploadWorkerSleepMillis * time.Millisecond)
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
