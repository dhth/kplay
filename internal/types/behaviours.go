package types

import "fmt"

type TUIBehaviours struct {
	CommitMessages  bool
	PersistMessages bool
	SkipMessages    bool
}

func (b TUIBehaviours) Display() string {
	return fmt.Sprintf(`
- commit messages         %v
- persist messages        %v
- skip messages           %v
`,
		b.CommitMessages,
		b.PersistMessages,
		b.SkipMessages,
	)
}

type WebBehaviours struct {
	CommitMessages bool `json:"commit_messages"`
	SelectOnHover  bool `json:"select_on_hover"`
}

func (b WebBehaviours) Display() string {
	return fmt.Sprintf(`
- commit messages         %v
- select on hover         %v
`,
		b.CommitMessages,
		b.SelectOnHover,
	)
}
