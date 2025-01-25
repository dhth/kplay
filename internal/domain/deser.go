package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/dhth/kplay/internal/domain/generated"
	"github.com/dhth/kplay/internal/utils"
	"github.com/tidwall/pretty"
	"github.com/twmb/franz-go/pkg/kgo"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

const (
	metadataKeyPadding = 20
)

var (
	errCouldntUnmarshalJSONData      = errors.New("couldn't unmarshal JSON encoded bytes")
	errCouldntUnmarshalWireFormatMsg = errors.New("couldn't unmarshal wire format message")
	errCouldntConvertProtoMsgToJSON  = errors.New("couldn't convert proto message to JSON")
)

func GetRecordMetadata(record *kgo.Record) string {
	var lines []string // nolint:prealloc
	if len(record.Key) > 0 {
		lines = append(lines, fmt.Sprintf("- %s %s", utils.RightPadTrim("key", metadataKeyPadding), record.Key))
	}
	lines = append(lines, fmt.Sprintf("- %s %s", utils.RightPadTrim("timestamp", metadataKeyPadding), record.Timestamp))
	lines = append(lines, fmt.Sprintf("- %s %d", utils.RightPadTrim("partition", metadataKeyPadding), record.Partition))
	lines = append(lines, fmt.Sprintf("- %s %d", utils.RightPadTrim("offset", metadataKeyPadding), record.Offset))

	for _, h := range record.Headers {
		lines = append(lines, fmt.Sprintf("%s: %s", utils.RightPadTrim(h.Key, metadataKeyPadding), string(h.Value)))
	}

	return strings.Join(lines, "\n")
}

func GetPrettyJSON(bytes []byte) ([]byte, error) {
	var data map[string]interface{}
	err := json.Unmarshal(bytes, &data)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", errCouldntUnmarshalJSONData, err.Error())
	}

	return pretty.Pretty(bytes), nil
}

func GetPrettyJSONFromProtoBytes(bytes []byte) ([]byte, error) {
	message := &generated.ApplicationState{}
	if err := proto.Unmarshal(bytes, message); err != nil {
		return nil, fmt.Errorf("%w: %s", errCouldntUnmarshalWireFormatMsg, err.Error())
	}
	jsonData, err := protojson.Marshal(message)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", errCouldntConvertProtoMsgToJSON, err.Error())
	}
	return pretty.Pretty(jsonData), nil
}
