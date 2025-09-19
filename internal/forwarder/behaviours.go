package forwarder

import (
	"fmt"
)

type Behaviours struct {
	ConsumerGroup                  string
	FetchBatchSize                 uint16
	NumUploadWorkers               uint16
	ForwarderShutdownTimeoutMillis uint16
	ServerShutdownTimeoutMillis    uint16
	PollSleepMillis                uint16
	UploadWorkerSleepMillis        uint16
	PollFetchTimeoutMillis         uint16
	UploadTimeoutMillis            uint16
	RunServer                      bool
	ServerHost                     string
	ServerPort                     uint16
}

func (b Behaviours) Display() string {
	value := fmt.Sprintf(`Forward Behaviours:
  consumer group          %s
  fetch batch size        %d
  upload workers          %d
  shutdown timeout (ms)   %d
  server shutdown (ms)    %d
  poll sleep (ms)         %d
  worker sleep (ms)       %d
  fetch timeout (ms)      %d
  upload timeout (ms)     %d
  run server              %v`,
		b.ConsumerGroup,
		b.FetchBatchSize,
		b.NumUploadWorkers,
		b.ForwarderShutdownTimeoutMillis,
		b.ServerShutdownTimeoutMillis,
		b.PollSleepMillis,
		b.UploadWorkerSleepMillis,
		b.PollFetchTimeoutMillis,
		b.UploadTimeoutMillis,
		b.RunServer,
	)

	if b.RunServer {
		value = fmt.Sprintf(`%s
  server host             %s
  server port             %d`,
			value,
			b.ServerHost,
			b.ServerPort,
		)
	}

	return value
}
