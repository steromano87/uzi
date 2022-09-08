package db

import (
	"time"
)

type baseRecord struct {
	ID        uint      `gorm:"primaryKey;autoincrement"`
	Timestamp time.Time `gorm:"index;type:uint"`
}
