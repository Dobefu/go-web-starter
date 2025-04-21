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

func runUserDetailsCmd(cmd *cobra.Command, args []string) {
	log := logger.New(config.GetLogLevel(), os.Stdout)

	flagEmail, _ := cmd.Flags().GetString("email")
	flagID, _ := cmd.Flags().GetInt("id")

	var foundUser *user.User

	lookupMethod := ""
	lookupValueStr := ""
	idToLookup := 0
	emailToLookup := ""

	if flagID > 0 {
		lookupMethod = "ID"
		idToLookup = flagID
		lookupValueStr = strconv.Itoa(flagID)
		log.Info("Attempting to find user by ID (from flag)...", logger.Fields{"id": idToLookup})

	} else if flagEmail != "" {
		lookupMethod = "Email"
		emailToLookup = flagEmail
		lookupValueStr = flagEmail
		log.Info("Attempting to find user by email (from flag)...", logger.Fields{"email": emailToLookup})

	} else {
		log.Info("No flags provided, prompting for input...", nil)
		input, err := promptForString("Enter user's email or ID: ")

		if err != nil {
			log.Error("Failed to get input", logger.Fields{"error": err.Error()})
			osExit(1)
		}

		if input == "" {
			log.Error("Email or ID must be provided.", nil)
			osExit(1)
		}

		parsedID, parseErr := strconv.Atoi(input)

		if parseErr == nil && parsedID > 0 {
			lookupMethod = "ID"
			idToLookup = parsedID
			lookupValueStr = input
			log.Info("Attempting to find user by ID (from prompt)...", logger.Fields{"id": idToLookup})
		} else {
			lookupMethod = "Email"
			emailToLookup = input
			lookupValueStr = input
			log.Info("Attempting to find user by email (from prompt)...", logger.Fields{"email": emailToLookup})
		}
	}

	dbConfig := getDatabaseConfigForCmd()
	db, dbErr := database.New(dbConfig, log)

	if dbErr != nil {
		log.Error("Failed to connect to database", logger.Fields{"error": dbErr.Error()})
		osExit(1)
	}

	defer func() { _ = db.Close() }()

	var findErr error

	if lookupMethod == "ID" {
		foundUser, findErr = user.FindByID(db, idToLookup)
	} else {
		foundUser, findErr = user.FindByEmail(db, emailToLookup)
	}

	if findErr != nil {
		if (lookupMethod == "Email" && errors.Is(findErr, user.ErrInvalidCredentials)) ||
			(lookupMethod == "ID" && strings.Contains(findErr.Error(), "not found")) {
			log.Error("User not found", logger.Fields{lookupMethod: lookupValueStr})
		} else {
			log.Error("Database error finding user", logger.Fields{
				"method": lookupMethod,
				"value":  lookupValueStr,
				"error":  findErr.Error(),
			})
		}

		osExit(1)
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
