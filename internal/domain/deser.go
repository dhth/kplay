package domain

import (
	"encoding/json"
	"fmt"

	"github.com/dhth/kplay/internal/domain/generated"
	"github.com/dhth/kplay/internal/utils"
	"github.com/tidwall/pretty"
	"github.com/twmb/franz-go/pkg/kgo"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func GetRecordMetadata(record *kgo.Record) string {
	var msgMetadata string
	var headers string
	var other string
	other += fmt.Sprintf("%s: %s\n", utils.RightPadTrim("timestamp", 20), record.Timestamp)
	other += fmt.Sprintf("%s: %d\n", utils.RightPadTrim("partition", 20), record.Partition)
	other += fmt.Sprintf("%s: %d\n", utils.RightPadTrim("offset", 20), record.Offset)
	for _, h := range record.Headers {
		headers += fmt.Sprintf("%s: %s\n", utils.RightPadTrim(h.Key, 20), string(h.Value))
	}
	if len(record.Headers) > 0 {
		msgMetadata = fmt.Sprintf("%s\nHeaders:\n%s", other, headers)
	} else {
		msgMetadata = other
	}

	return msgMetadata
}

func GetRecordValue(record *kgo.Record) (string, error) {
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

func GetRecordValueJSON(record *kgo.Record) (string, error) {
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
