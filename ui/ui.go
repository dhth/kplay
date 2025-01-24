package ui

import (
	"errors"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dhth/kplay/ui/model"
	"github.com/twmb/franz-go/pkg/kgo"
)

var errCouldntSetupDebugLogging = errors.New("couldn't set up debug logging")

func RenderUI(kCl *kgo.Client, kconfig model.Config) error {
	if len(os.Getenv("DEBUG")) > 0 {
		f, err := tea.LogToFile("debug.log", "debug")
		if err != nil {
			return fmt.Errorf("%w: %w", errCouldntSetupDebugLogging, err)
		}
		defer f.Close()
	}

	p := tea.NewProgram(model.InitialModel(kCl, kconfig), tea.WithAltScreen())
	_, err := p.Run()

	return err
}
