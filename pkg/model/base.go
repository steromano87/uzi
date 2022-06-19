package model

import (
	"time"
)

type Base struct {
	ID        uint      `gorm:"primaryKey;autoincrement"`
	Timestamp time.Time `gorm:"index;type:uint"`
}
