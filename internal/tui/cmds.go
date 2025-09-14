package tui

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/dhth/kplay/internal/fs"
	t "github.com/dhth/kplay/internal/types"
	"github.com/twmb/franz-go/pkg/kgo"
)

func FetchMessages(cl *kgo.Client, config t.Config, numRecords int) tea.Cmd {
	return func() tea.Msg {
		fetchCtx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
		defer cancel()

		records := cl.PollRecords(fetchCtx, numRecords).Records()

		messages := make([]t.Message, len(records))
		for i, record := range records {
			messages[i] = t.GetMessageFromRecord(record, config, true)
		}

		return msgsFetchedMsg{messages: messages, err: nil}
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
