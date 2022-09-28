package injector

type RunnerStats struct {
	Scheduled              int
	Ready                  int
	Started                int
	Running                int
	Completed              int
	GracefullyShuttingDown int
	Stopped                int
	ForcefullyShuttingDown int
	ForcefullyStopped      int
	Error                  int
}
