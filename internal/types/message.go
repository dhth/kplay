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
	Metadata          string    `json:"metadata"`
	Topic             string    `json:"-"`
	Offset            int64     `json:"offset"`
	Partition         int32     `json:"partition"`
	Timestamp         time.Time `json:"-"`
	Value             []byte    `json:"-"`
	Key               string    `json:"key"`
	DecodeErr         error     `json:"-"`
	DecodeErrFallback string    `json:"decode_error_fallback,omitempty"`
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
		var decodeErrFallback string
		if len(m.DecodeErrFallback) > 0 {
			decodeErrFallback = fmt.Sprintf("\n\n%s", m.DecodeErrFallback)
		}
		msgValue = fmt.Sprintf("Decode Error: %s%s", m.DecodeErr.Error(), decodeErrFallback)
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
	msg := Message{
		Metadata:  utils.GetRecordMetadata(record),
		Topic:     record.Topic,
		Offset:    record.Offset,
		Partition: record.Partition,
		Timestamp: record.Timestamp,
		Key:       string(record.Key),
		Value:     record.Value,
	}

	if len(record.Value) == 0 || !decode || config.Encoding == Raw {
		return msg
	}

	var decodedValueBytes []byte
	var decodeErr error
	var decodeErrFallback string

	switch config.Encoding {
	case JSON:
		decodedValueBytes, decodeErr = s.PrettifyJSON(record.Value)
	case Protobuf:
		if config.Proto == nil {
			decodeErr = fmt.Errorf("%w: %s", errProtoDescriptorNil, unexpectedErrorMessage)
		} else {
			decodedValueBytes, decodeErr = s.TranscodeProto(record.Value, config.Proto.MsgDescriptor)
			if decodeErr != nil {
				rawDecodedBytes, rawDecodeErr := s.DecodeRaw(record.Value)
				if rawDecodeErr == nil {
					decodeErrFallback = fmt.Sprintf("Raw decoded value: \n\n%s", rawDecodedBytes)
				}
			}
		}
	}

	if decodeErr != nil {
		msg.DecodeErr = decodeErr
		msg.DecodeErrFallback = decodeErrFallback
	} else {
		msg.Value = decodedValueBytes
	}

	return msg
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
