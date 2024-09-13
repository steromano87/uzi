package db_test

import (
	"context"
	"github.com/Flaque/filet"
	"github.com/steromano87/harkonnen/v1/pkg/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"path"
	"testing"
)

type AdapterTestSuite struct {
	suite.Suite
	tempWorkingDir string
}

func (s *AdapterTestSuite) SetupTest() {
	s.tempWorkingDir = filet.TmpDir(s.T(), "")
}

func (s *AdapterTestSuite) TearDownTest() {
	filet.CleanUp(s.T())
}

func (s *AdapterTestSuite) TestCreateNewSQLiteDBOnFile() {
	adapter := db.NewAdapter(path.Join(s.tempWorkingDir, "results.db"))

	err := adapter.Connect()

	if assert.NoError(s.T(), err) {
		assert.FileExists(s.T(), path.Join(s.tempWorkingDir, "results.db"))
	}
}

func (s *AdapterTestSuite) TestCreateNewSQLiteDBOnMemory() {
	adapter := db.NewAdapter(db.SQLiteDSNForInMemoryDB)

	err := adapter.Connect()
	assert.NoError(s.T(), err)
}

func (s *AdapterTestSuite) TestAutoMigrate() {
	adapter := db.NewAdapter(db.SQLiteDSNForInMemoryDB)

	ctx := context.TODO()
	err := adapter.Connect()

	if assert.NoError(s.T(), err) {
		err := adapter.MigrateAll(ctx)
		if assert.NoError(s.T(), err) {
			var tables []string
			err := adapter.Table("sqlite_schema").Where(
				"name not like ?", "sqlite_%").Where(
				"name not like ?", "idx_%").Pluck("name", &tables).Error
			if assert.NoError(s.T(), err) {
				assert.Len(s.T(), tables, 5)
				assert.Contains(s.T(), tables, "host_metrics")
				assert.Contains(s.T(), tables, "raw_logs")
				assert.Contains(s.T(), tables, "samples")
				assert.Contains(s.T(), tables, "transactions")
				assert.Contains(s.T(), tables, "iteration_counters")
			}
		}
	}
}

func TestAdapterTestSuite(t *testing.T) {
	suite.Run(t, new(AdapterTestSuite))
}
