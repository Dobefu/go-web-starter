package cmd

import (
	"testing"

	"github.com/Dobefu/go-web-starter/internal/config"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockMigrator struct {
	mock.Mock
}

func (m *MockMigrator) MigrateUp(cfg config.Database) error {
	args := m.Called(cfg)
	return args.Error(0)
}

func (m *MockMigrator) MigrateDown(cfg config.Database) error {
	args := m.Called(cfg)
	return args.Error(0)
}

func (m *MockMigrator) MigrateVersion(cfg config.Database) (int, error) {
	args := m.Called(cfg)
	return args.Int(0), args.Error(1)
}

func TestDefaultDatabaseMigrator(t *testing.T) {
	var _ DatabaseMigrator = &defaultDatabaseMigrator{}
	assert.NotNil(t, &defaultDatabaseMigrator{})

	cfg := config.Database{
		Host:     "localhost",
		Port:     5432,
		User:     "test",
		Password: "test",
		DBName:   "test",
	}

	migrator := &defaultDatabaseMigrator{}
	assert.Panics(t, func() { _ = migrator.MigrateUp(cfg) })
	assert.Panics(t, func() { _ = migrator.MigrateDown(cfg) })
	assert.Panics(t, func() { _, _ = migrator.MigrateVersion(cfg) })
}

func TestMigrateCommands(t *testing.T) {
	testCommand := func(
		t *testing.T,
		command func(*cobra.Command, []string),
		mockSetup func(*MockMigrator),
		expectPanic bool,
		expectedError error,
	) {
		mockMigrator := new(MockMigrator)
		originalMigrator := migrator
		migrator = mockMigrator
		defer func() { migrator = originalMigrator }()

		mockSetup(mockMigrator)
		cmd := &cobra.Command{}
		args := []string{}

		if expectPanic {
			assert.PanicsWithValue(t, expectedError, func() {
				command(cmd, args)
			})
		} else {
			assert.NotPanics(t, func() {
				command(cmd, args)
			})
		}
		mockMigrator.AssertExpectations(t)
	}

	t.Run("migrate up success", func(t *testing.T) {
		testCommand(t, migrateUp, func(m *MockMigrator) {
			m.On("MigrateUp", mock.Anything).Return(nil)
		}, false, nil)
	})

	t.Run("migrate up failure", func(t *testing.T) {
		testCommand(t, migrateUp, func(m *MockMigrator) {
			m.On("MigrateUp", mock.Anything).Return(assert.AnError)
		}, true, assert.AnError)
	})

	t.Run("migrate down success", func(t *testing.T) {
		testCommand(t, migrateDown, func(m *MockMigrator) {
			m.On("MigrateDown", mock.Anything).Return(nil)
		}, false, nil)
	})

	t.Run("migrate down failure", func(t *testing.T) {
		testCommand(t, migrateDown, func(m *MockMigrator) {
			m.On("MigrateDown", mock.Anything).Return(assert.AnError)
		}, true, assert.AnError)
	})

	t.Run("migrate version success", func(t *testing.T) {
		testCommand(t, migrateVersion, func(m *MockMigrator) {
			m.On("MigrateVersion", mock.Anything).Return(5, nil)
		}, false, nil)
	})

	t.Run("migrate version failure", func(t *testing.T) {
		testCommand(t, migrateVersion, func(m *MockMigrator) {
			m.On("MigrateVersion", mock.Anything).Return(0, assert.AnError)
		}, true, assert.AnError)
	})
}
