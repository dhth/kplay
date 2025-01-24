package model

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/twmb/franz-go/pkg/kgo"
)

func FetchRecords(cl *kgo.Client, numRecords int) tea.Cmd {
	return func() tea.Msg {
		fetches := cl.PollRecords(context.Background(), numRecords)
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

func saveRecordMetadataToDisk(record *kgo.Record, msgMetadata string) tea.Cmd {
	return func() tea.Msg {
		filePath := fmt.Sprintf("records/%s/%d/%d-%s-metadata.md",
			record.Topic,
			record.Partition,
			record.Offset,
			record.Key,
		)
		dir := filepath.Dir(filePath)
		err := os.MkdirAll(dir, 0o755)
		if err != nil {
			return RecordSavedToDiskMsg{err: err}
		}
		data := fmt.Sprintf("Metadata\n---\n\n```\n%s```", msgMetadata)
		err = os.WriteFile(filePath, []byte(data), 0o644)
		if err != nil {
			return RecordSavedToDiskMsg{err: err}
		}
		return RecordSavedToDiskMsg{path: filePath}
	}
}

func saveRecordValueToDisk(record *kgo.Record) tea.Cmd {
	return func() tea.Msg {
		filePath := fmt.Sprintf("records/%s/%d/%d-%s-value.md",
			record.Topic,
			record.Partition,
			record.Offset,
			record.Key,
		)
		dir := filepath.Dir(filePath)
		err := os.MkdirAll(dir, 0o755)
		if err != nil {
			return RecordSavedToDiskMsg{err: err}
		}
		var data string
		if len(record.Value) == 0 {
			data = fmt.Sprintf("Value\n---\n\n%s\n", "Tombstone")
		} else {
			data = fmt.Sprintf("Value\n---\n\n```json\n%s\n```", string(record.Value))
		}
		err = os.WriteFile(filePath, []byte(data), 0o644)
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
		return KMsgMetadataReadyMsg{storeKey: uniqueKey, record: record, msgMetadata: msgMetadata}
	}
}

func saveRecordValue(record *kgo.Record, deserializationFmt DeserializationFmt) tea.Cmd {
	return func() tea.Msg {
		var msgValue string
		var err error
		switch deserializationFmt {
		case JSON:
			msgValue, err = getRecordValueJSON(record)
		case Protobuf:
			msgValue, err = getRecordValue(record)
		}
		if err != nil {
			return KMsgValueReadyMsg{err: err}
		}
		uniqueKey := fmt.Sprintf("-%d-%d", record.Partition, record.Offset)
		return KMsgValueReadyMsg{storeKey: uniqueKey, record: record, msgValue: msgValue}
	}
}

func showItemDetails(key string) tea.Cmd {
	return func() tea.Msg {
		return KMsgChosenMsg{key}
	}
}

func hideHelp(interval time.Duration) tea.Cmd {
	return tea.Tick(interval, func(time.Time) tea.Msg {
		return HideHelpMsg{}
	})
}
