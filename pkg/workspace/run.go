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

type Run struct {
	name         string
	parentFolder string
	persistor    *db.Persistor
}

func NewRun(ctx context.Context, name string, parentFolder string) (*Run, error) {
	run := new(Run)
	run.name = name
	run.parentFolder = parentFolder
	if err := os.MkdirAll(path.Join(parentFolder, RunsFolder, name), FolderPerms); err != nil {
		return nil, err
	}

	persistor, err := db.NewPersistor(path.Join(parentFolder, RunsFolder, name, DBDataFile))
	if err != nil {
		return nil, err
	}
	run.persistor = persistor

	if err := run.persistor.AutoMigrate(ctx); err != nil {
		return nil, err
	}

	return run, nil
}

func LoadRun(name string, parentFolder string) (*Run, error) {
	run := new(Run)
	run.name = name
	run.parentFolder = parentFolder
	persistor, err := db.NewPersistor(path.Join(parentFolder, RunsFolder, name, DBDataFile))
	if err != nil {
		return nil, err
	}
	run.persistor = persistor

	return run, nil
}

func (r *Run) Persistor() *db.Persistor {
	return r.persistor
}
