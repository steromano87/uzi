package db

import (
	"github.com/steromano87/harkonnen/v1/pkg/model"
	"github.com/steromano87/harkonnen/v1/pkg/runtime"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Manager struct {
	config Config
	*gorm.DB
}

func NewManager(ctx runtime.Context) *Manager {
	manager := new(Manager)
	manager.config = NewConfig(ctx)

	return manager
}

func (m *Manager) Connect() error {
	dialector, err := m.getDialector()
	if err != nil {
		return err
	}

	m.DB, err = gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		return err
	}

	err = m.migrateAll()
	return err
}

func (m *Manager) getDialector() (gorm.Dialector, error) {
	switch m.config.Type {
	case SQLite:
		return sqlite.Open(m.config.DSN), nil

	case MySQL:
		return mysql.Open(m.config.DSN), nil

	default:
		return gorm.Config{}, ErrUnknownType{unkType: m.config.Type}
	}
}

func (m *Manager) migrateAll() error {
	for _, element := range []any{&model.Log{}, &model.Sample{}} {
		err := m.AutoMigrate(element)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *Manager) Save(item any) error {
	result := m.Create(item)
	return result.Error
}
