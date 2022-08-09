package pipeline

const (
	Ready     = "READY"
	Running   = "RUNNING"
	Completed = "COMPLETED"

	GracefullyShuttingDown = "GRACEFULLY_SHUTTING_DOWN"
	Stopped                = "STOPPED"

	ForcefullyShuttingDown = "FORCEFULLY_STOPPING"
	ForcefullyStopped      = "FORCEFULLY_STOPPED"

	Error = "ERROR"
)
