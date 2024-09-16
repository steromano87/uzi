package telemetry

type LoadMetricsStorer interface {
	StoreSample(sample *Sample) error
	StoreTransaction(transaction *Transaction) error
	StoreIterationCounters(counters *IterationCounters) error
}

type HostMetricsStorer interface {
	StoreHostMetrics(agentMetrics *HostMetrics) error
}
