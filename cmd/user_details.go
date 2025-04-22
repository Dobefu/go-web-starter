package cmd

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/Dobefu/go-web-starter/internal/config"
	"github.com/Dobefu/go-web-starter/internal/database"
	"github.com/Dobefu/go-web-starter/internal/logger"
	"github.com/Dobefu/go-web-starter/internal/user"
	"github.com/spf13/cobra"
)

var userDetailsCmd = &cobra.Command{
	Use:   "user:details",
	Short: "Show details for a specific user",
	Run:   runUserDetailsCmd,
}

func init() {
	rootCmd.AddCommand(userDetailsCmd)

	userDetailsCmd.Flags().StringP("email", "e", "", "Email of the user to show details for")
	userDetailsCmd.Flags().IntP("id", "i", 0, "ID of the user to show details for (takes precedence over email)")
}

func runUserDetails(db database.DatabaseInterface, log *logger.Logger, identifier string) (*user.User, error) {
	log.Info("Attempting to find user...", logger.Fields{"identifier": identifier})

	var foundUser *user.User
	var findErr error

	parsedID, parseErr := strconv.Atoi(identifier)

	if parseErr == nil && parsedID > 0 {
		log.Info("Interpreted identifier as ID", logger.Fields{"id": parsedID})
		foundUser, findErr = user.FindByID(db, parsedID)
	} else {
		log.Info("Interpreted identifier as Email", logger.Fields{"email": identifier})
		foundUser, findErr = user.FindByEmail(db, identifier)
	}

	if findErr != nil {
		if errors.Is(findErr, user.ErrInvalidCredentials) || strings.Contains(findErr.Error(), "not found") {
			log.Warn("User not found", logger.Fields{"identifier": identifier})
		} else {
			log.Error("Database error finding user", logger.Fields{
				"identifier": identifier,
				"error":      findErr.Error(),
			})
		}

		return nil, findErr
	}

	log.Info("Found user successfully", logger.Fields{"identifier": identifier, "userID": foundUser.GetID()})
	return foundUser, nil
}

func runUserDetailsCmd(cmd *cobra.Command, args []string) {
	log := logger.New(logger.Level(config.GetLogLevel()), os.Stdout)

	identifier := ""
	flagEmail, _ := cmd.Flags().GetString("email")
	flagID, _ := cmd.Flags().GetInt("id")

	if flagID > 0 {
		identifier = strconv.Itoa(flagID)
		log.Info("Using ID from flag", logger.Fields{"id": flagID})
	} else if flagEmail != "" {
		identifier = flagEmail
		log.Info("Using email from flag", logger.Fields{"email": flagEmail})
	} else {
		log.Info("No flags provided, prompting for input...", nil)
		input, err := promptForString("Enter user's email or ID: ")
		if err != nil {
			log.Error("Failed to get input", logger.Fields{"error": err.Error()})
			osExit(1)
			return
		}
		if input == "" {
			log.Error("Email or ID must be provided.", nil)
			osExit(1)
			return
		}
		identifier = input
		log.Info("Using identifier from prompt", logger.Fields{"identifier": identifier})
	}

	dbConfig := getDatabaseConfigForCmd()
	db, dbErr := database.New(dbConfig, log)

	if dbErr != nil {
		log.Error("Failed to connect to database", logger.Fields{"error": dbErr.Error()})
		osExit(1)

		return
	}

	defer func() { _ = db.Close() }()

	foundUser, runErr := runUserDetails(db, log, identifier)

	if runErr != nil {
		fmt.Fprintf(os.Stderr, "Error finding user: %v\n", runErr)
		osExit(1)

		return
	}

	fmt.Println("--- User Details ---")
	fmt.Printf("ID:        %d\n", foundUser.GetID())
	fmt.Printf("Username:  %s\n", foundUser.GetUsername())
	fmt.Printf("Email:     %s\n", foundUser.GetEmail())
	fmt.Printf("Status:    %s\n", strconv.FormatBool(foundUser.GetStatus()))
	fmt.Printf("Created:   %s\n", foundUser.GetCreatedAt().Format("2006-01-02 15:04:05"))
	fmt.Printf("Updated:   %s\n", foundUser.GetUpdatedAt().Format("2006-01-02 15:04:05"))
	fmt.Println("--------------------")
}
