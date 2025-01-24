package ui

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
	d "github.com/dhth/kplay/internal/domain"
	"github.com/twmb/franz-go/pkg/kgo"
)

func InitialModel(kCl *kgo.Client, config d.Config) Model {
	appDelegateKeys := newAppDelegateKeyMap()
	appDelegate := newAppItemDelegate(appDelegateKeys)
	jobItems := make([]list.Item, 0)

	m := Model{
		config:              config,
		kCl:                 kCl,
		kMsgsList:           list.New(jobItems, appDelegate, listWidth, 0),
		kMsgsCurrentIndex:   -1,
		persistRecords:      false,
		recordMetadataStore: make(map[string]string),
		recordValueStore:    make(map[string]string),
		showHelpIndicator:   true,
	}
	m.kMsgsList.Title = "Messages"
	m.kMsgsList.SetStatusBarItemName("message", "messages")
	m.kMsgsList.SetFilteringEnabled(false)
	m.kMsgsList.DisableQuitKeybindings()
	m.kMsgsList.SetShowHelp(false)
	m.kMsgsList.Styles.Title = m.kMsgsList.Styles.Title.Background(lipgloss.Color(listColor)).
		Foreground(lipgloss.Color(defaultBackgroundColor)).
		Bold(true)

	return m
}
