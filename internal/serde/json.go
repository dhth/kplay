package serde

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/tidwall/pretty"
)

var errCouldntUnmarshalJSONData = errors.New("couldn't unmarshal JSON encoded bytes")

func ParseJSONEncodedBytes(bytes []byte) ([]byte, error) {
	var data map[string]interface{}
	err := json.Unmarshal(bytes, &data)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", errCouldntUnmarshalJSONData, err.Error())
	}

	return pretty.Pretty(bytes), nil
}
