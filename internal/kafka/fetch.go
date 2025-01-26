package kafka

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
)

var errCouldntCommitMessagesToKafka = errors.New("couldn't commit messages to kafka")

func FetchMessages(cl *kgo.Client, commit bool, numRecords int) ([]*kgo.Record, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	fetches := cl.PollRecords(ctx, numRecords)
	records := fetches.Records()
	if len(records) == 0 {
		return records, nil
	}

	if commit {
		err := cl.CommitRecords(ctx, records...)
		if err != nil {
			return nil, fmt.Errorf("%w: %s", errCouldntCommitMessagesToKafka, err.Error())
		}
	}

	return records, nil
}
