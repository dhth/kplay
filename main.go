package main

import (
	"errors"
	"fmt"
	"os"

	_ "embed"

	"github.com/dhth/kplay/cmd"
)

//go:embed cmd/assets/sample-config.yml
var sampleConfig []byte

func main() {
	err := cmd.Execute()
	if err != nil {
		if errors.Is(err, cmd.ErrConfigInvalid) || errors.Is(err, cmd.ErrCouldntReadConfigFile) {
			fmt.Fprintf(os.Stderr, `
kplay's config looks like this:
---
%s---

Run kplay -h for more details.
`, sampleConfig)
		}
		os.Exit(1)
	}
}
