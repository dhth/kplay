package ui

import (
	"fmt"

	"github.com/dhth/kplay/internal/utils"
	"github.com/twmb/franz-go/pkg/kgo"
)

type KMsgItem struct {
	record kgo.Record
}

func (item KMsgItem) Title() string {
	return utils.RightPadTrim(string(item.record.Key), listWidth-4)
}

func (item KMsgItem) Description() string {
	var tombstoneInfo string
	if len(item.record.Value) == 0 {
		tombstoneInfo = " 🪦"
	}
	offsetInfo := fmt.Sprintf("offset: %d, partition: %d", item.record.Partition, item.record.Offset)
	return utils.RightPadTrim(fmt.Sprintf("%s%s", offsetInfo, tombstoneInfo), listWidth-4)
}

func (item KMsgItem) FilterValue() string {
	return fmt.Sprintf("records/%s/%d/%d-%s",
		item.record.Topic,
		item.record.Partition,
		item.record.Offset,
		item.record.Key,
	)
}

type messageDetails struct {
	metadata  string
	value     []byte
	tombstone bool
	err       error
}
