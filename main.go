package main

import (
	"errors"
	"fmt"
	"os"

	_ "embed"

	"github.com/dhth/kplay/internal/cmd"
)

//go:embed internal/cmd/assets/sample-config.yml
var sampleConfig []byte

func main() {
	err := cmd.Execute()
	if err != nil {
		if errors.Is(err, cmd.ErrConfigInvalid) || errors.Is(err, cmd.ErrCouldntReadConfigFile) {
			if errors.Is(err, cmd.ErrIssueWithProtobufFileDescriptorSet) {
				fmt.Fprint(os.Stderr, `
Hint: A protobuf file descriptor set can be created using the "Protocol Buffer Compiler" (https://grpc.io/docs/protoc-installation) as follows:
$ protoc path/to/proto/file.proto --descriptor_set_out=path/to/descriptor_set.pb --include_imports 
`)
			} else {
				fmt.Fprintf(os.Stderr, `
kplay's config looks like this:
---
%s---
`, sampleConfig)
			}
		}
		os.Exit(1)
	}
}
