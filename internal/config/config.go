package config

import (
	"google.golang.org/protobuf/reflect/protoreflect"
)

type Config struct {
	Name               string
	Authentication     AuthType
	Encoding           EncodingFormat
	Brokers            []string
	Topic              string
	ConsumerGroup      string
	ProtoMsgDescriptor *protoreflect.MessageDescriptor
}

func (c Config) AuthenticationDisplay() string {
	switch c.Authentication {
	case NoAuth:
		return "none"
	case AWSMSKIAM:
		return "aws_msk_iam"
	default:
		return "unknown"
	}
}

func (c Config) EncodingDisplay() string {
	switch c.Encoding {
	case JSON:
		return "json"
	case Protobuf:
		return "protobuf"
	case Raw:
		return "raw"
	default:
		return "unknown"
	}
}

type Behaviours struct {
	PersistMessages bool
	SkipMessages    bool
}
