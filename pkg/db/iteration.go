package db

import "gorm.io/datatypes"

type Iteration struct {
	Iteration uint   `gorm:"primaryKey"`
	Origin    string `gorm:"index"`
	Passed    bool   `gorm:"index"`
	Started   datatypes.Date
	Ended     datatypes.Date
}
