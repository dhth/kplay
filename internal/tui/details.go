package tui

import (
	"fmt"

	t "github.com/dhth/kplay/internal/types"
	"github.com/tidwall/pretty"
)

func getMsgDetails(m t.Message) string {
	var msgValue string
	if len(m.Value) == 0 {
		msgValue = "tombstone"
	} else if m.Err != nil {
		msgValue = m.Err.Error()
	} else {
		msgValue = string(m.Value)
	}

	return fmt.Sprintf(`%s

%s

%s

%s`,
		"Metadata",
		m.Metadata,
		"Value",
		msgValue,
	)
}

func getMsgDetailsStylized(m t.Message) string {
	var msgValue string
	if len(m.Value) == 0 {
		msgValue = msgDetailsTombstoneStyle.Render("tombstone")
	} else if m.Err != nil {
		msgValue = msgDetailsErrorStyle.Render(m.Err.Error())
	} else {
		msgValue = string(pretty.Color(m.Value, nil))
	}

	return fmt.Sprintf(`%s

%s

%s

%s
`,
		msgDetailsHeadingStyle.Render("Metadata"),
		m.Metadata,
		msgDetailsHeadingStyle.Render("Value"),
		msgValue,
	)
}
