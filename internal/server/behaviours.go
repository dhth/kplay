package server

import "fmt"

type Behaviours struct {
	CommitMessages bool `json:"commit_messages"`
	SelectOnHover  bool `json:"select_on_hover"`
}

func (b Behaviours) Display() string {
	return fmt.Sprintf(`Web Behaviours:
  commit messages         %v
  select on hover         %v`,
		b.CommitMessages,
		b.SelectOnHover,
	)
}
