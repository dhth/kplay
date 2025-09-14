package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

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

		numMessages := 1
		if numMessagesStr != "" {
			num, err := strconv.Atoi(numMessagesStr)
			if err != nil || num < 1 {
				http.Error(w, fmt.Sprintf("incorrect value provided for query param \"num\": %s", err.Error()), http.StatusBadRequest)
				return
			}
			numMessages = num
		}
		if numMessages > 10 {
			numMessages = 10
		}

		ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
		defer cancel()

		records := client.PollRecords(ctx, numMessages).Records()

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
