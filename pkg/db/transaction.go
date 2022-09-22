package db

import (
	"gorm.io/datatypes"
)

type Transaction struct {
	ID        uint `gorm:"primaryKey;autoincrement"`
	Iteration uint
	Name      string         `gorm:"index"`
	Started   datatypes.Date `gorm:"type:uint"`
	Ended     datatypes.Date `gorm:"type:uint"`
	Passed    bool
}
