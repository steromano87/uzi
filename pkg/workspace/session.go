package workspace

import (
	"context"
	"github.com/steromano87/harkonnen/v1/pkg/db"
	"os"
	"path"
)

const (
	DBDataFile = "data.db"
)

type Session struct {
	name         string
	parentFolder string
	persistor    *db.Persistor
}

func NewSession(ctx context.Context, name string, parentFolder string) (*Session, error) {
	session := new(Session)
	session.name = name
	session.parentFolder = parentFolder
	if err := os.MkdirAll(path.Join(parentFolder, SessionsFolder, name), FolderPerms); err != nil {
		return nil, err
	}

	persistor, err := db.NewPersistor(path.Join(parentFolder, SessionsFolder, name, DBDataFile))
	if err != nil {
		return nil, err
	}
	session.persistor = persistor

	if err := session.persistor.AutoMigrate(ctx); err != nil {
		return nil, err
	}

	return session, nil
}

func LoadSession(name string, parentFolder string) (*Session, error) {
	session := new(Session)
	session.name = name
	session.parentFolder = parentFolder
	persistor, err := db.NewPersistor(path.Join(parentFolder, SessionsFolder, name, DBDataFile))
	if err != nil {
		return nil, err
	}
	session.persistor = persistor

	return session, nil
}

func (r *Session) Persistor() *db.Persistor {
	return r.persistor
}
