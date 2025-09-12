package types

import (
	"fmt"
	"time"
)

type ConsumeBehaviours struct {
	StartOffset    *int64
	StartTimeStamp *time.Time
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

	return fmt.Sprintf(`Consume Behaviours:
  start offset            %s
  start timestamp         %s`,
		startOffset,
		startTimeStamp,
	)
}
