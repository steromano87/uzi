package db_test

import (
	"context"
	"github.com/Flaque/filet"
	"github.com/steromano87/harkonnen/v1/internal/db"
	"github.com/steromano87/harkonnen/v1/pkg/project"
	"github.com/steromano87/harkonnen/v1/pkg/runtime"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
	"os"
	"path"
	"testing"
)

type MockedDBEntry struct {
	gorm.Model

	Data string
}

type ManagerTestSuite struct {
	suite.Suite
	tempWorkingDir string
}

func (s *ManagerTestSuite) SetupTest() {
	s.tempWorkingDir = filet.TmpDir(s.T(), "")
	filet.File(s.T(), path.Join(s.tempWorkingDir, "harkonnen.yaml"), "")
}

func (s *ManagerTestSuite) TearDownTest() {
	filet.CleanUp(s.T())
	os.Unsetenv("HARK_DB_DSN")
}

func (s *ManagerTestSuite) TestNewManager() {
	os.Setenv("HARK_DB_DSN", "file::memory:?cache=shared")

	config, err := project.NewConfig(s.tempWorkingDir)

	if assert.NoError(s.T(), err) {
		ctx := runtime.NewContext(context.TODO(), *config)
		manager := db.NewManager(ctx)

		assert.IsType(s.T(), &db.Manager{}, manager)
	}
}

func (s *ManagerTestSuite) TestCreateNewSQLiteDBOnFile() {
	os.Setenv("HARK_DB_DSN", path.Join(s.tempWorkingDir, "results.db"))

	config, _ := project.NewConfig(s.tempWorkingDir)
	ctx := runtime.NewContext(context.TODO(), *config)

	manager := db.NewManager(ctx)

	err := manager.Connect()

	if assert.NoError(s.T(), err) {
		assert.FileExists(s.T(), path.Join(s.tempWorkingDir, "results.db"))
	}
}

func (s *ManagerTestSuite) TestCreateNewSQLiteDBOnMemory() {
	os.Setenv("HARK_DB_DSN", "file::memory:?cache=shared")

	config, _ := project.NewConfig(s.tempWorkingDir)
	ctx := runtime.NewContext(context.TODO(), *config)

	manager := db.NewManager(ctx)

	err := manager.Connect()
	assert.NoError(s.T(), err)
}

func (s *ManagerTestSuite) TestSaveItemToDB() {
	os.Setenv("HARK_DB_DSN", path.Join(s.tempWorkingDir, "results.db"))

	config, _ := project.NewConfig(s.tempWorkingDir)
	ctx := runtime.NewContext(context.TODO(), *config)

	manager := db.NewManager(ctx)

	err := manager.Connect()

	if assert.NoError(s.T(), err) {
		item := MockedDBEntry{
			Data: "Random data",
		}

		err := manager.AutoMigrate(&item)
		if assert.NoError(s.T(), err) {
			result := manager.Create(&item)

			if assert.NoError(s.T(), result.Error) {
				assert.Equal(s.T(), int64(1), result.RowsAffected)

				fileInfo, _ := os.Stat(path.Join(s.tempWorkingDir, "results.db"))
				assert.Greater(s.T(), fileInfo.Size(), int64(0))
			}
		}
	}
}

func TestManagerTestSuite(t *testing.T) {
	suite.Run(t, new(ManagerTestSuite))
}
