package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	d "github.com/dhth/kplay/internal/domain"
	k "github.com/dhth/kplay/internal/kafka"
	"github.com/twmb/franz-go/pkg/kgo"
)

func FetchRecords(cl *kgo.Client, numRecords int) tea.Cmd {
	return func() tea.Msg {
		records, err := k.FetchMessages(cl, true, numRecords)
		return msgFetchedMsg{records, err}
	}
}

func saveRecordValueToDisk(uniqueKey string, value string) tea.Cmd {
	return func() tea.Msg {
		filePath := fmt.Sprintf("records/%s.txt", uniqueKey)
		dir := filepath.Dir(filePath)
		err := os.MkdirAll(dir, 0o755)
		if err != nil {
			return msgSavedToDiskMsg{err: err}
		}
		err = os.WriteFile(filePath, []byte(value), 0o644)
		if err != nil {
			return msgSavedToDiskMsg{err: err}
		}
		return msgSavedToDiskMsg{path: filePath}
	}
}

func generateRecordDetails(record *kgo.Record, deserializationFmt d.DeserializationFmt) tea.Cmd {
	return func() tea.Msg {
		msgMetadata := d.GetRecordMetadata(record)
		uniqueKey := fmt.Sprintf("records/%s/%d/%d-%s",
			record.Topic,
			record.Partition,
			record.Offset,
			record.Key,
		)

		var zeroValue []byte

		if len(record.Value) == 0 {
			return msgDataReadyMsg{uniqueKey, messageDetails{msgMetadata, zeroValue, true, nil}}
		}

		var valueBytes []byte
		var err error
		switch deserializationFmt {
		case d.JSON:
			valueBytes, err = d.GetPrettyJSON(record.Value)
		case d.Protobuf:
			valueBytes, err = d.GetPrettyJSONFromProtoBytes(record.Value)
		default:
			valueBytes = record.Value
		}

		if err != nil {
			return msgDataReadyMsg{uniqueKey, messageDetails{msgMetadata, zeroValue, false, err}}
		}

		return msgDataReadyMsg{uniqueKey, messageDetails{msgMetadata, valueBytes, false, nil}}
	}
}

func hideHelp(interval time.Duration) tea.Cmd {
	return tea.Tick(interval, func(time.Time) tea.Msg {
		return hideHelpMsg{}
	})
}

func copyToClipboard(data string) tea.Cmd {
	return func() tea.Msg {
		err := clipboard.WriteAll(data)
		return dataWrittenToClipboard{err}
	}
}
