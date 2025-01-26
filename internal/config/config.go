package config

import (
	"fmt"

	"google.golang.org/protobuf/reflect/protoreflect"
)

const (
	awsMSKIAM = "aws_msk_iam"
)

type EncodingFormat uint

const (
	JSON EncodingFormat = iota
	Protobuf
)

func ValidateEncodingFmtValue(value string) (EncodingFormat, error) {
	switch value {
	case "json":
		return JSON, nil
	case "protobuf":
		return Protobuf, nil
	default:
		return JSON, fmt.Errorf("encoding format is missing/incorrect; possible values: [json, protobuf]")
	}
}

type AuthType uint

const (
	NoAuth AuthType = iota
	AWSMSKIAM
)

func ValidateAuthValue(value string) (AuthType, error) {
	switch value {
	case "none":
		return NoAuth, nil
	case awsMSKIAM:
		return AWSMSKIAM, nil
	default:
		return NoAuth, fmt.Errorf("auth value is missing/incorrect; possible values: [none, %s]", awsMSKIAM)
	}
}

type Config struct {
	Name               string
	Authentication     AuthType
	Encoding           EncodingFormat
	Brokers            []string
	Topic              string
	ConsumerGroup      string
	ProtoMsgDescriptor *protoreflect.MessageDescriptor
}

type Behaviours struct {
	PersistMessages bool
	SkipMessages    bool
}

func (c Config) AuthenticationValue() string {
	switch c.Authentication {
	case NoAuth:
		return "none"
	case AWSMSKIAM:
		return "aws_msk_iam"
	default:
		return "unknown"
	}
}

func (c Config) EncodingValue() string {
	switch c.Encoding {
	case JSON:
		return "json"
	case Protobuf:
		return "protobuf"
	default:
		return "unknown"
	}
}
