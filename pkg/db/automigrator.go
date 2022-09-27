package db

import (
	"context"
	"gorm.io/gorm"
)

type autoMigrator struct {
	registeredEntities []any
}

func (a *autoMigrator) register(entity any) {
	a.registeredEntities = append(a.registeredEntities, entity)
}

func (a *autoMigrator) migrate(ctx context.Context, db *gorm.DB) error {
	for _, element := range a.registeredEntities {
		err := db.WithContext(ctx).AutoMigrate(element)
		if err != nil {
			return err
		}
	}

	return nil
}
