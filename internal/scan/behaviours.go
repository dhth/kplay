package scan

import (
	"regexp"
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
