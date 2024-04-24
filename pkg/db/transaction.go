package db

import (
	"gorm.io/datatypes"
)

func init() {
	registerEntityForAutoMigration(&Transaction{})
}

type Transaction struct {
	Name      string `gorm:"primaryKey"`
	Iteration uint   `gorm:"primaryKey"`
	Origin    string `gorm:"index"`
	Started   datatypes.Date
	Ended     datatypes.Date
	Passed    bool `gorm:"index"`
}
