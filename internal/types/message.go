package types

import (
	"errors"
	"fmt"

	s "github.com/dhth/kplay/internal/serde"
	"github.com/dhth/kplay/internal/utils"
	"github.com/twmb/franz-go/pkg/kgo"
)

var (
	errKafkaRecordIsNil   = errors.New("kafka record is nil when it shouldn't be")
	errProtoDescriptorNil = errors.New("protobuf descriptor is nil when it shouldn't be")
)

var listWidth = 44

var unexpectedErrorMessage = "this is not expected; let @dhth know via https://github.com/dhth/kplay/issues"

type Message struct {
	Metadata  string `json:"metadata"`
	Offset    int64  `json:"offset"`
	Partition int32  `json:"partition"`
	Value     []byte `json:"-"`
	Key       string `json:"key"`
	Err       error  `json:"-"`
}

type SerializableMessage struct {
	Message
	Value *string `json:"value"`
	Err   *string `json:"error"`
}

func (m Message) ToSerializable() SerializableMessage {
	var err *string
	var value *string
	if len(m.Value) > 0 {
		valueStr := string(m.Value)
		value = &valueStr
	}

	if m.Err != nil {
		errStr := m.Err.Error()
		err = &errStr
	}

	return SerializableMessage{
		Message: m,
		Value:   value,
		Err:     err,
	}
}

func GetMessageFromRecord(rec *kgo.Record, config Config) Message {
	if rec == nil {
		return Message{
			Err: fmt.Errorf("%w: %s", errKafkaRecordIsNil, unexpectedErrorMessage),
		}
	}

	record := *rec

	if len(record.Value) == 0 {
		return Message{
			Metadata:  utils.GetRecordMetadata(record),
			Offset:    record.Offset,
			Partition: record.Partition,
			Key:       string(record.Key),
		}
	}

	var bodyBytes []byte
	var err error

	switch config.Encoding {
	case JSON:
		bodyBytes, err = s.ParseJSONEncodedBytes(record.Value)
	case Protobuf:
		if config.Proto == nil {
			err = fmt.Errorf("%w: %s", errProtoDescriptorNil, unexpectedErrorMessage)
		} else {
			bodyBytes, err = s.ParseProtobufEncodedBytes(record.Value, config.Proto.MsgDescriptor)
		}
	case Raw:
		bodyBytes = record.Value
	}

	if err != nil {
		return Message{
			Err: err,
		}
	}

	return Message{
		Metadata:  utils.GetRecordMetadata(record),
		Offset:    record.Offset,
		Partition: record.Partition,
		Value:     bodyBytes,
		Key:       string(record.Key),
	}
}

func (m Message) Title() string {
	if m.Err != nil {
		return "error"
	}

	return utils.RightPadTrim(m.Key, listWidth-4)
}

func (m Message) Description() string {
	if m.Err != nil {
		return ""
	}

	var tombstoneInfo string
	if len(m.Value) == 0 {
		tombstoneInfo = " 🪦"
	}
	offsetInfo := fmt.Sprintf("offset: %d, partition: %d", m.Offset, m.Partition)
	return utils.RightPadTrim(fmt.Sprintf("%s%s", offsetInfo, tombstoneInfo), listWidth-4)
}

func (m Message) FilterValue() string {
	return m.Key
}
