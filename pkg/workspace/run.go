package workspace

import (
	"context"
	"github.com/steromano87/harkonnen/v1/pkg/db"
	"gorm.io/gorm"
	"os"
	"path"
)

const (
	DBDataFile = "data.db"
)

type Run struct {
	name         string
	parentFolder string
	dbAdapter    *db.Adapter
}

func NewRun(ctx context.Context, name string, parentFolder string) (*Run, error) {
	run := new(Run)
	run.name = name
	run.parentFolder = parentFolder
	if err := os.MkdirAll(path.Join(parentFolder, RunsFolder, name), FolderPerms); err != nil {
		return nil, err
	}

	run.dbAdapter = db.NewAdapter(path.Join(parentFolder, RunsFolder, name, DBDataFile))

	if err := run.dbAdapter.Connect(); err != nil {
		return nil, err
	}

	if err := run.dbAdapter.MigrateAll(ctx); err != nil {
		return nil, err
	}

	return run, nil
}

func LoadRun(name string, parentFolder string) *Run {
	run := new(Run)
	run.name = name
	run.parentFolder = parentFolder
	run.dbAdapter = db.NewAdapter(path.Join(parentFolder, RunsFolder, name, DBDataFile))

	return run
}

func (r *Run) DB() *gorm.DB {
	return r.dbAdapter.DB
}
