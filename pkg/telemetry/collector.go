package telemetry

type Collector interface {
	AddSample(sample *Sample)
	AddTransaction(transaction *Transaction)
}
