package pipeline

const (
	Running   = "RUNNING"
	Completed = "COMPLETED"

	GracefullyShuttingDown = "GRACEFULLY_SHUTTING_DOWN"
	Stopped                = "STOPPED"

	ForcefullyStopping = "FORCEFULLY_STOPPING"
	ForcefullyStopped  = "FORCEFULLY_STOPPED"

	Error = "ERROR"
)
