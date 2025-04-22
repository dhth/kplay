package utils

import (
	"fmt"
	"strings"

	"github.com/twmb/franz-go/pkg/kgo"
)

const (
	metadataKeyPadding = 20
)

func GetRecordMetadata(record kgo.Record) string {
	var lines []string // nolint:prealloc
	if len(record.Key) > 0 {
		lines = append(lines, fmt.Sprintf("- %s %s", RightPadTrim("key", metadataKeyPadding), record.Key))
	}
	lines = append(lines, fmt.Sprintf("- %s %s", RightPadTrim("timestamp", metadataKeyPadding), record.Timestamp))
	lines = append(lines, fmt.Sprintf("- %s %d", RightPadTrim("partition", metadataKeyPadding), record.Partition))
	lines = append(lines, fmt.Sprintf("- %s %d", RightPadTrim("offset", metadataKeyPadding), record.Offset))

	for _, h := range record.Headers {
		lines = append(lines, fmt.Sprintf("- %s %s", RightPadTrim(h.Key, metadataKeyPadding), h.Value))
	}

	return strings.Join(lines, "\n")
}
