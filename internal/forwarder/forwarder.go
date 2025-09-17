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
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	a "github.com/dhth/kplay/internal/awsweb"
	k "github.com/dhth/kplay/internal/kafka"
	t "github.com/dhth/kplay/internal/types"

	"github.com/twmb/franz-go/pkg/kgo"
)

type Forwarder struct {
	client   *kgo.Client
	config   t.Config
	s3Client *s3.Client
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

		timeout := time.After(10 * time.Second)
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
		shutDownCtx, shutDownRelease := context.WithTimeout(context.Background(), time.Second*3)
		defer shutDownRelease()
		err := server.Shutdown(shutDownCtx)
		if err != nil {
			slog.Error("error shutting down; trying forceful shutdown", "error", err)

			err := server.Close()
			if err != nil {
				slog.Error("forceful shutdown failed", "error", err)
				return
			}
		}
		slog.Info("server shut down")
	case err := <-serverErrChan:
		slog.Error("couldn't start http server", "error", err)
		return
	}
}

func (f *Forwarder) start(ctx context.Context) {
	resultChan := make(chan a.UploadResult)
	numFetchTokens := 100

	uploadSem := make(chan struct{}, 10)
	for range 10 {
		uploadSem <- struct{}{}
	}

	defer close(resultChan)
	defer close(uploadSem)

	slog.Debug("starting message forwarder")

	for {
		select {
		case <-ctx.Done():
			return
		case <-resultChan:
			numFetchTokens++
		default:
			if numFetchTokens < 100 {
				slog.Debug("fetch tokens below threshold", "num", numFetchTokens)
				continue
			}

			fetchCtx, fetchCancel := context.WithTimeout(ctx, 10*time.Second)
			records, err := k.FetchRecords(fetchCtx, f.client, 100)
			fetchCancel()

			if err != nil {
				slog.Error("couldn't fetch records from Kafka", "error", err)
			}

			for _, record := range records {
				if record == nil {
					continue
				}
				msg := t.GetMessageFromRecord(*record, f.config, true)
				msgBody := msg.GetDetails()

				reader := strings.NewReader(msgBody)

				// guard this behind a semaphore
				numFetchTokens--
				go func() {
					<-uploadSem
					a.UploadToS3(ctx, resultChan, f.s3Client, reader, "test-bucket", "abc")
					uploadSem <- struct{}{}
				}()
			}
		}
	}
}
