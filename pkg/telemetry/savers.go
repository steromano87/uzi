package telemetry

type LoadMetricsSaver interface {
	SaveSample(sample *Sample)
	SaveTransaction(transaction *Transaction)
}

type HostMetricsSaver interface {
	SaveHostMetricsSample(hostMetrics *HostMetricsSample)
}
