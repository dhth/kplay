package types

import (
	"fmt"
)

type Config struct {
	Name           string         `json:"profile_name"`
	Authentication AuthType       `json:"-"`
	Encoding       EncodingFormat `json:"-"`
	Brokers        []string       `json:"brokers"`
	Topic          string         `json:"topic"`
	ConsumerGroup  string         `json:"consumer_group"`
	Proto          *ProtoConfig   `json:"-"`
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
		return fmt.Sprintf("protobuf (descriptor set: %s, descriptor name: %s)", c.Proto.DescriptorSetFile, c.Proto.DescriptorName)
	case Raw:
		return "raw"
	default:
		return "unknown"
	}
}
