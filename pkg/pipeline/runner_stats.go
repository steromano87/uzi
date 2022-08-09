package pipeline

type RunnerStats struct {
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
