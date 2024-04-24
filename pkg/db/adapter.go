package db

import (
	"context"
	"errors"
	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

const (
	SQLite string = "sqlite"
	MySQL         = "mysql"

	SQLiteDSNForInMemoryDB = "file::memory:?cache=shared"
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

func (a *Adapter) Connect() error {
	dialector, err := a.getDialector()
	if err != nil {
		return err
	}

	a.DB, err = gorm.Open(dialector, &gorm.Config{PrepareStmt: true})
	if err != nil {
		return err
	}

	return nil
}

func (a *Adapter) Close() error {
	rawDB, err := a.DB.DB()
	if err != nil {
		return err
	}

	return rawDB.Close()
}

func (a *Adapter) getDialector() (gorm.Dialector, error) {
	switch a.dbType {
	case SQLite:
		return sqlite.Open(a.dsn), nil

	case MySQL:
		return mysql.Open(a.dsn), nil

	default:
		return gorm.Config{}, errors.New("unknown database type: " + a.dbType)
	}
}

func (a *Adapter) MigrateAll(ctx context.Context) error {
	return autoMigrate(ctx, a.DB)
}
