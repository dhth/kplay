package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	t "github.com/dhth/kplay/internal/types"
	"github.com/tidwall/pretty"
)

func getMsgDetailsStylized(m t.Message, encoding t.EncodingFormat, width int) string {
	var msgValue string
	wrappedStyle := lipgloss.NewStyle().Width(width)
	if len(m.Value) == 0 {
		msgValue = msgDetailsTombstoneStyle.Render("tombstone")
	} else if m.DecodeErr != nil {
		var decodeErrFallback string
		if len(m.DecodeErrFallback) > 0 {
			decodeErrFallback = fmt.Sprintf("\n\n%s", m.DecodeErrFallback)
		}
		errorText := fmt.Sprintf("Decode Error: %s%s", m.DecodeErr.Error(), decodeErrFallback)
		msgValue = msgDetailsErrorStyle.Render(wrappedStyle.Render(errorText))
	} else {
		var rawValue string
		switch encoding {
		case t.JSON, t.Protobuf:
			rawValue = string(pretty.Color(m.Value, nil))
		case t.Raw:
			rawValue = string(m.Value)
		}
		msgValue = wrappedStyle.Render(rawValue)
	}

	return fmt.Sprintf(`%s

%s

%s

%s
`,
		msgDetailsHeadingStyle.Render("Metadata"),
		wrappedStyle.Render(m.Metadata),
		msgDetailsHeadingStyle.Render("Value"),
		msgValue,
	)
}
