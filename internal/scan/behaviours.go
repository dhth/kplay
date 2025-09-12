package scan

import (
	"fmt"
	"regexp"

	t "github.com/dhth/kplay/internal/types"
)

const (
	ScanNumRecordsDefault = 1000
)

type Behaviours struct {
	NumMessages    uint
	KeyFilterRegex *regexp.Regexp
	SaveMessages   bool
	Decode         bool
	BatchSize      uint
}

func (b Behaviours) Display() string {
	keyFilterRegex := t.NotProvided
	if b.KeyFilterRegex != nil {
		keyFilterRegex = b.KeyFilterRegex.String()
	}

	value := fmt.Sprintf(`Scan Behaviours:
  number of messages      %d
  key filter regex        %s
  save messages           %v
  decode values           %v
  batch size              %d`,
		b.NumMessages,
		keyFilterRegex,
		b.SaveMessages,
		b.Decode,
		b.BatchSize,
	)

	return value
}
