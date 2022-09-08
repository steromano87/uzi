package db

import (
	"time"
)

type Transaction struct {
	baseRecord
	Name    string    `gorm:"index"`
	Started time.Time `gorm:"type:uint"`
	Ended   time.Time `gorm:"type:uint"`
	Passed  bool
}
