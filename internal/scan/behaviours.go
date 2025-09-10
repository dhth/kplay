package scan

import "regexp"

type Behaviours struct {
	NumRecords     uint
	OutPathFull    string
	KeyFilterRegex *regexp.Regexp
	BatchSize      uint
}
