package model

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/twmb/franz-go/pkg/kgo"
)

func FetchNextKMsg(cl *kgo.Client, numRecords int) tea.Cmd {
	return func() tea.Msg {
		fetches := cl.PollRecords(nil, numRecords)
		records := fetches.Records()
		for _, rec := range records {
			err := cl.CommitRecords(context.Background(), rec)
			if err != nil {
				return KMsgFetchedMsg{
					records: nil,
					err:     err,
				}
			}
		}
		return KMsgFetchedMsg{
			records: fetches.Records(),
			err:     nil,
		}
	}
}

func SaveRecordToDisk(record *kgo.Record, msgMetadata string, msgValue string) tea.Cmd {
	return func() tea.Msg {
		filePath := fmt.Sprintf("records/%s/%d/%d-%s.md",
			record.Topic,
			record.Partition,
			record.Offset,
			record.Key,
		)
		dir := filepath.Dir(filePath)
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return RecordSavedToDiskMsg{err: err}
		}
		var data string
		if len(record.Value) == 0 {
			data = fmt.Sprintf("Metadata\n---\n\n```\n%s```\n\nValue\n---\n\n%s\n", msgMetadata, msgValue)
		} else {
			data = fmt.Sprintf("Metadata\n---\n\n```\n%s```\n\nValue\n---\n\n```json\n%s\n```", msgMetadata, msgValue)
		}
		err = os.WriteFile(filePath, []byte(data), 0644)
		if err != nil {
			return RecordSavedToDiskMsg{err: err}
		}
		return RecordSavedToDiskMsg{path: filePath}
	}
}

func saveRecordMetadata(record *kgo.Record) tea.Cmd {
	return func() tea.Msg {
		msgMetadata := getRecordMetadata(record)
		uniqueKey := fmt.Sprintf("-%d-%d", record.Partition, record.Offset)
		return KMsgMetadataReadyMsg{uniqueKey, msgMetadata}
	}
}

func saveRecordData(record *kgo.Record) tea.Cmd {
	return func() tea.Msg {
		msgMetadata := getRecordMetadata(record)
		msgValue, err := getRecordValue(record)
		if err != nil {
			return KMsgDataReadyMsg{err: err}
		} else {
			uniqueKey := fmt.Sprintf("-%d-%d", record.Partition, record.Offset)
			return KMsgDataReadyMsg{storeKey: uniqueKey, record: record, msgMetadata: msgMetadata, msgValue: msgValue}
		}
	}
}

func showItemDetails(key string) tea.Cmd {
	return func() tea.Msg {
		return KMsgChosenMsg{key}
	}
}
