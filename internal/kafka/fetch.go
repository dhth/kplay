package kafka

import (
	"context"
	"errors"
	"fmt"

	"github.com/twmb/franz-go/pkg/kgo"
)

var errCouldntCommitRecordsToKafka = errors.New("couldn't commit records to kafka")

func FetchMessages(cl *kgo.Client, commit bool, numRecords int) ([]*kgo.Record, error) {
	fetches := cl.PollRecords(context.TODO(), numRecords)
	records := fetches.Records()
	if commit {
		err := cl.CommitRecords(context.TODO(), records...)
		if err != nil {
			return nil, fmt.Errorf("%w: %s", errCouldntCommitRecordsToKafka, err.Error())
		}
	}

	return records, nil
}
