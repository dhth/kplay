package ui

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dhth/kplay/ui/model"
	"github.com/twmb/franz-go/pkg/kgo"
)

func RenderUI(kCl *kgo.Client) {
	p := tea.NewProgram(model.InitialModel(kCl), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatalf("Something went wrong %s", err)
	}
}
