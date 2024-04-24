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

	err := adapter.Connect()

	if assert.NoError(s.T(), err) {
		assert.FileExists(s.T(), path.Join(s.tempWorkingDir, "results.db"))
	}
}

func (s *AdapterTestSuite) TestCreateNewSQLiteDBOnMemory() {
	adapter := db.NewAdapter(db.SQLite, db.SQLiteDSNForInMemoryDB)

	err := adapter.Connect()
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

		err := adapter.Connect()
		if assert.NoError(s.T(), err) {
			item := MockedDBEntry{
				Data: "Random data",
			}
			err := adapter.AutoMigrate(&item)

			assert.NoError(s.T(), err)
		}
	}
}

func (s *AdapterTestSuite) TestAutoMigrate() {
	adapter := db.NewAdapter(db.SQLite, db.SQLiteDSNForInMemoryDB)

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
				assert.Len(s.T(), tables, 4)
				assert.Contains(s.T(), tables, "host_metrics")
				assert.Contains(s.T(), tables, "logs")
				assert.Contains(s.T(), tables, "samples")
				assert.Contains(s.T(), tables, "transactions")
			}
		}
	}
}

func TestAdapterTestSuite(t *testing.T) {
	suite.Run(t, new(AdapterTestSuite))
}
