package cmd

import (
	"errors"
	"io"
	"os"
	"testing"

	"github.com/Dobefu/go-web-starter/internal/config"
	"github.com/Dobefu/go-web-starter/internal/database"
	"github.com/Dobefu/go-web-starter/internal/logger"
	"github.com/Dobefu/go-web-starter/internal/user"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

type mockUserRepo struct {
	FindByEmailFunc func(email string) (*user.User, error)
	SaveUserFunc    func(u *user.User) error
}

func (m *mockUserRepo) FindByEmail(email string) (*user.User, error) {
	return m.FindByEmailFunc(email)
}

func (m *mockUserRepo) SaveUser(u *user.User) error {
	return m.SaveUserFunc(u)
}

func TestGetUserDetails_AllFlagsProvided(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("username", "testuser", "")
	cmd.Flags().String("email", "test@example.com", "")
	cmd.Flags().String("password", "secret", "")

	username, email, password, err := getUserDetails(cmd)
	assert.NoError(t, err)
	assert.Equal(t, "testuser", username)
	assert.Equal(t, "test@example.com", email)
	assert.Equal(t, "secret", password)
}

func TestGetUserDetails_PromptForMissing(t *testing.T) {
	originalPromptForString := promptForString
	originalPromptForPassword := promptForPassword

	defer func() {
		promptForString = originalPromptForString
		promptForPassword = originalPromptForPassword
	}()

	promptForString = func(prompt string) (string, error) {
		if prompt == "Enter Email: " {
			return "test@example.com", nil
		}

		return "testuser", nil
	}

	promptForPassword = func(prompt string) (string, error) {
		return "secret", nil
	}

	cmd := &cobra.Command{}
	username, email, password, err := getUserDetails(cmd)
	assert.NoError(t, err)
	assert.Equal(t, "testuser", username)
	assert.Equal(t, "test@example.com", email)
	assert.Equal(t, "secret", password)
}

func TestGetUserDetails_ErrorOnPrompt(t *testing.T) {
	originalPromptForString := promptForString
	defer func() { promptForString = originalPromptForString }()

	promptForString = func(prompt string) (string, error) {
		return "", errors.New("input error")
	}

	cmd := &cobra.Command{}
	_, _, _, err := getUserDetails(cmd)
	assert.Error(t, err)
}

func TestGetUserDetails_EmptyFields(t *testing.T) {
	originalPromptForString := promptForString
	originalPromptForPassword := promptForPassword

	defer func() {
		promptForString = originalPromptForString
		promptForPassword = originalPromptForPassword
	}()

	promptForString = func(prompt string) (string, error) { return "", nil }
	promptForPassword = func(prompt string) (string, error) { return "", nil }

	cmd := &cobra.Command{}
	_, _, _, err := getUserDetails(cmd)
	assert.Error(t, err)
}

func TestGetUserDetails_PromptForPasswordOnly(t *testing.T) {
	originalPromptForPassword := promptForPassword
	defer func() { promptForPassword = originalPromptForPassword }()

	promptForPassword = func(prompt string) (string, error) {
		return "secret", nil
	}

	cmd := &cobra.Command{}
	cmd.Flags().String("username", "testuser", "")
	cmd.Flags().String("email", "test@example.com", "")

	username, email, password, err := getUserDetails(cmd)
	assert.NoError(t, err)
	assert.Equal(t, "testuser", username)
	assert.Equal(t, "test@example.com", email)
	assert.Equal(t, "secret", password)
}

func TestGetUserDetails_PromptForUsernameOnly(t *testing.T) {
	originalPromptForString := promptForString
	defer func() { promptForString = originalPromptForString }()

	promptForString = func(prompt string) (string, error) {
		return "testuser", nil
	}

	cmd := &cobra.Command{}
	cmd.Flags().String("email", "test@example.com", "")
	cmd.Flags().String("password", "secret", "")

	username, email, password, err := getUserDetails(cmd)
	assert.NoError(t, err)
	assert.Equal(t, "testuser", username)
	assert.Equal(t, "test@example.com", email)
	assert.Equal(t, "secret", password)
}

func TestRunCreateUser_Success(t *testing.T) {
	log := logger.New(logger.InfoLevel, io.Discard)

	repo := &mockUserRepo{
		FindByEmailFunc: func(email string) (*user.User, error) {
			return nil, user.ErrInvalidCredentials
		},
		SaveUserFunc: func(u *user.User) error {
			return nil
		},
	}

	usr, err := runCreateUser(repo, log, "testuser", "test@example.com", "secret")
	assert.NoError(t, err)
	assert.NotNil(t, usr)
	assert.Equal(t, "testuser", usr.GetUsername())
}

func TestRunCreateUser_Error_UserExists(t *testing.T) {
	log := logger.New(logger.InfoLevel, io.Discard)

	repo := &mockUserRepo{
		FindByEmailFunc: func(email string) (*user.User, error) {
			return &user.User{}, nil
		},
		SaveUserFunc: func(u *user.User) error {
			return nil
		},
	}

	usr, err := runCreateUser(repo, log, "testuser", "test@example.com", "secret")
	assert.Error(t, err)
	assert.Nil(t, usr)
}

func TestRunCreateUser_Error_SaveFails(t *testing.T) {
	log := logger.New(logger.InfoLevel, io.Discard)

	repo := &mockUserRepo{
		FindByEmailFunc: func(email string) (*user.User, error) {
			return nil, user.ErrInvalidCredentials
		},
		SaveUserFunc: func(u *user.User) error {
			return errors.New("save error")
		},
	}

	usr, err := runCreateUser(repo, log, "testuser", "test@example.com", "secret")
	assert.Error(t, err)
	assert.Nil(t, usr)
}

func captureStdoutStderr(f func()) (string, string) {
	oldStdout := os.Stdout
	oldStderr := os.Stderr

	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()

	os.Stdout = wOut
	os.Stderr = wErr

	f()

	_ = wOut.Close()
	_ = wErr.Close()

	os.Stdout = oldStdout
	os.Stderr = oldStderr

	var bufOut, bufErr []byte

	bufOut = make([]byte, 1024)
	bufErr = make([]byte, 1024)

	nOut, _ := rOut.Read(bufOut)
	nErr, _ := rErr.Read(bufErr)

	return string(bufOut[:nOut]), string(bufErr[:nErr])
}

func TestRunCreateUserCmd_Success(t *testing.T) {
	cmd := &cobra.Command{}

	deps := userCreateDeps{
		getUserDetails: func(cmd *cobra.Command) (string, string, string, error) {
			return "testuser", "test@example.com", "secret", nil
		},
		dbNew: func(cfg config.Database, log *logger.Logger) (database.DatabaseInterface, error) {
			return &mockDB{}, nil
		},
		runCreateUser: func(repo user.UserRepository, log *logger.Logger, username, email, password string) (*user.User, error) {
			return &user.User{}, nil
		},
		osExit: func(int) {},
	}

	out, errOut := captureStdoutStderr(func() {
		runCreateUserCmdWithDeps(cmd, []string{}, deps)
	})

	assert.Contains(t, out, "User created successfully!")
	assert.Empty(t, errOut)
}

func TestRunCreateUserCmd_GetUserDetailsError(t *testing.T) {
	cmd := &cobra.Command{}
	exitCalled := false

	deps := userCreateDeps{
		getUserDetails: func(cmd *cobra.Command) (string, string, string, error) {
			return "", "", "", errors.New("input error")
		},
		dbNew: func(cfg config.Database, log *logger.Logger) (database.DatabaseInterface, error) {
			return &mockDB{}, nil
		},
		runCreateUser: func(repo user.UserRepository, log *logger.Logger, username, email, password string) (*user.User, error) {
			return &user.User{}, nil
		},
		osExit: func(int) { exitCalled = true },
	}

	_, _ = captureStdoutStderr(func() {
		runCreateUserCmdWithDeps(cmd, []string{}, deps)
	})

	assert.True(t, exitCalled)
}

func TestRunCreateUserCmd_DatabaseError(t *testing.T) {
	cmd := &cobra.Command{}
	exitCalled := false

	deps := userCreateDeps{
		getUserDetails: func(cmd *cobra.Command) (string, string, string, error) {
			return "testuser", "test@example.com", "secret", nil
		},
		dbNew: func(cfg config.Database, log *logger.Logger) (database.DatabaseInterface, error) {
			return nil, errors.New("db error")
		},
		runCreateUser: func(repo user.UserRepository, log *logger.Logger, username, email, password string) (*user.User, error) {
			return &user.User{}, nil
		},
		osExit: func(int) { exitCalled = true },
	}

	_, _ = captureStdoutStderr(func() {
		runCreateUserCmdWithDeps(cmd, []string{}, deps)
	})

	assert.True(t, exitCalled)
}

func TestRunCreateUserCmd_RunCreateUserError(t *testing.T) {
	cmd := &cobra.Command{}
	exitCalled := false

	deps := userCreateDeps{
		getUserDetails: func(cmd *cobra.Command) (string, string, string, error) {
			return "testuser", "test@example.com", "secret", nil
		},
		dbNew: func(cfg config.Database, log *logger.Logger) (database.DatabaseInterface, error) {
			return &mockDB{}, nil
		},
		runCreateUser: func(repo user.UserRepository, log *logger.Logger, username, email, password string) (*user.User, error) {
			return nil, errors.New("create error")
		},
		osExit: func(int) { exitCalled = true },
	}

	_, errOut := captureStdoutStderr(func() {
		runCreateUserCmdWithDeps(cmd, []string{}, deps)
	})

	assert.True(t, exitCalled)
	assert.Contains(t, errOut, "Error creating user.")
}

func TestDefaultUserCreateDeps(t *testing.T) {
	deps := defaultUserCreateDeps()
	assert.NotNil(t, deps.getUserDetails)
	assert.NotNil(t, deps.dbNew)
	assert.NotNil(t, deps.runCreateUser)
	assert.NotNil(t, deps.osExit)
}

func TestRunCreateUserCmd_ProductionEntrypoint(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("username", "", "")
	cmd.Flags().String("email", "", "")
	cmd.Flags().String("password", "", "")

	originalOsExit := osExit
	osExitCalled := false
	osExit = func(code int) { osExitCalled = true }
	defer func() { osExit = originalOsExit }()

	runCreateUserCmd(cmd, []string{})
	assert.True(t, osExitCalled)
}

func TestGetUserDetails_ErrorPromptForStringEmail(t *testing.T) {
	originalPromptForString := promptForString
	defer func() { promptForString = originalPromptForString }()

	promptForString = func(prompt string) (string, error) {
		if prompt == "Enter Email: " {
			return "", errors.New("email error")
		}

		return "", nil
	}

	cmd := &cobra.Command{}
	cmd.Flags().String("username", "", "")
	cmd.Flags().String("email", "", "")
	cmd.Flags().String("password", "secret", "")

	_, _, _, err := getUserDetails(cmd)
	assert.Error(t, err)
}

func TestGetUserDetails_ErrorPromptForStringUsername(t *testing.T) {
	originalPromptForString := promptForString
	defer func() { promptForString = originalPromptForString }()

	promptForString = func(prompt string) (string, error) {
		if prompt == "Enter Username: " {
			return "", errors.New("username error")
		}

		return "test@example.com", nil
	}

	cmd := &cobra.Command{}
	cmd.Flags().String("username", "", "")
	cmd.Flags().String("email", "", "")
	cmd.Flags().String("password", "secret", "")

	_, _, _, err := getUserDetails(cmd)
	assert.Error(t, err)
}

func TestGetUserDetails_ErrorPromptForPassword(t *testing.T) {
	originalPromptForPassword := promptForPassword
	defer func() { promptForPassword = originalPromptForPassword }()

	promptForPassword = func(prompt string) (string, error) {
		return "", errors.New("password error")
	}

	cmd := &cobra.Command{}
	cmd.Flags().String("username", "testuser", "")
	cmd.Flags().String("email", "test@example.com", "")
	cmd.Flags().String("password", "", "")

	_, _, _, err := getUserDetails(cmd)
	assert.Error(t, err)
}

func TestGetUserDetails_AllEmptyFlagsAndPrompts(t *testing.T) {
	originalPromptForString := promptForString
	originalPromptForPassword := promptForPassword

	defer func() {
		promptForString = originalPromptForString
		promptForPassword = originalPromptForPassword
	}()

	promptForString = func(prompt string) (string, error) { return "", nil }
	promptForPassword = func(prompt string) (string, error) { return "", nil }

	cmd := &cobra.Command{}
	cmd.Flags().String("username", "", "")
	cmd.Flags().String("email", "", "")
	cmd.Flags().String("password", "", "")

	_, _, _, err := getUserDetails(cmd)
	assert.Error(t, err)
}

func TestGetUserDetails_FlagsSetButEmpty(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("username", "", "")
	cmd.Flags().String("email", "", "")
	cmd.Flags().String("password", "", "")

	_, _, _, err := getUserDetails(cmd)
	assert.Error(t, err)
}
