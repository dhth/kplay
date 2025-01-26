package tui

import (
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	c "github.com/dhth/kplay/internal/config"
	"github.com/twmb/franz-go/pkg/kgo"
)

type stateView uint

const (
	msgListView stateView = iota
	msgDetailsView
	helpView
)

type Model struct {
	config             c.Config
	client             *kgo.Client
	activeView         stateView
	lastView           stateView
	msgsList           list.Model
	currentMsgIndex    int
	firstMsgValueSet   bool
	fetchingInProgress bool
	helpVP             viewport.Model
	msgDetailsVP       viewport.Model
	msgDetailsVPReady  bool
	msgDetailsStore    map[string]messageDetails
	showHelpIndicator  bool
	behaviours         c.Behaviours
	helpVPReady        bool
	terminalWidth      int
	terminalHeight     int
	msg                string
	errorMsg           string
}

func (Model) Init() tea.Cmd {
	return hideHelp(time.Second * 30)
}
