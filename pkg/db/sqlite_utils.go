package db

import (
	"github.com/steromano87/harkonnen/v1/pkg/workingfolder"
	"path/filepath"
)

const SQLiteDSNForInMemoryDB = "file::memory:?cache=shared"

func SQLiteDSNForSession(workingFolder string, sessionName string) string {
	return filepath.Join(workingFolder, workingfolder.SessionsFolder, sessionName, workingfolder.DBDataFile)
}
