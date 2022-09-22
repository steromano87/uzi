package db

import (
	"context"
	"errors"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const (
	SQLite string = "sqlite"
	MySQL         = "mysql"
)

type Adapter struct {
	dbType string
	dsn    string
	*gorm.DB
}

func NewAdapter(dbType string, dsn string) *Adapter {
	adapter := new(Adapter)
	adapter.dbType = dbType
	adapter.dsn = dsn

	return adapter
}

func (m *Adapter) Connect(ctx context.Context) error {
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

func (m *Adapter) getDialector() (gorm.Dialector, error) {
	switch m.dbType {
	case SQLite:
		return sqlite.Open(m.dsn), nil

	case MySQL:
		return mysql.Open(m.dsn), nil

	default:
		return gorm.Config{}, errors.New("unknown database type: " + m.dbType)
	}
}

func (m *Adapter) migrateAll(ctx context.Context) error {
	autoMigrator := automigrator{[]any{
		Log{},
		Sample{},
		Transaction{},
		Iteration{},
		HostMetric{},
	}}

	return autoMigrator.migrate(ctx, m.DB)
}
