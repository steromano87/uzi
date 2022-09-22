package db

import (
	"gorm.io/datatypes"
)

type Log struct {
	ID        uint           `gorm:"primaryKey;autoincrement"`
	Timestamp datatypes.Date `gorm:"index"`
	Origin    string         `gorm:"index"`
	Level     string         `gorm:"index"`
	Component string
	Message   string
	Data      datatypes.JSON
}
