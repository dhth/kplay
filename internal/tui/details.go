package tui

import (
	"fmt"

	t "github.com/dhth/kplay/internal/types"
	"github.com/tidwall/pretty"
)

func getMsgDetailsStylized(m t.Message, encoding t.EncodingFormat) string {
	var msgValue string
	if len(m.Value) == 0 {
		msgValue = msgDetailsTombstoneStyle.Render("tombstone")
	} else if m.DecodeErr != nil {
		var decodeErrFallback string
		if len(m.DecodeErrFallback) > 0 {
			decodeErrFallback = fmt.Sprintf("\n\n%s", m.DecodeErrFallback)
		}
		msgValue = msgDetailsErrorStyle.Render(fmt.Sprintf("Decode Error: %s%s", m.DecodeErr.Error(), decodeErrFallback))
	} else {
		switch encoding {
		case t.JSON, t.Protobuf:
			msgValue = string(pretty.Color(m.Value, nil))
		case t.Raw:
			msgValue = string(m.Value)
		}
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
