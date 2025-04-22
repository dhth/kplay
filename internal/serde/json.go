package serde

import (
	"encoding/json"
	"errors"
	"fmt"
)

var errCouldntUnmarshalJSONData = errors.New("couldn't unmarshal JSON encoded bytes")

func ParseJSONEncodedBytes(bytes []byte) ([]byte, error) {
	var data map[string]any
	err := json.Unmarshal(bytes, &data)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", errCouldntUnmarshalJSONData, err.Error())
	}

	indentedBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		// nolint:nilerr
		return bytes, nil
	}

	return indentedBytes, nil
}
