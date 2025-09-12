package tui

import (
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	t "github.com/dhth/kplay/internal/types"
	"github.com/twmb/franz-go/pkg/kgo"
)

type stateView uint

const (
	msgListView stateView = iota
	msgDetailsView
	helpView
)

type Model struct {
	config             t.Config
	client             *kgo.Client
	activeView         stateView
	lastView           stateView
	msgsList           list.Model
	currentMsgIndex    int
	fetchingInProgress bool
	helpVP             viewport.Model
	msgDetailsVP       viewport.Model
	msgDetailsVPReady  bool
	showHelpIndicator  bool
	outputDir          string
	behaviours         t.TUIBehaviours
	helpVPReady        bool
	terminalWidth      int
	terminalHeight     int
	msg                string
	errorMsg           string
}

func (Model) Init() tea.Cmd {
	return hideHelp(time.Second * 30)
}
