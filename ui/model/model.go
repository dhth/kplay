package model

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/twmb/franz-go/pkg/kgo"
)

type stateView uint

const (
	kMsgsListView stateView = iota
	kMsgMetadataView
	kMsgValueView
	helpView
)

type model struct {
	kCl                 *kgo.Client
	activeView          stateView
	lastView            stateView
	kMsgsList           list.Model
	helpVP              viewport.Model
	helpSeen            uint
	msgMetadataVP       viewport.Model
	msgValueVP          viewport.Model
	recordMetadataStore map[string]string
	recordValueStore    map[string]string
	skipRecords         bool
	persistRecords      bool
	filteredKeys        []string
	msgMetadataVPReady  bool
	msgValueVPReady     bool
	helpVPReady         bool
	vpFullScreen        bool
	terminalWidth       int
	terminalHeight      int
	msg                 string
	errorMsg            string
}

func (m model) Init() tea.Cmd {
	return nil
}
