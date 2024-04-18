package telemetry

type LogSink interface {
	AddLog(log *Log)
}

type StepMetricsSink interface {
	AddSample(sample *Sample)
	AddTransaction(transaction *Transaction)
}

type HostMetricsSink interface {
	AddHostMetrics(hostMetrics *HostMetrics)
}
