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
	record      kgo.Record
	msgMetadata string
	msgValue    string
}

func (item KMsgItem) Title() string {
	return RightPadTrim(string(item.record.Key), listWidth)
}

func (item KMsgItem) Description() string {
	var tombstoneInfo string
	if len(item.record.Value) == 0 {
		tombstoneInfo = " 🪦"
	}
	offsetInfo := fmt.Sprintf("offset: %d, partition: %d", item.record.Partition, item.record.Offset)
	return RightPadTrim(fmt.Sprintf("%s%s", offsetInfo, tombstoneInfo), listWidth)
}

func (item KMsgItem) FilterValue() string {
	return fmt.Sprintf("-%d-%d", item.record.Partition, item.record.Offset)
}
