package cmd

import (
	"bytes"
	"errors"
	"io"
	"os"
	"testing"
	"time"

	"github.com/Dobefu/go-web-starter/internal/config"
	"github.com/Dobefu/go-web-starter/internal/database"
	"github.com/Dobefu/go-web-starter/internal/logger"
	"github.com/Dobefu/go-web-starter/internal/user"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestRunUserDetails_ID_Success(t *testing.T) {
	findByID := func(db database.DatabaseInterface, id int) (*user.User, error) {
		return user.New(42, "foo", "foo@bar.com", "", true, time.Now(), time.Now(), time.Now()), nil
	}

	findByEmail := func(db database.DatabaseInterface, email string) (*user.User, error) {
		return nil, errors.New("should not be called")
	}

	log := logger.New(logger.InfoLevel, io.Discard)
	db := &mockDB{}

	usr, err := runUserDetails(db, log, "42", findByID, findByEmail)
	assert.NoError(t, err)
	assert.NotNil(t, usr)
	assert.Equal(t, 42, (*usr).GetID())
}

func TestRunUserDetails_Email_Success(t *testing.T) {
	findByID := func(db database.DatabaseInterface, id int) (*user.User, error) {
		return nil, errors.New("should not be called")
	}

	findByEmail := func(db database.DatabaseInterface, email string) (*user.User, error) {
		return user.New(7, "bar", email, "", false, time.Now(), time.Now(), time.Now()), nil
	}

	log := logger.New(logger.InfoLevel, io.Discard)
	db := &mockDB{}

	usr, err := runUserDetails(db, log, "bar@baz.com", findByID, findByEmail)
	assert.NoError(t, err)
	assert.NotNil(t, usr)
	assert.Equal(t, "bar@baz.com", (*usr).GetEmail())
}

func TestRunUserDetails_NotFound(t *testing.T) {
	findByID := func(db database.DatabaseInterface, id int) (*user.User, error) {
		return nil, user.ErrInvalidCredentials
	}

	findByEmail := func(db database.DatabaseInterface, email string) (*user.User, error) {
		return nil, user.ErrInvalidCredentials
	}

	log := logger.New(logger.InfoLevel, io.Discard)
	db := &mockDB{}

	usr, err := runUserDetails(db, log, "999", findByID, findByEmail)
	assert.Error(t, err)
	assert.Nil(t, usr)
}

func captureStdout(f func()) string {
	r, w, _ := os.Pipe()
	old := os.Stdout
	os.Stdout = w
	f()

	_ = w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)

	return buf.String()
}

func TestRunUserDetailsCmd_FlagID_Success(t *testing.T) {
	mockDBNew := func(cfg config.Database, log *logger.Logger) (database.DatabaseInterface, error) {
		return &mockDB{}, nil
	}

	findByID := func(db database.DatabaseInterface, id int) (*user.User, error) {
		return user.New(1, "test", "test@x.com", "", true, time.Now(), time.Now(), time.Now()), nil
	}

	findByEmail := func(db database.DatabaseInterface, email string) (*user.User, error) {
		return nil, errors.New("should not be called")
	}

	originalOsExit := osExit
	defer func() { osExit = originalOsExit }()

	osExitCalled := false
	osExit = func(code int) { osExitCalled = true }

	cmd := &cobra.Command{}
	cmd.Flags().Int("id", 1, "")
	cmd.Flags().String("email", "", "")

	output := captureStdout(func() {
		runUserDetailsCmdWithDeps(cmd, []string{}, userDetailsDeps{
			dbNew:       mockDBNew,
			findByID:    findByID,
			findByEmail: findByEmail,
		})
	})

	assert.Contains(t, output, "User Details")
	assert.False(t, osExitCalled)
}

func TestRunUserDetailsCmd_FlagEmail_Success(t *testing.T) {
	mockDBNew := func(cfg config.Database, log *logger.Logger) (database.DatabaseInterface, error) {
		return &mockDB{}, nil
	}

	findByID := func(db database.DatabaseInterface, id int) (*user.User, error) {
		return nil, errors.New("should not be called")
	}

	findByEmail := func(db database.DatabaseInterface, email string) (*user.User, error) {
		return user.New(2, "em", email, "", true, time.Now(), time.Now(), time.Now()), nil
	}

	originalOsExit := osExit
	defer func() { osExit = originalOsExit }()

	osExitCalled := false
	osExit = func(code int) { osExitCalled = true }

	cmd := &cobra.Command{}
	cmd.Flags().Int("id", 0, "")
	cmd.Flags().String("email", "em@x.com", "")

	output := captureStdout(func() {
		runUserDetailsCmdWithDeps(cmd, []string{}, userDetailsDeps{
			dbNew:       mockDBNew,
			findByID:    findByID,
			findByEmail: findByEmail,
		})
	})

	assert.Contains(t, output, "User Details")
	assert.False(t, osExitCalled)
}

func TestRunUserDetailsCmd_Prompt_Error(t *testing.T) {
	originalPrompt := promptForString
	defer func() { promptForString = originalPrompt }()

	promptForString = func(prompt string) (string, error) {
		return "", errors.New("input error")
	}

	originalOsExit := osExit
	defer func() { osExit = originalOsExit }()

	osExitCalled := false
	osExit = func(code int) { osExitCalled = true }

	cmd := &cobra.Command{}
	cmd.Flags().Int("id", 0, "")
	cmd.Flags().String("email", "", "")

	output := captureStdout(func() {
		runUserDetailsCmd(cmd, []string{})
	})

	assert.True(t, osExitCalled)
	assert.NotContains(t, output, "User Details")
}

func TestRunUserDetailsCmd_UserNotFound(t *testing.T) {
	originalPrompt := promptForString
	defer func() { promptForString = originalPrompt }()

	promptForString = func(prompt string) (string, error) {
		return "999", nil
	}

	originalOsExit := osExit
	defer func() { osExit = originalOsExit }()

	osExitCalled := false
	osExit = func(code int) { osExitCalled = true }

	cmd := &cobra.Command{}
	cmd.Flags().Int("id", 0, "")
	cmd.Flags().String("email", "", "")

	output := captureStdout(func() {
		runUserDetailsCmd(cmd, []string{})
	})

	assert.True(t, osExitCalled)
	assert.NotContains(t, output, "User Details")
}
