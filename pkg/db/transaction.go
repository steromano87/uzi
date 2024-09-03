package db

import (
	"gorm.io/datatypes"
	"time"
)

func init() {
	registerEntityForAutoMigration(&Transaction{})
}

type Transaction struct {
	Name            string `gorm:"primaryKey"`
	Iteration       uint   `gorm:"primaryKey"`
	AgentId         string `gorm:"index:agent_user"`
	SyntheticUserId string `gorm:"index:agent_user"`
	Start           datatypes.Date
	End             datatypes.Date
	Duration        time.Duration
	Successful      bool `gorm:"index"`
}
