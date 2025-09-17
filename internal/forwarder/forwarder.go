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
	client   *kgo.Client
	config   t.Config
	s3Client *s3.Client
}

type uploadResult struct {
	record *kgo.Record
	err    error
}

func New(client *kgo.Client, s3Client *s3.Client, config t.Config) Forwarder {
	forwarder := Forwarder{
		client:   client,
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

	forwarderShutDownChan := make(chan struct{})
	serverShutDownChan := make(chan struct{})

	go func(shutDownChan chan struct{}) {
		f.start(forwarderCtx)
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

func (f *Forwarder) start(ctx context.Context) {
	uploadWorkChan := make(chan *kgo.Record, 10)
	uploadResultChan := make(chan uploadResult, 10)

	numFetchTokens := fetchBatchSize
	numUploadsInProgress := 0

	// uploadSem := make(chan struct{}, 10)

	// defer close(uploadFinishedChan)
	// defer close(uploadSem)

	slog.Info("starting message forwarder")

	var uploadedRecords []*kgo.Record
	for range 10 {
		go f.startUploadWorker(ctx, uploadWorkChan, uploadResultChan)
	}

	for {
		select {
		case <-ctx.Done():
			if numUploadsInProgress == 0 {
				slog.Info("forwarder shut down")
				return
			}

			slog.Info("forwarder finishing up pending uploads", "num_pending", numUploadsInProgress)

			timeout := time.After(10 * time.Second)
			for numUploadsInProgress > 0 {
				select {
				case <-timeout:
					slog.Info("finishing up pending uploads timed out; forwarder shut down", "uploads_still_in_progress", numUploadsInProgress)
					return
				case <-uploadResultChan:
					numFetchTokens++
					numUploadsInProgress--
				}
			}
			return
		case uploadResult := <-uploadResultChan:
			numFetchTokens++
			numUploadsInProgress--
			uploadedRecords = append(uploadedRecords, uploadResult.record)

			if len(uploadedRecords) > 10 {
				commitCtx, commitCancel := context.WithTimeout(context.TODO(), 5*time.Second)
				err := f.client.CommitRecords(commitCtx, uploadedRecords...)
				commitCancel()
				if err != nil && !errors.Is(err, context.Canceled) {
					slog.Error("couldn't commit records to kafka", "error", err)
				}

				uploadedRecords = make([]*kgo.Record, 0)
			}
		default:
			if numFetchTokens < fetchBatchSize {
				continue
			}

			fetchCtx, fetchCancel := context.WithTimeout(ctx, 5*time.Second)
			records, err := k.FetchRecords(fetchCtx, f.client, fetchBatchSize)
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

			toBreak := false
			for _, record := range records {
				if record == nil {
					continue
				}

				select {
				case <-ctx.Done():
					toBreak = true
					// default:
					// 	numFetchTokens--
					// 	numUploadsInProgress++
					// 	go f.uploadToS3(ctx, uploadResultChan, record)
				case uploadWorkChan <- record:
					numFetchTokens--
					numUploadsInProgress++
				}

				if toBreak {
					break
				}
			}
		}
	}
}

// func (f *Forwarder) uploadToS3(ctx context.Context, resultChan chan<- uploadResult, record *kgo.Record) {
// 	msg := t.GetMessageFromRecord(*record, f.config, true)
// 	bodyReader := strings.NewReader(msg.GetDetails())
// 	objectKey := fmt.Sprintf("%s/partition-%d/offset-%d.txt", f.config.Topic, record.Partition, record.Offset)
//
// 	result := uploadResult{
// 		record: record,
// 	}
//
// 	for i := range 5 {
// 		uploadCtx, uploadCancel := context.WithTimeout(ctx, 5*time.Second)
// 		result.err = upload(uploadCtx, f.s3Client, bodyReader, "kafka-forwarder-aa18b37f-5d40-4265-82fb-54ec346a4683", objectKey)
// 		uploadCancel()
//
// 		if result.err != nil {
// 			if errors.Is(result.err, context.Canceled) {
// 				slog.Error("uploading to s3 cancelled", "object_key", objectKey, "attempt_num", i+1)
// 			} else {
// 				slog.Error("uploading to s3 failed", "object_key", objectKey, "attempt_num", i+1, "error", result.err)
// 			}
// 		}
//
// 		select {
// 		case <-ctx.Done():
// 			return
// 		default:
// 			if result.err == nil {
// 				slog.Info("uploaded to s3", "object_key", objectKey, "attempt_num", i+1)
// 				resultChan <- result
// 				return
// 			}
// 		}
// 	}
//
// 	resultChan <- result
// }

func (f *Forwarder) startUploadWorker(ctx context.Context, workChan <-chan *kgo.Record, resultChan chan<- uploadResult) {
	slog.Info("starting s3 upload worker")

	for {
		select {
		case <-ctx.Done():
			slog.Info("shutting down s3 upload worker")
			return
		case record := <-workChan:
			objectKey := fmt.Sprintf("%s/partition-%d/offset-%d.txt", f.config.Topic, record.Partition, record.Offset)

			result := uploadResult{
				record: record,
			}

			toBreak := false
			for i := range 5 {
				msg := t.GetMessageFromRecord(*record, f.config, true)
				bodyReader := strings.NewReader(msg.GetDetails())
				uploadCtx, uploadCancel := context.WithTimeout(ctx, 5*time.Second)
				result.err = upload(uploadCtx, f.s3Client, bodyReader, "kafka-forwarder-aa18b37f-5d40-4265-82fb-54ec346a4683", objectKey)
				uploadCancel()

				if result.err != nil {
					if errors.Is(result.err, context.Canceled) {
						slog.Info("uploading to s3 cancelled", "object_key", objectKey, "attempt_num", i+1)
					} else {
						slog.Error("uploading to s3 failed", "object_key", objectKey, "attempt_num", i+1, "error", result.err)
					}
				}

				select {
				case <-ctx.Done():
					return
				default:
					if result.err == nil {
						slog.Info("uploaded to s3", "object_key", objectKey, "attempt_num", i+1)
						toBreak = true
					}
				}
				if toBreak {
					break
				}
			}
			resultChan <- result
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
