package workspace

import (
	"context"
	"github.com/steromano87/uzi/v1/pkg/db"
	"gorm.io/gorm"
	"os"
	"path"
)

const (
	DBDataFile = "data.db"
)

type Session struct {
	name         string
	parentFolder string
	dbAdapter    *db.Adapter
}

func NewSession(ctx context.Context, name string, parentFolder string) (*Session, error) {
	session := new(Session)
	session.name = name
	session.parentFolder = parentFolder
	if err := os.MkdirAll(path.Join(parentFolder, SessionsFolder, name), FolderPerms); err != nil {
		return nil, err
	}

	session.dbAdapter = db.NewAdapter(path.Join(parentFolder, SessionsFolder, name, DBDataFile))

	if err := session.dbAdapter.Connect(); err != nil {
		return nil, err
	}

	if err := session.dbAdapter.MigrateAll(ctx); err != nil {
		return nil, err
	}

	return session, nil
}

func LoadSession(name string, parentFolder string) *Session {
	run := new(Session)
	run.name = name
	run.parentFolder = parentFolder
	run.dbAdapter = db.NewAdapter(path.Join(parentFolder, SessionsFolder, name, DBDataFile))

	return run
}

func (r *Session) DB() *gorm.DB {
	return r.dbAdapter.DB
}
