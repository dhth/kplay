package types

import (
	"errors"
)

var (
	ErrCouldntCreateDir              = errors.New("couldn't create directory")
	ErrCouldntWriteToFile            = errors.New("couldn't write to file")
	ErrCouldntShutDownGracefully     = errors.New("couldn't shut down gracefully")
	ErrCouldntLoadAwsConfig          = errors.New("couldn't load AWS config")
	ErrCouldntRetrieveAWSCredentials = errors.New("couldn't retrieve AWS credentials")
	ErrForcefulServerShutdownFailed  = errors.New("forceful shutdown of http server failed")
	ErrCouldntStartHTTPServer        = errors.New("couldn't start http server")
)
