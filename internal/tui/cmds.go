package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	c "github.com/dhth/kplay/internal/config"
	k "github.com/dhth/kplay/internal/kafka"
	s "github.com/dhth/kplay/internal/serde"
	"github.com/dhth/kplay/internal/utils"
	"github.com/twmb/franz-go/pkg/kgo"
)

func FetchRecords(cl *kgo.Client, commit bool, numRecords int) tea.Cmd {
	return func() tea.Msg {
		records, err := k.FetchMessages(cl, commit, numRecords)
		return msgsFetchedMsg{records, err}
	}
}

func saveRecordDetailsToDisk(record *kgo.Record, details string, notifyUserOnSuccess bool) tea.Cmd {
	return func() tea.Msg {
		filePath := filepath.Join("messages",
			record.Topic,
			fmt.Sprintf("partition-%d", record.Partition),
			fmt.Sprintf("offset-%d.txt", record.Offset),
		)

		dir := filepath.Dir(filePath)
		err := os.MkdirAll(dir, 0o755)
		if err != nil {
			return msgSavedToDiskMsg{err: err}
		}

		err = os.WriteFile(filePath, []byte(details), 0o644)
		if err != nil {
			return msgSavedToDiskMsg{err: err}
		}

		return msgSavedToDiskMsg{path: filePath, notifyUserOnSuccess: notifyUserOnSuccess}
	}
}

func generateRecordDetails(record *kgo.Record, deserializationFmt c.EncodingFormat, protoConfig *c.ProtoConfig) tea.Cmd {
	return func() tea.Msg {
		msgMetadata := utils.GetRecordMetadata(record)
		uniqueKey := utils.GetUniqueKey(record)

		var zeroValue []byte

		if len(record.Value) == 0 {
			return msgDataReadyMsg{uniqueKey, record, messageDetails{msgMetadata, zeroValue, true, nil}}
		}

		var valueBytes []byte
		var err error
		switch deserializationFmt {
		case c.JSON:
			valueBytes, err = s.ParseJSONEncodedBytes(record.Value)
		case c.Protobuf:
			if protoConfig == nil {
				err = fmt.Errorf("%w: protobuf descriptor is nil when it shouldn't be", errSomethingUnexpectedHappened)
			} else {
				valueBytes, err = s.ParseProtobufEncodedBytes(record.Value, protoConfig.MsgDescriptor)
			}
		default:
			valueBytes = record.Value
		}

		if err != nil {
			return msgDataReadyMsg{uniqueKey, record, messageDetails{msgMetadata, zeroValue, false, err}}
		}

		return msgDataReadyMsg{uniqueKey, record, messageDetails{msgMetadata, valueBytes, false, nil}}
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
