package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	c "github.com/dhth/kplay/internal/config"
	k "github.com/dhth/kplay/internal/kafka"
	yaml "github.com/goccy/go-yaml"
)

var (
	errCouldntParseConfig           = errors.New("couldn't parse config file")
	errProfileNotFound              = errors.New("profile not found")
	errBrokersEmpty                 = errors.New("brokers cannot be empty")
	errTopicEmpty                   = errors.New("topic cannot be empty")
	errConsumerGroupEmpty           = errors.New("consumer group cannot be empty")
	errNoProfilesDefined            = errors.New("no profiles defined")
	errProtoConfigMissing           = errors.New("protobuf config missing")
	errCouldntReadDescriptorSetFile = errors.New("couldn't read descriptor set file")
)

type kplayConfig struct {
	Profiles []profile
}

type profile struct {
	Name           string
	Authentication string
	EncodingFormat string       `yaml:"encodingFormat"`
	ProtoConfig    *protoConfig `yaml:"protoConfig"`
	Brokers        []string
	Topic          string
	ConsumerGroup  string `yaml:"consumerGroup"`
}

type protoConfig struct {
	DescriptorSetFile string `yaml:"descriptorSetFile"`
	DescriptorName    string `yaml:"descriptorName"`
}

func GetProfileConfig(bytes []byte, profileName string) (c.Config, error) {
	var kConfig kplayConfig
	var config c.Config

	err := yaml.Unmarshal(bytes, &kConfig)
	if err != nil {
		return config, fmt.Errorf("%w: %s", errCouldntParseConfig, err.Error())
	}

	if len(kConfig.Profiles) == 0 {
		return config, errNoProfilesDefined
	}

	availableProfiles := make([]string, len(kConfig.Profiles))
	for i, pr := range kConfig.Profiles {
		availableProfiles[i] = pr.Name
		if pr.Name != profileName {
			continue
		}

		auth, err := c.ValidateAuthValue(pr.Authentication)
		if err != nil {
			return config, err
		}

		encodingFmt, err := c.ValidateEncodingFmtValue(pr.EncodingFormat)
		if err != nil {
			return config, err
		}

		if len(pr.Brokers) == 0 {
			return config, errBrokersEmpty
		}

		if strings.TrimSpace(pr.Topic) == "" {
			return config, errTopicEmpty
		}

		if strings.TrimSpace(pr.ConsumerGroup) == "" {
			return config, errConsumerGroupEmpty
		}

		if encodingFmt == c.Protobuf {
			if pr.ProtoConfig == nil {
				return config, errProtoConfigMissing
			}

			if strings.TrimSpace(pr.ProtoConfig.DescriptorSetFile) == "" {
				return config, fmt.Errorf("protobuf descriptor set file is empty/missing")
			}

			if strings.TrimSpace(pr.ProtoConfig.DescriptorName) == "" {
				return config, fmt.Errorf("protobuf descriptor name is empty/missing")
			}

			descriptorBytes, err := os.ReadFile(pr.ProtoConfig.DescriptorSetFile)
			if err != nil {
				return config, fmt.Errorf("%w: %s", errCouldntReadDescriptorSetFile, err.Error())
			}

			msgDescriptor, err := k.GetDescriptorFromDescriptorSet(descriptorBytes, pr.ProtoConfig.DescriptorName)
			if err != nil {
				return config, err
			}

			return c.Config{
				Name:               profileName,
				Authentication:     auth,
				Encoding:           encodingFmt,
				Brokers:            pr.Brokers,
				Topic:              pr.Topic,
				ConsumerGroup:      pr.ConsumerGroup,
				ProtoMsgDescriptor: &msgDescriptor,
			}, nil
		}

		return c.Config{
			Name:           profileName,
			Authentication: auth,
			Encoding:       encodingFmt,
			Brokers:        pr.Brokers,
			Topic:          pr.Topic,
			ConsumerGroup:  pr.ConsumerGroup,
		}, nil
	}

	return config, fmt.Errorf("%w; available profiles: %v", errProfileNotFound, availableProfiles)
}
