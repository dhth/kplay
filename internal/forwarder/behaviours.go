package forwarder

import "fmt"

type Behaviours struct {
	Host string
	Port uint
}

func (b Behaviours) Display() string {
	value := fmt.Sprintf(`Forward Behaviours:
  host                    %s
  port                    %d`,
		b.Host,
		b.Port,
	)

	return value
}
