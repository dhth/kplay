package model

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
	"github.com/twmb/franz-go/pkg/kgo"
)

func InitialModel(kCl *kgo.Client, kconfig KConfig) model {

	var appDelegateKeys = newAppDelegateKeyMap()
	appDelegate := newAppItemDelegate(appDelegateKeys)
	jobItems := make([]list.Item, 0)

	m := model{
		kconfig:             kconfig,
		kCl:                 kCl,
		kMsgsList:           list.New(jobItems, appDelegate, listWidth, 0),
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
	m.kMsgsList.Styles.Title.Background(lipgloss.Color(listColor))
	m.kMsgsList.Styles.Title.Foreground(lipgloss.Color(defaultBackgroundColor))
	m.kMsgsList.Styles.Title.Bold(true)

	return m
}
