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
)

type model struct {
	kCl                 *kgo.Client
	activeView          stateView
	kMsgsList           list.Model
	msgMetadataVP       viewport.Model
	msgValueVP          viewport.Model
	recordMetadataStore map[string]string
	recordValueStore    map[string]string
	persistRecords      bool
	filteredKeys        []string
	msgMetadataVPReady  bool
	msgValueVPReady     bool
	vpFullScreen        bool
	terminalWidth       int
	terminalHeight      int
	msg                 string
	errorMsg            string
}

func (m model) Init() tea.Cmd {
	return nil
}
