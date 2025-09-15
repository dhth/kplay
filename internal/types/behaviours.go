package types

import (
	"fmt"
	"time"
)

type ConsumeBehaviours struct {
	StartOffset      *int64
	StartTimeStamp   *time.Time
	PartitionOffsets map[int32]int64
}

func (b ConsumeBehaviours) Display() string {
	startOffset := NotProvided
	if b.StartOffset != nil {
		startOffset = fmt.Sprintf("%d", *b.StartOffset)
	}

	startTimeStamp := NotProvided
	if b.StartTimeStamp != nil {
		startTimeStamp = b.StartTimeStamp.Format(time.RFC3339)
	}

	partitionOffsets := NotProvided
	if len(b.PartitionOffsets) > 0 {
		partitionOffsets = fmt.Sprintf("%v", b.PartitionOffsets)
	}

	return fmt.Sprintf(`Consume Behaviours:
  start offset            %s
  start timestamp         %s
  partition offsets       %s`,
		startOffset,
		startTimeStamp,
		partitionOffsets,
	)
}
