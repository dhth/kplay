package tui

import (
	t "github.com/dhth/kplay/internal/types"
)

type hideHelpMsg struct{}

type msgsFetchedMsg struct {
	messages []t.Message
	err      error
}

type msgSavedToDiskMsg struct {
	path                string
	notifyUserOnSuccess bool
	err                 error
}

type dataWrittenToClipboard struct {
	err error
}
