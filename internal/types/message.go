package types

import (
	"errors"
	"fmt"
	"time"

	s "github.com/dhth/kplay/internal/serde"
	"github.com/dhth/kplay/internal/utils"
	"github.com/twmb/franz-go/pkg/kgo"
)

var errProtoDescriptorNil = errors.New("protobuf descriptor is nil when it shouldn't be")

var unexpectedErrorMessage = "this is not expected; let @dhth know via https://github.com/dhth/kplay/issues"

type Message struct {
	Metadata  string `json:"metadata"`
	Offset    int64  `json:"offset"`
	Partition int32  `json:"partition"`
	Timestamp string `json:"timestamp"`
	Value     []byte `json:"-"`
	Key       string `json:"key"`
	DecodeErr error  `json:"-"`
}

type SerializableMessage struct {
	Message
	Value     *string `json:"value"`
	DecodeErr *string `json:"decode_error"`
}

func (m Message) ToSerializable() SerializableMessage {
	var decodeErr *string
	var value *string
	if len(m.Value) > 0 {
		valueStr := string(m.Value)
		value = &valueStr
	}

	if m.DecodeErr != nil {
		errStr := m.DecodeErr.Error()
		decodeErr = &errStr
	}

	return SerializableMessage{
		Message:   m,
		Value:     value,
		DecodeErr: decodeErr,
	}
}

func (m Message) GetDetails() string {
	var msgValue string
	if len(m.Value) == 0 {
		msgValue = "tombstone"
	} else if m.DecodeErr != nil {
		msgValue = fmt.Sprintf("Decode Error: %s", m.DecodeErr.Error())
	} else {
		msgValue = string(m.Value)
	}

	return fmt.Sprintf(`%s

%s

%s

%s`,
		"Metadata",
		m.Metadata,
		"Value",
		msgValue,
	)
}

func GetMessageFromRecord(record kgo.Record, config Config, decode bool) Message {
	ts := record.Timestamp.Format(time.RFC3339)

	if len(record.Value) == 0 {
		return Message{
			Metadata:  utils.GetRecordMetadata(record),
			Offset:    record.Offset,
			Partition: record.Partition,
			Timestamp: ts,
			Key:       string(record.Key),
		}
	}

	if !decode {
		return Message{
			Metadata:  utils.GetRecordMetadata(record),
			Offset:    record.Offset,
			Partition: record.Partition,
			Timestamp: ts,
			Value:     record.Value,
			Key:       string(record.Key),
		}
	}

	var bodyBytes []byte
	var decodeErr error

	switch config.Encoding {
	case JSON:
		bodyBytes, decodeErr = s.PrettifyJSON(record.Value)
	case Protobuf:
		if config.Proto == nil {
			decodeErr = fmt.Errorf("%w: %s", errProtoDescriptorNil, unexpectedErrorMessage)
		} else {
			bodyBytes, decodeErr = s.TranscodeProto(record.Value, config.Proto.MsgDescriptor)
		}
	case Raw:
		bodyBytes = record.Value
	}

	if decodeErr != nil {
		return Message{
			Metadata:  utils.GetRecordMetadata(record),
			Offset:    record.Offset,
			Partition: record.Partition,
			Timestamp: ts,
			Value:     record.Value,
			Key:       string(record.Key),
			DecodeErr: decodeErr,
		}
	}

	return Message{
		Metadata:  utils.GetRecordMetadata(record),
		Offset:    record.Offset,
		Partition: record.Partition,
		Timestamp: ts,
		Value:     bodyBytes,
		Key:       string(record.Key),
	}
}

func (m Message) Title() string {
	return m.Key
}

func (m Message) Description() string {
	var tombstoneMarker string
	if len(m.Value) == 0 {
		tombstoneMarker = " 🪦"
	}

	var decodeErrorMarker string
	if m.DecodeErr != nil {
		decodeErrorMarker = " (e)"
	}

	return fmt.Sprintf("offset: %d, partition: %d%s%s", m.Offset, m.Partition, decodeErrorMarker, tombstoneMarker)
}

func (m Message) FilterValue() string {
	return m.Key
}
