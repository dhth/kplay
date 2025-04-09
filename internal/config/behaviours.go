package config

import "fmt"

type TUIBehaviours struct {
	PersistMessages bool
	SkipMessages    bool
	CommitMessages  bool
}

func (b TUIBehaviours) Display() string {
	return fmt.Sprintf(`
- persist messages        %v
- skip messages           %v
- commit messages         %v
`,
		b.PersistMessages,
		b.SkipMessages,
		b.CommitMessages,
	)
}

type WebBehaviours struct {
	SelectOnHover bool `json:"select_on_hover"`
}

func (b WebBehaviours) Display() string {
	return fmt.Sprintf(`
- select on hover         %v
`,
		b.SelectOnHover,
	)
}
