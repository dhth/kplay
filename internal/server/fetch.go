package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	c "github.com/dhth/kplay/internal/config"
	k "github.com/dhth/kplay/internal/kafka"
	s "github.com/dhth/kplay/internal/serde"
	"github.com/dhth/kplay/internal/utils"
	"github.com/twmb/franz-go/pkg/kgo"
)

const (
	contentType     = "Content-Type"
	applicationJSON = "application/json; charset=utf-8"
	unexpected      = "something unexpected happened (let @dhth know about this via https://github.com/dhth/kplay/issues)"
)

type KafkaMessage struct {
	Key       string  `json:"key"`
	Offset    int64   `json:"offset"`
	Partition int32   `json:"partition"`
	Metadata  string  `json:"metadata"`
	Value     *string `json:"value"`
	Tombstone bool    `json:"tombstone"`
	Err       error   `json:"error"`
}

func getMessages(client *kgo.Client, config c.Config) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		queryParams := r.URL.Query()
		numMessagesStr := queryParams.Get("num")

		numMessages := 1
		if numMessagesStr != "" {
			num, err := strconv.Atoi(numMessagesStr)
			if err == nil && num > 1 {
				numMessages = num
			}
		}

		records, err := k.FetchMessages(client, true, numMessages)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to fetch messages: %s", err.Error()), http.StatusInternalServerError)
			return
		}

		if records == nil {
			http.Error(w, fmt.Sprintf("%s: kafka client sent a nil response", unexpected), http.StatusInternalServerError)
			return
		}

		messages := make([]KafkaMessage, 0)
		for _, record := range records {
			messages = append(messages, getMessageFromRecord(record, config.Encoding, config.Proto))
		}

		jsonBytes, err := json.Marshal(messages)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to encode JSON: %s", err.Error()), http.StatusInternalServerError)
			return
		}

		w.Header().Set(contentType, applicationJSON)
		if _, err := w.Write(jsonBytes); err != nil {
			log.Printf("failed to write bytes to HTTP connection: %s", err.Error())
		}
	}
}

func getConfig(config c.Config) func(w http.ResponseWriter, _ *http.Request) {
	return func(w http.ResponseWriter, _ *http.Request) {
		jsonBytes, err := json.Marshal(config)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to encode JSON: %s", err.Error()), http.StatusInternalServerError)
			return
		}

		w.Header().Set(contentType, applicationJSON)
		if _, err := w.Write(jsonBytes); err != nil {
			log.Printf("failed to write bytes to HTTP connection: %s", err.Error())
		}
	}
}

func getMessageFromRecord(record *kgo.Record, deserializationFmt c.EncodingFormat, protoConfig *c.ProtoConfig) KafkaMessage {
	msgMetadata := utils.GetRecordMetadata(record)

	if len(record.Value) == 0 {
		return KafkaMessage{
			Key:       string(record.Key),
			Offset:    record.Offset,
			Partition: record.Partition,
			Metadata:  msgMetadata,
			Tombstone: true,
		}
	}

	var valueBytes []byte
	var err error
	switch deserializationFmt {
	case c.JSON:
		valueBytes, err = s.ParseJSONEncodedBytes(record.Value)
	case c.Protobuf:
		if protoConfig == nil {
			err = fmt.Errorf("%s: protobuf descriptor is nil when it shouldn't be", unexpected)
		} else {
			valueBytes, err = s.ParseProtobufEncodedBytes(record.Value, protoConfig.MsgDescriptor)
		}
	default:
		valueBytes = record.Value
	}

	value := string(valueBytes)

	return KafkaMessage{
		Key:       string(record.Key),
		Offset:    record.Offset,
		Partition: record.Partition,
		Metadata:  msgMetadata,
		Value:     &value,
		Err:       err,
	}
}
