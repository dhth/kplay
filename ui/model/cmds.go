package model

import (
	"context"

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

func ShowMsgDetails(item KMsgItem) tea.Cmd {
	return func() tea.Msg {
		return KMsgChosenMsg{item}
	}
}
