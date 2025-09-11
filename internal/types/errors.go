package types

import (
	"errors"
)

var (
	ErrCouldntCreateDir          = errors.New("couldn't create directory")
	ErrCouldntWriteToFile        = errors.New("couldn't write to file")
	ErrCouldntShutDownGracefully = errors.New("couldn't shut down gracefully")
)
