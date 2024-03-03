package model

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/twmb/franz-go/pkg/kgo"
)

type delegateKeyMap struct {
	choose key.Binding
}

type KMsgItem struct {
	record kgo.Record
}

func (item KMsgItem) Title() string {
	return string(item.record.Key)
}

func (item KMsgItem) Description() string {
	return fmt.Sprintf("offset: %d", item.record.Offset)
}

func (item KMsgItem) FilterValue() string {
	return string(item.record.Key)
}
