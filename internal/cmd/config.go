package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	k "github.com/dhth/kplay/internal/kafka"
	t "github.com/dhth/kplay/internal/types"
	"github.com/dhth/kplay/internal/utils"
	yaml "github.com/goccy/go-yaml"
	"google.golang.org/protobuf/reflect/protoreflect"
)

var (
	errCouldntParseConfig                 = errors.New("couldn't parse config file")
	errProfileNotFound                    = errors.New("profile not found")
	errBrokersEmpty                       = errors.New("brokers cannot be empty")
	errTopicEmpty                         = errors.New("topic cannot be empty")
	errNoProfilesDefined                  = errors.New("no profiles defined")
	errProtoConfigMissing                 = errors.New("protobuf config missing")
	errCouldntReadDescriptorSetFile       = errors.New("couldn't read descriptor set file")
	ErrIssueWithProtobufFileDescriptorSet = errors.New("there's an issue with the file descriptor set")
	errDescriptorNameIsInvalid            = errors.New("descriptor name is invalid")
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
}

type protoConfig struct {
	DescriptorSetFile string `yaml:"descriptorSetFile"`
	DescriptorName    string `yaml:"descriptorName"`
}

func ParseProfileConfig(bytes []byte, profileName string, homeDir string) (t.Config, error) {
	var kConfig kplayConfig
	var config t.Config

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

		auth, err := t.ValidateAuthValue(pr.Authentication)
		if err != nil {
			return config, err
		}

		encodingFmt, err := t.ValidateEncodingFmtValue(pr.EncodingFormat)
		if err != nil {
			return config, err
		}

		if len(pr.Brokers) == 0 {
			return config, errBrokersEmpty
		}

		if strings.TrimSpace(pr.Topic) == "" {
			return config, errTopicEmpty
		}

		if encodingFmt == t.Protobuf {
			if pr.ProtoConfig == nil {
				return config, errProtoConfigMissing
			}

			if strings.TrimSpace(pr.ProtoConfig.DescriptorSetFile) == "" {
				return config, fmt.Errorf("protobuf descriptor set file is empty/missing")
			}

			pr.ProtoConfig.DescriptorSetFile = utils.ExpandTilde(os.ExpandEnv(pr.ProtoConfig.DescriptorSetFile), homeDir)

			if strings.TrimSpace(pr.ProtoConfig.DescriptorName) == "" {
				return config, fmt.Errorf("protobuf descriptor name is empty/missing")
			}

			descriptorBytes, err := os.ReadFile(pr.ProtoConfig.DescriptorSetFile)
			if err != nil {
				return config, fmt.Errorf("%w: %s", errCouldntReadDescriptorSetFile, err.Error())
			}

			descriptorName := protoreflect.FullName(pr.ProtoConfig.DescriptorName)
			if !descriptorName.IsValid() {
				return config, errDescriptorNameIsInvalid
			}

			msgDescriptor, err := k.GetDescriptorFromDescriptorSet(descriptorBytes, descriptorName)
			if err != nil {
				return config, fmt.Errorf("%w: %s", ErrIssueWithProtobufFileDescriptorSet, err.Error())
			}

			protoCfg := t.ProtoConfig{
				DescriptorSetFile: pr.ProtoConfig.DescriptorSetFile,
				DescriptorName:    pr.ProtoConfig.DescriptorName,
				MsgDescriptor:     msgDescriptor,
			}
			return t.Config{
				Name:           profileName,
				Authentication: auth,
				Encoding:       encodingFmt,
				Brokers:        pr.Brokers,
				Topic:          pr.Topic,
				Proto:          &protoCfg,
			}, nil
		}

		return t.Config{
			Name:           profileName,
			Authentication: auth,
			Encoding:       encodingFmt,
			Brokers:        pr.Brokers,
			Topic:          pr.Topic,
		}, nil
	}

	return config, fmt.Errorf("%w; available profiles: %v", errProfileNotFound, availableProfiles)
}

func ParseProfileConfigs(bytes []byte, profileNames []string, homeDir string) ([]t.Config, error) {
	var kConfig kplayConfig
	var configs []t.Config

	err := yaml.Unmarshal(bytes, &kConfig)
	if err != nil {
		return configs, fmt.Errorf("%w: %s", errCouldntParseConfig, err.Error())
	}

	if len(kConfig.Profiles) == 0 {
		return configs, errNoProfilesDefined
	}

	availableProfiles := make([]string, len(kConfig.Profiles))
	for i, pr := range kConfig.Profiles {
		availableProfiles[i] = pr.Name
	}

	for _, profileName := range profileNames {
		found := false
		for _, pr := range kConfig.Profiles {
			if pr.Name != profileName {
				continue
			}
			found = true

			auth, err := t.ValidateAuthValue(pr.Authentication)
			if err != nil {
				return configs, err
			}

			encodingFmt, err := t.ValidateEncodingFmtValue(pr.EncodingFormat)
			if err != nil {
				return configs, err
			}

			if len(pr.Brokers) == 0 {
				return configs, errBrokersEmpty
			}

			if strings.TrimSpace(pr.Topic) == "" {
				return configs, errTopicEmpty
			}

			if encodingFmt == t.Protobuf {
				if pr.ProtoConfig == nil {
					return configs, errProtoConfigMissing
				}

				if strings.TrimSpace(pr.ProtoConfig.DescriptorSetFile) == "" {
					return configs, fmt.Errorf("protobuf descriptor set file is empty/missing")
				}

				pr.ProtoConfig.DescriptorSetFile = utils.ExpandTilde(os.ExpandEnv(pr.ProtoConfig.DescriptorSetFile), homeDir)

				if strings.TrimSpace(pr.ProtoConfig.DescriptorName) == "" {
					return configs, fmt.Errorf("protobuf descriptor name is empty/missing")
				}

				descriptorBytes, err := os.ReadFile(pr.ProtoConfig.DescriptorSetFile)
				if err != nil {
					return configs, fmt.Errorf("%w: %s", errCouldntReadDescriptorSetFile, err.Error())
				}

				descriptorName := protoreflect.FullName(pr.ProtoConfig.DescriptorName)
				if !descriptorName.IsValid() {
					return configs, errDescriptorNameIsInvalid
				}

				msgDescriptor, err := k.GetDescriptorFromDescriptorSet(descriptorBytes, descriptorName)
				if err != nil {
					return configs, fmt.Errorf("%w: %s", ErrIssueWithProtobufFileDescriptorSet, err.Error())
				}

				protoCfg := t.ProtoConfig{
					DescriptorSetFile: pr.ProtoConfig.DescriptorSetFile,
					DescriptorName:    pr.ProtoConfig.DescriptorName,
					MsgDescriptor:     msgDescriptor,
				}
				configs = append(configs, t.Config{
					Name:           profileName,
					Authentication: auth,
					Encoding:       encodingFmt,
					Brokers:        pr.Brokers,
					Topic:          pr.Topic,
					Proto:          &protoCfg,
				})
			} else {
				configs = append(configs, t.Config{
					Name:           profileName,
					Authentication: auth,
					Encoding:       encodingFmt,
					Brokers:        pr.Brokers,
					Topic:          pr.Topic,
				})
			}
			break
		}

		if !found {
			return configs, fmt.Errorf("%w; available profiles: %v", errProfileNotFound, availableProfiles)
		}
	}

	return configs, nil
}
