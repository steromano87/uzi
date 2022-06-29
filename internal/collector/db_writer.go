package collector

import (
	"context"
	"github.com/steromano87/harkonnen/v1/pkg/model"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type DbWriter struct {
	Type string
	DSN  string
	*gorm.DB
}

func (m *DbWriter) Connect(ctx context.Context) error {
	dialector, err := m.getDialector()
	if err != nil {
		return err
	}

	m.DB, err = gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		return err
	}

	err = m.migrateAll(ctx)
	return err
}

func (m *DbWriter) getDialector() (gorm.Dialector, error) {
	switch m.Type {
	case SQLite:
		return sqlite.Open(m.DSN), nil

	case MySQL:
		return mysql.Open(m.DSN), nil

	default:
		return gorm.Config{}, ErrUnknownType{unkType: m.Type}
	}
}

func (m *DbWriter) migrateAll(ctx context.Context) error {
	for _, element := range []any{&model.Log{}, &model.Sample{}} {
		err := m.WithContext(ctx).AutoMigrate(element)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *DbWriter) Save(ctx context.Context, item any) error {
	result := m.WithContext(ctx).Create(item)
	return result.Error
}
