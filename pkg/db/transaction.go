package db

import (
	"gorm.io/datatypes"
	"time"
)

func init() {
	registerEntityForAutoMigration(&Transaction{})
}

type Transaction struct {
	Id              uint   `gorm:"primaryKey;autoincrement"`
	Name            string `gorm:"index:idx_transaction_name"`
	AgentId         string `gorm:"index:idx_transaction_agent_user"`
	SyntheticUserId string `gorm:"index:idx_transaction_agent_user"`
	Start           datatypes.Date
	End             datatypes.Date
	Duration        time.Duration
	Successful      bool `gorm:"index"`
}
