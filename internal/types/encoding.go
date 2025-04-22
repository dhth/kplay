package types

import (
	"fmt"

	"google.golang.org/protobuf/reflect/protoreflect"
)

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

type ProtoConfig struct {
	DescriptorSetFile string
	DescriptorName    string
	MsgDescriptor     protoreflect.MessageDescriptor
}
