package model

import (
	"time"

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

type DeserializationFmt uint

const (
	JSON DeserializationFmt = iota
	Protobuf
)

type Model struct {
	config              Config
	kCl                 *kgo.Client
	activeView          stateView
	lastView            stateView
	kMsgsList           list.Model
	helpVP              viewport.Model
	helpSeen            uint
	msgMetadataVP       viewport.Model
	msgMetadataVPHeight int
	msgValueVP          viewport.Model
	msgValueVPHeight    int
	recordMetadataStore map[string]string
	recordValueStore    map[string]string
	showHelpIndicator   bool
	skipRecords         bool
	persistRecords      bool
	msgMetadataVPReady  bool
	msgValueVPReady     bool
	helpVPReady         bool
	vpFullScreen        bool
	terminalWidth       int
	terminalHeight      int
	msg                 string
	errorMsg            string
}

func (Model) Init() tea.Cmd {
	return hideHelp(time.Second * 3)
}
