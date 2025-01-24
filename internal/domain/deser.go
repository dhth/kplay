package domain

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/dhth/kplay/internal/domain/generated"
	"github.com/dhth/kplay/internal/utils"
	"github.com/tidwall/pretty"
	"github.com/twmb/franz-go/pkg/kgo"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func GetRecordMetadata(record *kgo.Record) string {
	var lines []string // nolint:prealloc
	if len(record.Key) > 0 {
		lines = append(lines, fmt.Sprintf("%s: %s", utils.RightPadTrim("key", 20), record.Key))
	}
	lines = append(lines, fmt.Sprintf("%s: %s", utils.RightPadTrim("timestamp", 20), record.Timestamp))
	lines = append(lines, fmt.Sprintf("%s: %d", utils.RightPadTrim("partition", 20), record.Partition))
	lines = append(lines, fmt.Sprintf("%s: %d", utils.RightPadTrim("offset", 20), record.Offset))

	for _, h := range record.Headers {
		lines = append(lines, fmt.Sprintf("%s: %s", utils.RightPadTrim(h.Key, 20), string(h.Value)))
	}

	return strings.Join(lines, "\n")
}

func GetRecordValue(record *kgo.Record) (string, error) {
	if len(record.Value) == 0 {
		return "Tombstone", nil
	}

	message := &generated.ApplicationState{}
	if err := proto.Unmarshal(record.Value, message); err != nil {
		return "", err
	}
	jsonData, err := protojson.Marshal(message)
	if err != nil {
		return "", err
	}
	prettyJSON := pretty.Pretty(jsonData)
	return string(prettyJSON), nil
}

func GetRecordValueJSON(record *kgo.Record) (string, error) {
	if len(record.Value) == 0 {
		return "Tombstone", nil
	}

	// this is to just ensure that the value is valid JSON
	var data map[string]interface{}
	err := json.Unmarshal(record.Value, &data)
	if err != nil {
		return "", err
	}

	prettyJSON := pretty.Pretty(record.Value)
	return string(prettyJSON), nil
}
