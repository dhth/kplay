package types

import (
	"fmt"
	"strings"
)

type TLSConfig struct {
	Enabled            bool
	InsecureSkipVerify bool
	RootCAFile         string // Path to custom root CA certificate file for verifying server certificates
	ClientCertFile     string // Path to client certificate file for mTLS authentication
	ClientKeyFile      string // Path to client private key file for mTLS authentication
}

type Config struct {
	Name           string         `json:"profile_name"`
	Authentication AuthType       `json:"-"`
	Encoding       EncodingFormat `json:"-"`
	Brokers        []string       `json:"brokers"`
	Topic          string         `json:"topic"`
	Proto          *ProtoConfig   `json:"-"`
	TLS            *TLSConfig     `json:"-"`
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

func (c Config) TLSDisplay() string {
	if c.TLS == nil || !c.TLS.Enabled {
		return "disabled"
	}
	if c.TLS.InsecureSkipVerify {
		return "enabled (insecure - skip verify)"
	}
	return "enabled"
}

func (c Config) Display() string {
	return fmt.Sprintf(`Profile:
  name                    %s
  topic                   %s
  authentication          %s
  encoding                %s
  tls                     %s
  brokers                 %s`,
		c.Name,
		c.Topic,
		c.AuthenticationDisplay(),
		c.EncodingDisplay(),
		c.TLSDisplay(),
		strings.Join(c.Brokers, "\n                          "))
}
