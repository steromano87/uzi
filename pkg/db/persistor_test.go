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

type PersistorTestSuite struct {
	suite.Suite
	tempWorkingDir string
}

func (s *PersistorTestSuite) SetupTest() {
	s.tempWorkingDir = filet.TmpDir(s.T(), "")
}

func (s *PersistorTestSuite) TearDownTest() {
	filet.CleanUp(s.T())
}

func (s *PersistorTestSuite) TestCreateNewSQLiteDBOnFile() {
	persistor, err := db.NewPersistor(path.Join(s.tempWorkingDir, "results.db"))

	if assert.NoError(s.T(), err) {
		assert.IsType(s.T(), &db.Persistor{}, persistor)
		assert.FileExists(s.T(), path.Join(s.tempWorkingDir, "results.db"))
	}
}

func (s *PersistorTestSuite) TestCreateNewSQLiteDBOnMemory() {
	_, err := db.NewPersistor(db.SQLiteDSNForInMemoryDB)

	assert.NoError(s.T(), err)
}

func (s *PersistorTestSuite) TestAutoMigrate() {
	persistor, err := db.NewPersistor(db.SQLiteDSNForInMemoryDB)

	ctx := context.TODO()

	if assert.NoError(s.T(), err) {
		err := persistor.AutoMigrate(ctx)
		if assert.NoError(s.T(), err) {
			var tables []string
			err := persistor.DB().Table("sqlite_schema").Where(
				"name not like ?", "sqlite_%").Where(
				"name not like ?", "idx_%").Pluck("name", &tables).Error
			if assert.NoError(s.T(), err) {
				assert.Len(s.T(), tables, 6)
				assert.Contains(s.T(), tables, "host_metrics")
				assert.Contains(s.T(), tables, "raw_logs")
				assert.Contains(s.T(), tables, "samples")
				assert.Contains(s.T(), tables, "transactions")
				assert.Contains(s.T(), tables, "iteration_counters")
				assert.Contains(s.T(), tables, "synthetic_user_counters")
			}
		}
	}
}

func TestAdapterTestSuite(t *testing.T) {
	suite.Run(t, new(PersistorTestSuite))
}
