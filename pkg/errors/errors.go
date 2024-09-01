package errors

import "errors"

var (
	MaxIterationsReached      = errors.New("max iterations reached")
	GracefulShutdownRequested = errors.New("graceful shutdown requested")
	ForcedShutdownRequested   = errors.New("forced shutdown requested")
)
