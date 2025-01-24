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
		config:            config,
		client:            kCl,
		msgsList:          list.New(jobItems, appDelegate, listWidth, 0),
		currentMsgIndex:   -1,
		persistRecords:    false,
		msgDetailsStore:   make(map[string]messageDetails),
		showHelpIndicator: true,
	}
	m.msgsList.Title = "Messages"
	m.msgsList.SetStatusBarItemName("message", "messages")
	m.msgsList.SetFilteringEnabled(false)
	m.msgsList.DisableQuitKeybindings()
	m.msgsList.SetShowHelp(false)
	m.msgsList.Styles.Title = m.msgsList.Styles.Title.Background(lipgloss.Color(listColor)).
		Foreground(lipgloss.Color(defaultBackgroundColor)).
		Bold(true)
	m.msgsList.KeyMap.PrevPage.SetKeys("left", "h", "pgup")
	m.msgsList.KeyMap.NextPage.SetKeys("right", "l", "pgdown")

	return m
}
