package forwarder

import "fmt"

type Behaviours struct {
	RunServer bool
	Host      string
	Port      uint
}

func (b Behaviours) Display() string {
	value := fmt.Sprintf(`Forward Behaviours:
  run server              %v`,
		b.RunServer,
	)

	if b.RunServer {
		value = fmt.Sprintf(`%s
  host                    %s
  port                    %d`,
			value,
			b.Host,
			b.Port,
		)
	}

	return value
}
