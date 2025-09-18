package forwarder

import "fmt"

type Behaviours struct {
	Host       string
	Port       uint
	BucketName string
}

func (b Behaviours) Display() string {
	value := fmt.Sprintf(`Forward Behaviours:
  host                    %s
  port                    %d,
  bucket name             %s`,
		b.Host,
		b.Port,
		b.BucketName,
	)

	return value
}
