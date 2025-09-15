package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	k "github.com/dhth/kplay/internal/kafka"
	t "github.com/dhth/kplay/internal/types"
	"github.com/twmb/franz-go/pkg/kgo"
)

const (
	contentType     = "Content-Type"
	applicationJSON = "application/json; charset=utf-8"
)

func getMessages(client *kgo.Client, config t.Config) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		queryParams := r.URL.Query()
		numMessagesStr := queryParams.Get("num")

		var numMessages uint = 1
		if numMessagesStr != "" {
			num, err := strconv.Atoi(numMessagesStr)
			if err != nil || num < 1 {
				http.Error(w, fmt.Sprintf("incorrect value provided for query param \"num\": %s", err.Error()), http.StatusBadRequest)
				return
			}
			numMessages = uint(num)
		}
		if numMessages > 10 {
			numMessages = 10
		}

		fetchCtx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
		defer cancel()

		records, err := k.FetchRecords(fetchCtx, client, numMessages)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to fetch messages: %s", err.Error()), http.StatusInternalServerError)
			return
		}

		messages := make([]t.SerializableMessage, 0)
		for _, record := range records {
			messages = append(messages, t.GetMessageFromRecord(record, config, true).ToSerializable())
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

func getConfig(config t.Config) func(w http.ResponseWriter, _ *http.Request) {
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

func getBehaviours(behaviours Behaviours) func(w http.ResponseWriter, _ *http.Request) {
	return func(w http.ResponseWriter, _ *http.Request) {
		jsonBytes, err := json.Marshal(behaviours)
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
