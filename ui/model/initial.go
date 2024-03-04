package model

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/twmb/franz-go/pkg/kgo"
)

func InitialModel(kCl *kgo.Client, deserFmt DeserializationFmt) model {

	var appDelegateKeys = newAppDelegateKeyMap()
	appDelegate := newAppItemDelegate(appDelegateKeys)
	jobItems := make([]list.Item, 0)

	m := model{
		deserializationFmt:  deserFmt,
		kCl:                 kCl,
		kMsgsList:           list.New(jobItems, appDelegate, 60, 0),
		persistRecords:      false,
		recordMetadataStore: make(map[string]string),
		recordValueStore:    make(map[string]string),
	}
	m.kMsgsList.Title = "Messages"
	m.kMsgsList.SetFilteringEnabled(false)
	m.kMsgsList.SetShowHelp(false)

	return m
}
