package tui

import (
	"github.com/twmb/franz-go/pkg/kgo"
)

type hideHelpMsg struct{}

type msgFetchedMsg struct {
	records []*kgo.Record
	err     error
}

type msgSavedToDiskMsg struct {
	path string
	err  error
}

type msgDataReadyMsg struct {
	uniqueKey string
	details   messageDetails
}

type dataWrittenToClipboard struct {
	err error
}
