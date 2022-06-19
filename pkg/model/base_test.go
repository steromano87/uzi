package model_test

import (
	"github.com/steromano87/harkonnen/v1/pkg/model"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"testing"
	"time"
)

var sqliteInMemoryDSN = "file::memory:?cache=shared"

func TestWriteModel(t *testing.T) {
	database, _ := gorm.Open(sqlite.Open(sqliteInMemoryDSN))
	myDate, _ := time.Parse("2006-01-02 15:04", "2000-01-01 01:00")

	sampleModel := model.Base{
		Timestamp: myDate,
	}

	_ = database.AutoMigrate(sampleModel)
	result := database.Create(&sampleModel)

	if assert.NoError(t, result.Error) {
		assert.Equal(t, int64(1), result.RowsAffected)
	}
}
