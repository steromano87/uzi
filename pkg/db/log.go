package db

import (
	"gorm.io/datatypes"
)

type Log struct {
	ID        uint           `gorm:"primaryKey;autoincrement"`
	Timestamp datatypes.Date `gorm:"index;idx_timestamp"`
	Origin    string         `gorm:"index:idx_origin"`
	Level     string         `gorm:"index:idx_level"`
	Component string
	Message   string
	Data      datatypes.JSON
}
