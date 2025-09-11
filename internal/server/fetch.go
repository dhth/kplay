package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	k "github.com/dhth/kplay/internal/kafka"
	t "github.com/dhth/kplay/internal/types"
	"github.com/twmb/franz-go/pkg/kgo"
)

const (
	contentType     = "Content-Type"
	applicationJSON = "application/json; charset=utf-8"
	unexpected      = "something unexpected happened (let @dhth know about this via https://github.com/dhth/kplay/issues)"
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

		commitStr := queryParams.Get("commit")
		var commitMessages bool
		if commitStr != "" {
			parsed, err := strconv.ParseBool(commitStr)
			if err != nil {
				http.Error(w, fmt.Sprintf("incorrect value provided for query param \"commit\": %s", err.Error()), http.StatusBadRequest)
				return
			}
			commitMessages = parsed
		}

		records, err := k.FetchAndCommitRecords(client, commitMessages, numMessages)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to fetch messages: %s", err.Error()), http.StatusInternalServerError)
			return
		}

		if records == nil {
			http.Error(w, fmt.Sprintf("%s: kafka client sent a nil response", unexpected), http.StatusInternalServerError)
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

func getBehaviours(behaviours t.WebBehaviours) func(w http.ResponseWriter, _ *http.Request) {
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
