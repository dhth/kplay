package main

import (
	"fmt"
	"os"

	"github.com/dhth/kplay/internal/cmd"
)

var version = "dev"

func main() {
	err := cmd.Execute(version)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		followUp, ok := cmd.GetErrorFollowUp(err)
		if ok {
			fmt.Fprintf(os.Stderr, "%s", followUp)
		}
		os.Exit(1)
	}
}
