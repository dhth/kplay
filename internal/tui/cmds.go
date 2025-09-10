package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	k "github.com/dhth/kplay/internal/kafka"
	t "github.com/dhth/kplay/internal/types"
	"github.com/twmb/franz-go/pkg/kgo"
)

func FetchMessages(cl *kgo.Client, config t.Config, commit bool, numRecords int) tea.Cmd {
	return func() tea.Msg {
		records, err := k.FetchAndCommitRecords(cl, commit, uint(numRecords))
		if err != nil {
			return msgsFetchedMsg{
				err: err,
			}
		}

		messages := make([]t.Message, len(records))
		for i, record := range records {
			messages[i] = t.GetMessageFromRecord(record, config)
		}

		return msgsFetchedMsg{messages, err}
	}
}

func saveRecordDetailsToDisk(message t.Message, topic string, notifyUserOnSuccess bool) tea.Cmd {
	return func() tea.Msg {
		filePath := filepath.Join("messages",
			topic,
			fmt.Sprintf("partition-%d", message.Partition),
			fmt.Sprintf("offset-%d.txt", message.Offset),
		)

		dir := filepath.Dir(filePath)
		err := os.MkdirAll(dir, 0o755)
		if err != nil {
			return msgSavedToDiskMsg{err: err}
		}
		details := getMsgDetails(message)

		err = os.WriteFile(filePath, []byte(details), 0o644)
		if err != nil {
			return msgSavedToDiskMsg{err: err}
		}

		return msgSavedToDiskMsg{path: filePath, notifyUserOnSuccess: notifyUserOnSuccess}
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
