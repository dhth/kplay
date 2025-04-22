package tui

import (
	"errors"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	t "github.com/dhth/kplay/internal/types"
	"github.com/twmb/franz-go/pkg/kgo"
)

var errCouldntSetupDebugLogging = errors.New("couldn't set up debug logging")

func Render(kCl *kgo.Client, config t.Config, behaviours t.TUIBehaviours) error {
	if len(os.Getenv("DEBUG")) > 0 {
		f, err := tea.LogToFile("debug.log", "debug")
		if err != nil {
			return fmt.Errorf("%w: %w", errCouldntSetupDebugLogging, err)
		}
		defer f.Close()
	}

	p := tea.NewProgram(InitialModel(kCl, config, behaviours), tea.WithAltScreen())
	_, err := p.Run()

	return err
}
