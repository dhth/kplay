package forward

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	a "github.com/dhth/kplay/internal/aws"
	k "github.com/dhth/kplay/internal/kafka"
	t "github.com/dhth/kplay/internal/types"

	"github.com/twmb/franz-go/pkg/kgo"
)

type Forwarder struct {
	client   *kgo.Client
	s3Client *s3.Client
	config   t.Config
}

func New(client *kgo.Client, s3Client *s3.Client, config t.Config) Forwarder {
	scanner := Forwarder{
		client: client,
		config: config,
	}

	return scanner
}

func (f *Forwarder) Start(ctx context.Context) error {
	resultChan := make(chan a.UploadResult)
	numFetchTokens := 100

	for {
		select {
		case <-ctx.Done():
			return nil
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
				go a.UploadToS3(ctx, resultChan, f.s3Client, reader, "test-bucket", "abc")
			}
		}
	}
}
