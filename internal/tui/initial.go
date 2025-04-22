package tui

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
	t "github.com/dhth/kplay/internal/types"
	"github.com/twmb/franz-go/pkg/kgo"
)

func InitialModel(kCl *kgo.Client, config t.Config, behaviours t.TUIBehaviours) Model {
	appDelegateKeys := newAppDelegateKeyMap()
	appDelegate := newAppItemDelegate(appDelegateKeys)
	jobItems := make([]list.Item, 0)

	m := Model{
		config:            config,
		client:            kCl,
		msgsList:          list.New(jobItems, appDelegate, listWidth, 0),
		currentMsgIndex:   -1,
		behaviours:        behaviours,
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
