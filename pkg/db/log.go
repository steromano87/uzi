package db

import (
	"gorm.io/datatypes"
)

func init() {
	registerEntityForAutoMigration(&Log{})
}

type Log struct {
	ID        uint           `gorm:"primaryKey;autoincrement"`
	Timestamp datatypes.Date `gorm:"index"`
	HostID    string         `gorm:"index"`
	Level     string         `gorm:"index"`
	Component string
	Message   string
	Data      datatypes.JSON
}
