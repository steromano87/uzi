package telemetry

import "io"

type LoadMetricsStorer interface {
	StoreSample(sample *Sample) error
	StoreTransaction(transaction *Transaction) error
	StoreIterationCounters(counters *IterationCounters) error
}

type HostMetricsStorer interface {
	StoreHostMetrics(agentMetrics *HostMetrics) error
}

type LogStorer interface {
	io.Writer
	StoreLogEntry(entry *LogEntry) error
}
