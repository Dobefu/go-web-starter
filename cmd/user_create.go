package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/Dobefu/go-web-starter/internal/config"
	"github.com/Dobefu/go-web-starter/internal/database"
	"github.com/Dobefu/go-web-starter/internal/logger"
	"github.com/Dobefu/go-web-starter/internal/user"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var createUserCmd = &cobra.Command{
	Use:   "user:create",
	Short: "Create a new user in the database",
	Run:   runCreateUserCmd,
}

func init() {
	rootCmd.AddCommand(createUserCmd)

	createUserCmd.Flags().StringP("username", "u", "", "Username for the new user")
	createUserCmd.Flags().StringP("email", "e", "", "Email for the new user")
	createUserCmd.Flags().StringP("password", "p", "", "Password for the new user")
}

func getUserDetails(cmd *cobra.Command) (username, email, password string, err error) {
	log := logger.New(logger.Level(config.GetLogLevel()), os.Stdout)

	username, _ = cmd.Flags().GetString("username")
	email, _ = cmd.Flags().GetString("email")
	password, _ = cmd.Flags().GetString("password")

	if email == "" {
		email, err = promptForString("Enter Email: ")

		if err != nil {
			log.Error("Failed to get email input", logger.Fields{"error": err.Error()})
			return "", "", "", err
		}
	}

	if username == "" {
		username, err = promptForString("Enter Username: ")

		if err != nil {
			log.Error("Failed to get username input", logger.Fields{"error": err.Error()})
			return "", "", "", err
		}
	}

	if password == "" {
		password, err = promptForPassword("Enter Password: ")

		if err != nil {
			log.Error("Failed to get password input", logger.Fields{"error": err.Error()})
			return "", "", "", err
		}
	}

	if username == "" || email == "" || password == "" {
		err = errors.New("username, email, and password cannot be empty")
		log.Error(err.Error(), nil)

		return "", "", "", err
	}

	return username, email, password, nil
}

func runCreateUser(db database.DatabaseInterface, log *logger.Logger, username, email, password string) (*user.User, error) {
	log.Info("Attempting user creation in core logic...", logger.Fields{"email": email, "username": username})

	createdUser, createErr := user.Create(db, username, email, password)

	if createErr != nil {
		log.Error("user.Create failed", logger.Fields{
			"email":    email,
			"username": username,
			"error":    createErr.Error(),
		})

		return nil, createErr
	}

	log.Info("Successfully created user in core logic", logger.Fields{
		"email":    email,
		"username": username,
		"userID":   createdUser.GetID(),
	})

	return createdUser, nil
}

func runCreateUserCmd(cmd *cobra.Command, args []string) {
	log := logger.New(logger.Level(config.GetLogLevel()), os.Stdout)

	username, email, password, err := getUserDetails(cmd)
	if err != nil {
		osExit(1)
		return
	}

	dbConfig := getDatabaseConfigForCmd()
	db, dbErr := database.New(dbConfig, log)
	if dbErr != nil {
		log.Error("Failed to connect to database", logger.Fields{"error": dbErr.Error()})
		osExit(1)
		return
	}
	defer func() { _ = db.Close() }()

	_, runErr := runCreateUser(db, log, username, email, password)

	if runErr != nil {
		fmt.Fprintln(os.Stderr, "Error creating user.")
		osExit(1)

		return
	}

	fmt.Println("User created successfully!")
}

func getDatabaseConfigForCmd() config.Database {
	return config.Database{
		Host:     viper.GetString("database.host"),
		Port:     viper.GetInt("database.port"),
		User:     viper.GetString("database.user"),
		Password: viper.GetString("database.password"),
		DBName:   viper.GetString("database.dbname"),
	}
}
