package forwarder

import (
	"fmt"
)

type Behaviours struct {
	ConsumerGroup                  string
	FetchBatchSize                 uint16
	NumUploadWorkers               uint16
	ForwarderShutdownTimeoutMillis uint16
	PollFetchTimeoutMillis         uint16
	UploadTimeoutMillis            uint16
	PollSleepMillis                uint32
	UploadWorkerSleepMillis        uint32
	UploadReports                  bool
	ReportBatchSize                uint16
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
  fetch timeout (ms)      %d
  upload timeout (ms)     %d
  poll sleep (ms)         %d
  worker sleep (ms)       %d
  run server              %v
  upload reports          %v`,
		b.ConsumerGroup,
		b.FetchBatchSize,
		b.NumUploadWorkers,
		b.ForwarderShutdownTimeoutMillis,
		b.PollFetchTimeoutMillis,
		b.UploadTimeoutMillis,
		b.PollSleepMillis,
		b.UploadWorkerSleepMillis,
		b.RunServer,
		b.UploadReports,
	)

	if b.UploadReports {
		value = fmt.Sprintf(`%s
  report batch size       %d`,
			value,
			b.ReportBatchSize,
		)
	}

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
