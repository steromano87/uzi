package collector_test

import (
	"context"
	"fmt"
	"github.com/Flaque/filet"
	"github.com/steromano87/harkonnen/v1/internal/collector"
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

type ManagerTestSuite struct {
	suite.Suite
	tempWorkingDir string
}

func (s *ManagerTestSuite) SetupTest() {
	s.tempWorkingDir = filet.TmpDir(s.T(), "")
}

func (s *ManagerTestSuite) TearDownTest() {
	filet.CleanUp(s.T())
}

func (s *ManagerTestSuite) TestCreateNewSQLiteDBOnFile() {
	writer := collector.DbWriter{
		Type: collector.SQLite,
		DSN:  path.Join(s.tempWorkingDir, "results.db"),
	}

	err := writer.Connect(context.TODO())

	if assert.NoError(s.T(), err) {
		assert.FileExists(s.T(), path.Join(s.tempWorkingDir, "results.db"))
	}
}

func (s *ManagerTestSuite) TestCreateNewSQLiteDBOnMemory() {
	writer := collector.DbWriter{
		Type: collector.SQLite,
		DSN:  "file::memory:?cache=shared",
	}

	err := writer.Connect(context.TODO())
	assert.NoError(s.T(), err)
}

func (s *ManagerTestSuite) TestConnectOnMySQLDatabase() {
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

	defer container.Terminate(ctx)

	if assert.NoError(s.T(), err) {
		port, _ := container.MappedPort(ctx, "3306")
		hostIP, _ := container.Host(ctx)

		writer := collector.DbWriter{
			Type: collector.MySQL,
			DSN: fmt.Sprintf(
				"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
				dbUser,
				dbPassword,
				hostIP,
				port.Port(),
				dbName),
		}

		err := writer.Connect(ctx)
		if assert.NoError(s.T(), err) {
			item := MockedDBEntry{
				Data: "Random data",
			}
			err := writer.AutoMigrate(&item)

			assert.NoError(s.T(), err)
		}
	}
}

func (s *ManagerTestSuite) TestSaveItemToDB() {
	writer := collector.DbWriter{
		Type: collector.SQLite,
		DSN:  path.Join(s.tempWorkingDir, "results.db"),
	}

	ctx := context.TODO()
	err := writer.Connect(ctx)

	if assert.NoError(s.T(), err) {
		item := MockedDBEntry{
			Data: "Random data",
		}

		err := writer.AutoMigrate(&item)
		if assert.NoError(s.T(), err) {
			result := writer.Save(ctx, &item)

			if assert.NoError(s.T(), result) {
				fileInfo, _ := os.Stat(path.Join(s.tempWorkingDir, "results.db"))
				assert.Greater(s.T(), fileInfo.Size(), int64(0))
			}
		}
	}
}

func TestManagerTestSuite(t *testing.T) {
	suite.Run(t, new(ManagerTestSuite))
}
