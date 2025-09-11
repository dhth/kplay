package scan

import (
	"regexp"
	"strings"
)

const (
	ScanFormatTXT         = "txt"
	ScanReportFmtTable    = "table"
	ScanNumRecordsDefault = 1000
)

type Format uint8

const (
	ScanFormatCSV Format = iota
	ScanFormatJSONL
	ScanFormatTxt
)

func ParseScanFormat(value string) (Format, bool) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "csv":
		return ScanFormatCSV, true
	case "jsonl":
		return ScanFormatJSONL, true
	case "txt":
		return ScanFormatTxt, true
	default:
		return ScanFormatCSV, false
	}
}

func (f Format) Extension() string {
	switch f {
	case ScanFormatCSV:
		return "csv"
	case ScanFormatJSONL:
		return "jsonl"
	default:
		return "txt"
	}
}

func ValidScanFormats() []string {
	return []string{
		"csv",
		"jsonl",
		"txt",
	}
}

type Behaviours struct {
	NumMessages    uint
	OutputFormat   Format
	KeyFilterRegex *regexp.Regexp
	SaveMessages   bool
	Decode         bool
	BatchSize      uint
}
