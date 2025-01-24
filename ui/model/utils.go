package model

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/dhth/kplay/ui/model/generated"
	"github.com/tidwall/pretty"
	"github.com/twmb/franz-go/pkg/kgo"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func RightPadTrim(s string, length int) string {
	if len(s) >= length {
		if length > 3 {
			return s[:length-3] + "..."
		}
		return s[:length]
	}
	return s + strings.Repeat(" ", length-len(s))
}

func TrimLeft(s string, length int) string {
	if len(s) >= length {
		if length > 3 {
			return "..." + s[len(s)-length+3:]
		}
		return s[:length]
	}
	return s
}

func TrimRight(s string, length int) string {
	if len(s) >= length {
		if length > 3 {
			return s[:length-3] + "..."
		}
		return s[:length]
	}
	return s
}

func getRecordMetadata(record *kgo.Record) string {
	var msgMetadata string
	var headers string
	var other string
	other += fmt.Sprintf("%s: %s\n", RightPadTrim("timestamp", 20), record.Timestamp)
	other += fmt.Sprintf("%s: %d\n", RightPadTrim("partition", 20), record.Partition)
	other += fmt.Sprintf("%s: %d\n", RightPadTrim("offset", 20), record.Offset)
	for _, h := range record.Headers {
		headers += fmt.Sprintf("%s: %s\n", RightPadTrim(h.Key, 20), string(h.Value))
	}
	if len(record.Headers) > 0 {
		msgMetadata = fmt.Sprintf("%s\nHeaders:\n%s", other, headers)
	} else {
		msgMetadata = other
	}

	return msgMetadata
}

func getRecordValue(record *kgo.Record) (string, error) {
	var msgValue string
	if len(record.Value) == 0 {
		msgValue = "Tombstone"
	} else {
		message := &generated.ApplicationState{}
		if err := proto.Unmarshal(record.Value, message); err != nil {
			return "", err
		}
		jsonData, err := protojson.Marshal(message)
		if err != nil {
			return "", err
		}
		nestedPretty := pretty.Pretty(jsonData)
		msgValue = string(pretty.Color(nestedPretty, nil))
	}
	return msgValue, nil
}

func getRecordValueJSON(record *kgo.Record) (string, error) {
	var msgValue string
	if len(record.Value) == 0 {
		msgValue = "Tombstone"
	} else {
		// this is to just ensure that the value is valid JSON
		var data map[string]interface{}
		err := json.Unmarshal(record.Value, &data)
		if err != nil {
			return "", err
		}
		nestedPretty := pretty.Pretty(record.Value)
		msgValue = string(pretty.Color(nestedPretty, nil))
	}
	return msgValue, nil
}
