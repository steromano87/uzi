package db

import (
	"github.com/steromano87/harkonnen/v1/pkg/telemetry"
	"gorm.io/datatypes"
	"time"
)

func init() {
	registerEntityForAutoMigration(&Transaction{})
}

type Transaction struct {
	Id              uint           `gorm:"primaryKey;autoincrement"`
	AgentId         string         `gorm:"index:idx_transaction_agent_user"`
	SyntheticUserId string         `gorm:"index:idx_transaction_agent_user"`
	Name            string         `gorm:"index:idx_transaction_name"`
	Start           datatypes.Date `gorm:"index:idx_transaction_start"`
	End             datatypes.Date
	Duration        time.Duration
	Successful      bool `gorm:"index"`
}

func NewTransactionFromGrpc(transaction *telemetry.Transaction) Transaction {
	return Transaction{
		SyntheticUserId: transaction.GetSyntheticUserId(),
		Name:            transaction.GetName(),
		Start:           datatypes.Date(transaction.GetStart().AsTime()),
		End:             datatypes.Date(transaction.GetEnd().AsTime()),
		Duration:        transaction.GetDuration().AsDuration(),
		Successful:      transaction.GetSuccessful(),
	}
}
