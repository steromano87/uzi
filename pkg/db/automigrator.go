package db

import (
	"context"
	"gorm.io/gorm"
)

var autoMigrateRegisteredEntities []any

func registerEntityForAutoMigration(entity any) {
	autoMigrateRegisteredEntities = append(autoMigrateRegisteredEntities, entity)
}

func autoMigrate(ctx context.Context, db *gorm.DB) error {
	for _, element := range autoMigrateRegisteredEntities {
		err := db.WithContext(ctx).AutoMigrate(element)
		if err != nil {
			return err
		}
	}

	return nil
}
