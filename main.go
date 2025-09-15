package main

import (
	"fmt"
	"os"

	"github.com/dhth/kplay/internal/cmd"
)

func main() {
	err := cmd.Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s", err.Error())
		followUp, ok := cmd.HandleErrors(err)
		if ok {
			fmt.Fprintf(os.Stderr, `
%s`, followUp)
		}
		os.Exit(1)
	}
}
