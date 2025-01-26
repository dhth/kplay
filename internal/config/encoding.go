package config

import "fmt"

type EncodingFormat uint

const (
	JSON EncodingFormat = iota
	Protobuf
	Raw
)

func ValidateEncodingFmtValue(value string) (EncodingFormat, error) {
	switch value {
	case "json":
		return JSON, nil
	case "protobuf":
		return Protobuf, nil
	case "raw":
		return Raw, nil
	default:
		return JSON, fmt.Errorf("encoding format is missing/incorrect; possible values: [json, protobuf, raw]")
	}
}
