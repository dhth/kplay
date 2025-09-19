package cmd

import (
	_ "embed"
	"errors"
	"fmt"
)

//go:embed assets/sample-config.yml
var sampleConfig []byte

var (
	errCouldntCreateKafkaClient = errors.New("couldn't create kafka client")
	errCouldntPingBrokers       = errors.New("couldn't ping brokers")
	errCouldntGetUserHomeDir    = errors.New("couldn't get your home directory")
	errCouldntGetUserConfigDir  = errors.New("couldn't get your config directory")
	ErrCouldntReadConfigFile    = errors.New("couldn't read config file")
	ErrConfigInvalid            = errors.New("config is invalid")
	errInvalidTimestampProvided = errors.New(`invalid value provided for "from timestamp"`)
	errInvalidOffsetProvided    = errors.New(`invalid value provided for "from offset"`)
	errInvalidRegexProvided     = errors.New("invalid regex provided")
)

func GetErrorFollowUp(err error) (string, bool) {
	if errors.Is(err, ErrIssueWithProtobufFileDescriptorSet) {
		return `
Hint: A protobuf file descriptor set can be created using the "Protocol Buffer Compiler" (https://grpc.io/docs/protoc-installation) as follows:
$ protoc path/to/proto/file.proto --descriptor_set_out=path/to/descriptor_set.pb --include_imports
`, true
	}

	if errors.Is(err, ErrConfigInvalid) || errors.Is(err, ErrCouldntReadConfigFile) {
		return fmt.Sprintf(`
kplay's config looks like this:
---
%s---
`, sampleConfig), true
	}

	if errors.Is(err, errInvalidOffsetProvided) {
		return `
Hint: --from-offset can be either of the following:
- an integer value, which will apply to all partitions (eg. --from-offset=1000)
- a string in the format <PARTITION>:<OFFSET>,... where an offset is specified for each partition (eg. --from-offset='0:1000,1:1500')
`, true
	}

	return "", false
}
