package server

import "fmt"

type Behaviours struct {
	SelectOnHover bool `json:"select_on_hover"`
}

func (b Behaviours) Display() string {
	return fmt.Sprintf(`Web Behaviours:
  select on hover         %v`,
		b.SelectOnHover,
	)
}
