package tui

import (
	"fmt"

	"github.com/tidwall/pretty"
)

func getMsgDetails(details messageDetails) string {
	var msgValue string
	if details.tombstone {
		msgValue = "tombstone"
	} else if details.err != nil {
		msgValue = details.err.Error()
	} else {
		msgValue = string(details.value)
	}

	return fmt.Sprintf(`%s

%s

%s

%s`,
		"Metadata",
		details.metadata,
		"Value",
		msgValue,
	)
}

func getMsgDetailsStylized(details messageDetails) string {
	var msgValue string
	if details.tombstone {
		msgValue = msgDetailsTombstoneStyle.Render("tombstone")
	} else if details.err != nil {
		msgValue = msgDetailsErrorStyle.Render(details.err.Error())
	} else {
		msgValue = string(pretty.Color(details.value, nil))
	}

	return fmt.Sprintf(`%s

%s

%s

%s
`,
		msgDetailsHeadingStyle.Render("Metadata"),
		details.metadata,
		msgDetailsHeadingStyle.Render("Value"),
		msgValue,
	)
}
