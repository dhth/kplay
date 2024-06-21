package model

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func newAppDelegateKeyMap() *delegateKeyMap {
	return &delegateKeyMap{
		choose: key.NewBinding(
			key.WithKeys("ctrl+f"),
			key.WithHelp("enter", "check status"),
		),
	}
}

func newAppItemDelegate(keys *delegateKeyMap) list.DefaultDelegate {
	d := list.NewDefaultDelegate()

	d.Styles.SelectedTitle = d.Styles.
		SelectedTitle.
		Foreground(lipgloss.Color(listColor)).
		BorderLeftForeground(lipgloss.Color(listColor))
	d.Styles.SelectedDesc = d.Styles.
		SelectedTitle

	d.UpdateFunc = func(msg tea.Msg, m *list.Model) tea.Cmd {
		switch msgType := msg.(type) {
		case tea.KeyMsg:
			switch {
			case key.Matches(msgType,
				keys.choose,
				list.DefaultKeyMap().CursorUp,
				list.DefaultKeyMap().CursorDown,
				list.DefaultKeyMap().GoToStart,
				list.DefaultKeyMap().GoToEnd,
				list.DefaultKeyMap().NextPage,
				list.DefaultKeyMap().PrevPage,
			):
				if item, ok := m.SelectedItem().(KMsgItem); ok {
					uniqueKey := fmt.Sprintf("-%d-%d", item.record.Partition, item.record.Offset)
					return showItemDetails(uniqueKey)
				} else {
					return nil
				}
			}

		}
		return nil
	}

	help := []key.Binding{keys.choose}

	d.ShortHelpFunc = func() []key.Binding {
		return help
	}

	d.FullHelpFunc = func() [][]key.Binding {
		return [][]key.Binding{help}
	}
	return d
}
