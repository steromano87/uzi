package db

import "gorm.io/datatypes"

type Iteration struct {
	Iteration uint   `gorm:"primaryKey"`
	Origin    string `gorm:"index:idx_origin"`
	Passed    bool   `gorm:"index:idx_passed"`
	Started   datatypes.Date
	Ended     datatypes.Date
}
