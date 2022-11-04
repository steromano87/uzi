package telemetry

import "sync"

type TransactionCollector struct {
	transactions []*Transaction
	mu           sync.Mutex
}

func NewTransactionCollector() *TransactionCollector {
	collector := new(TransactionCollector)
	collector.transactions = make([]*Transaction, 0)

	return collector
}

func (c *TransactionCollector) AddTransaction(transaction *Transaction) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.transactions = append(c.transactions, transaction)
}

func (c *TransactionCollector) GetTransactions() []*Transaction {
	c.mu.Lock()
	defer c.mu.Unlock()

	var output []*Transaction
	copy(output, c.transactions)
	c.transactions = make([]*Transaction, 0)

	return output
}
