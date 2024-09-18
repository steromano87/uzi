package db

import (
	"context"
	"fmt"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"net/url"
)

const (
	SQLite string = "sqlite"
	MySQL         = "mysql"

	SQLiteDSNForInMemoryDB = "file::memory:"
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

func (a *Adapter) Connect(dsnOptions ...string) error {
	parsedOpts := url.Values{}
	for index := 0; index < len(dsnOptions); index += 2 {
		key := dsnOptions[index]
		val := dsnOptions[index+1]
		parsedOpts.Add(key, val)
	}

	completeDsn := a.dsn
	if len(parsedOpts) > 0 {
		completeDsn = fmt.Sprintf("%s?%s", a.dsn, parsedOpts.Encode())
	}

	DB, err := gorm.Open(sqlite.Open(completeDsn), &gorm.Config{
		PrepareStmt:            true,
		SkipDefaultTransaction: true,
	})
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
