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
	return string(item.record.Key)
}

func (item KMsgItem) Description() string {
	var tombstoneInfo string
	if len(item.record.Value) == 0 {
		tombstoneInfo = " 🪦"
	}
	offsetInfo := fmt.Sprintf("offset: %d", item.record.Offset)
	return fmt.Sprintf("%s%s", offsetInfo, tombstoneInfo)
}

func (item KMsgItem) FilterValue() string {
	return fmt.Sprintf("-%d-%d", item.record.Partition, item.record.Offset)
}
