package serde

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
)

var errCouldntUnmarshalToJSON = errors.New("couldn't unmarshal bytes to JSON")

func PrettifyJSON(data []byte) ([]byte, error) {
	var msg json.RawMessage
	err := json.Unmarshal(data, &msg)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", errCouldntUnmarshalToJSON, err.Error())
	}

	var out bytes.Buffer

	err = json.Indent(&out, data, "", "  ")
	if err != nil {
		// nolint:nilerr
		return data, nil
	}

	return out.Bytes(), nil
}
