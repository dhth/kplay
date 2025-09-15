package kafka

import (
	"context"
	"errors"

	"github.com/twmb/franz-go/pkg/kgo"
)

func FetchRecords(ctx context.Context, cl *kgo.Client, numRecords uint) ([]*kgo.Record, error) {
	fetches := cl.PollRecords(ctx, int(numRecords))

	err := fetches.Err()
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return fetches.Records(), nil
}
