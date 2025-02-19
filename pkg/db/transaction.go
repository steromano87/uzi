package db

import (
	"github.com/steromano87/uzi/v1/pkg/telemetry"
	"time"
)

func init() {
	registerEntityForAutoMigration(&Transaction{})
}

type Transaction struct {
	Id              uint      `gorm:"primaryKey;autoincrement"`
	AgentId         string    `gorm:"index:idx_transaction_agent_user"`
	SyntheticUserId string    `gorm:"index:idx_transaction_agent_user"`
	Name            string    `gorm:"index:idx_transaction_name"`
	Start           time.Time `gorm:"index:idx_transaction_start"`
	End             time.Time
	Duration        time.Duration
	Successful      bool `gorm:"index:idx_transaction_successful"`
	Global          bool `gorm:"index:idx_transaction_global"`
}

func NewTransactionFromGrpc(transaction *telemetry.Transaction) Transaction {
	return Transaction{
		SyntheticUserId: transaction.GetSyntheticUserId(),
		Name:            transaction.GetName(),
		Start:           transaction.GetStart().AsTime(),
		End:             transaction.GetEnd().AsTime(),
		Duration:        transaction.GetDuration().AsDuration(),
		Successful:      transaction.GetSuccessful(),
		Global:          transaction.GetGlobal(),
	}
}
