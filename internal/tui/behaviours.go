package tui

import "fmt"

type Behaviours struct {
	CommitMessages  bool
	PersistMessages bool
	SkipMessages    bool
}

func (b Behaviours) Display() string {
	return fmt.Sprintf(`TUI Behaviours:
  commit messages         %v
  persist messages        %v
  skip messages           %v`,
		b.CommitMessages,
		b.PersistMessages,
		b.SkipMessages,
	)
}
