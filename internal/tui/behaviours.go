package tui

import "fmt"

type Behaviours struct {
	PersistMessages bool
	SkipMessages    bool
}

func (b Behaviours) Display() string {
	return fmt.Sprintf(`TUI Behaviours:
  persist messages        %v
  skip messages           %v`,
		b.PersistMessages,
		b.SkipMessages,
	)
}
