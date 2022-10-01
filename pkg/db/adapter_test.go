package db_test

import (
	"context"
	"fmt"
	"github.com/Flaque/filet"
	"github.com/steromano87/harkonnen/v1/pkg/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/gorm"
	"os"
	"path"
	"testing"
)

type MockedDBEntry struct {
	gorm.Model

	Data string
}

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
	adapter := db.NewAdapter(db.SQLite, path.Join(s.tempWorkingDir, "results.db"))

	err := adapter.Connect(context.TODO())

	if assert.NoError(s.T(), err) {
		assert.FileExists(s.T(), path.Join(s.tempWorkingDir, "results.db"))
	}
}

func (s *AdapterTestSuite) TestCreateNewSQLiteDBOnMemory() {
	adapter := db.NewAdapter(db.SQLite, db.SQLiteDSNForInMemoryDB)

	err := adapter.Connect(context.TODO())
	assert.NoError(s.T(), err)
}

func (s *AdapterTestSuite) TestConnectOnMySQLDatabase() {
	if testing.Short() {
		s.T().Skip("Skipped in short mode")
	}

	ctx := context.TODO()
	dbName := "test"
	dbUser := "testuser"
	dbPassword := "testpassword"

	// Setup MySQL container
	req := testcontainers.ContainerRequest{
		Image:        "mysql:latest",
		ExposedPorts: []string{"3306/tcp"},
		WaitingFor:   wait.ForListeningPort("3306"),
		Env: map[string]string{
			"MYSQL_DATABASE":      dbName,
			"MYSQL_USER":          dbUser,
			"MYSQL_PASSWORD":      dbPassword,
			"MYSQL_ROOT_PASSWORD": dbPassword,
		},
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})

	defer func(container testcontainers.Container, ctx context.Context) {
		err := container.Terminate(ctx)
		if err != nil {
			println("Error terminating container: " + err.Error())
		}
	}(container, ctx)

	if assert.NoError(s.T(), err) {
		port, _ := container.MappedPort(ctx, "3306")
		hostIP, _ := container.Host(ctx)

		adapter := db.NewAdapter(db.MySQL, fmt.Sprintf(
			"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			dbUser,
			dbPassword,
			hostIP,
			port.Port(),
			dbName))

		err := adapter.Connect(ctx)
		if assert.NoError(s.T(), err) {
			item := MockedDBEntry{
				Data: "Random data",
			}
			err := adapter.AutoMigrate(&item)

			assert.NoError(s.T(), err)
		}
	}
}

func (s *AdapterTestSuite) TestSaveItemToDB() {
	adapter := db.NewAdapter(db.SQLite, path.Join(s.tempWorkingDir, "results.db"))

	ctx := context.TODO()
	err := adapter.Connect(ctx)

	if assert.NoError(s.T(), err) {
		item := MockedDBEntry{
			Data: "Random data",
		}

		err := adapter.AutoMigrate(&item)
		if assert.NoError(s.T(), err) {
			result := adapter.WithContext(ctx).Create(&item)

			if assert.NoError(s.T(), result.Error) {
				assert.EqualValues(s.T(), 1, result.RowsAffected)
				fileInfo, _ := os.Stat(path.Join(s.tempWorkingDir, "results.db"))
				assert.Greater(s.T(), fileInfo.Size(), int64(0))
			}
		}
	}
}

func TestAdapterTestSuite(t *testing.T) {
	suite.Run(t, new(AdapterTestSuite))
}
