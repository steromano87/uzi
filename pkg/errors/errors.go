package errors

import "errors"

var (
	MaxIterationsReached      = errors.New("max iterations reached")
	GracefulShutdownRequested = errors.New("graceful shutdown requested")
	ForcedShutdownRequested   = errors.New("forced shutdown requested")
	SessionAlreadyInProgress  = errors.New("cannot start a new session while another session is already in progress")
	NoSessionsInProgress      = errors.New("no sessions in progress")
)
