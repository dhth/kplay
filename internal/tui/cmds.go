package tui

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/dhth/kplay/internal/fs"
	k "github.com/dhth/kplay/internal/kafka"
	t "github.com/dhth/kplay/internal/types"
	"github.com/twmb/franz-go/pkg/kgo"
)

func FetchMessages(cl *kgo.Client, config t.Config, commit bool, numRecords int) tea.Cmd {
	return func() tea.Msg {
		records, err := k.FetchAndCommitRecords(cl, commit, numRecords)
		if err != nil {
			return msgsFetchedMsg{
				err: err,
			}
		}

		messages := make([]t.Message, len(records))
		for i, record := range records {
			messages[i] = t.GetMessageFromRecord(record, config, true)
		}

		return msgsFetchedMsg{messages, err}
	}
}

func saveRecordDetailsToDisk(msg t.Message, outputDir, topic string, notifyUserOnSuccess bool) tea.Cmd {
	return func() tea.Msg {
		filePath := filepath.Join(
			outputDir,
			"messages",
			topic,
			fmt.Sprintf("partition-%d", msg.Partition),
			fmt.Sprintf("offset-%d.txt", msg.Offset),
		)
		err := fs.SaveMessageToFileSystem(msg, filePath)
		if err != nil {
			return msgSavedToDiskMsg{err: err}
		}

		return msgSavedToDiskMsg{notifyUserOnSuccess: notifyUserOnSuccess}
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
