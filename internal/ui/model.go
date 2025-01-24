package ui

import (
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	d "github.com/dhth/kplay/internal/domain"
	"github.com/twmb/franz-go/pkg/kgo"
)

type stateView uint

const (
	kMsgsListView stateView = iota
	kMsgValueView
	helpView
)

type Model struct {
	config            d.Config
	client            *kgo.Client
	activeView        stateView
	lastView          stateView
	msgsList          list.Model
	currentMsgIndex   int
	firstMsgValueSet  bool
	helpVP            viewport.Model
	msgDetailsVP      viewport.Model
	msgDetailsVPReady bool
	msgDetailsStore   map[string]messageDetails
	showHelpIndicator bool
	skipRecords       bool
	persistRecords    bool
	helpVPReady       bool
	terminalWidth     int
	terminalHeight    int
	msg               string
	errorMsg          string
}

func (Model) Init() tea.Cmd {
	return hideHelp(time.Second * 30)
}
