package tui

import (
	"github.com/twmb/franz-go/pkg/kgo"
)

type hideHelpMsg struct{}

type msgsFetchedMsg struct {
	records []*kgo.Record
	err     error
}

type msgSavedToDiskMsg struct {
	path                string
	notifyUserOnSuccess bool
	err                 error
}

type msgDataReadyMsg struct {
	uniqueKey string
	record    *kgo.Record
	details   messageDetails
}

type dataWrittenToClipboard struct {
	err error
}
