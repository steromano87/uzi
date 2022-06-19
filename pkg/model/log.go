package model

type Log struct {
	Base
	Origin    string `gorm:"index:idx_origin"`
	Level     string `gorm:"index:idx_level"`
	Component string
	Message   string
	Data      map[string]any `gorm:"serializer:json"`
}
