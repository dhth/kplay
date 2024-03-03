package model

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/twmb/franz-go/pkg/kgo"
)

func InitialModel(kCl *kgo.Client) model {

	var appDelegateKeys = newAppDelegateKeyMap()
	appDelegate := newAppItemDelegate(appDelegateKeys)
	jobItems := make([]list.Item, 0)

	m := model{
		kCl:       kCl,
		kMsgsList: list.New(jobItems, appDelegate, 60, 0),
	}
	m.kMsgsList.Title = "Messages"
	// m.kMsgsList.SetShowTitle(false)
	m.kMsgsList.SetFilteringEnabled(false)
	m.kMsgsList.SetShowHelp(false)

	return m
}
