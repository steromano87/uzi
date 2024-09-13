package db

import (
	"context"
	"github.com/glebarez/sqlite"
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

func NewAdapter(dsn string) *Adapter {
	adapter := new(Adapter)
	adapter.dsn = dsn

	return adapter
}

func (a *Adapter) Connect() error {
	DB, err := gorm.Open(sqlite.Open(a.dsn), &gorm.Config{PrepareStmt: true})
	if err != nil {
		return err
	}

	a.DB = DB
	return nil
}

func (a *Adapter) Close() error {
	rawDB, err := a.DB.DB()
	if err != nil {
		return err
	}

	return rawDB.Close()
}

func (a *Adapter) MigrateAll(ctx context.Context) error {
	return autoMigrate(ctx, a.DB)
}
