package pipeline

import "errors"

var (
	ErrMaxIterationsReached      = errors.New("max iterations reached")
	ErrGracefulShutdownRequested = errors.New("graceful shutdown requested")
	ErrForcedShutdownRequested   = errors.New("forced shutdown requested")
)
