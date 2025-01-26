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
	"google.golang.org/protobuf/reflect/protoreflect"
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

func generateRecordDetails(record *kgo.Record, deserializationFmt c.EncodingFormat, protoMsgDescriptor *protoreflect.MessageDescriptor) tea.Cmd {
	return func() tea.Msg {
		msgMetadata := utils.GetRecordMetadata(record)
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
		case c.JSON:
			valueBytes, err = s.ParseJSONEncodedBytes(record.Value)
		case c.Protobuf:
			if protoMsgDescriptor == nil {
				err = fmt.Errorf("%w: protobuf descriptor is nil when it shouldn't be", errSomethingUnexpectedHappened)
			} else {
				valueBytes, err = s.ParseProtobufEncodedBytes(record.Value, *protoMsgDescriptor)
			}
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
