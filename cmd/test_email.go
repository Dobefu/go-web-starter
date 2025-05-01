package cmd

import (
	"os"

	"github.com/Dobefu/go-web-starter/internal/config"
	"github.com/Dobefu/go-web-starter/internal/email"
	"github.com/Dobefu/go-web-starter/internal/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var testEmailCmd = &cobra.Command{
	Use:   "test:email",
	Short: "Send a test email to verify that email is working",
	Run:   runTestEmailCmd,
}

func init() {
	rootCmd.AddCommand(testEmailCmd)

	testEmailCmd.Flags().StringP("email", "e", "", "Email address to send the email to")
}

func runTestEmailCmd(cmd *cobra.Command, args []string) {
	log := logger.New(logger.Level(config.GetLogLevel()), os.Stdout)

	recipient, err := cmd.Flags().GetString("email")

	if recipient == "" {
		recipient, err = promptForString("Enter Email: ")

		if err != nil {
			log.Error("Failed to get email input", logger.Fields{"error": err.Error()})
			return
		}

		if recipient == "" {
			log.Error("No email provided", nil)
			return
		}
	}

	emailClient := email.New(
		viper.GetString("email.host"),
		viper.GetString("email.port"),
		viper.GetString("email.identity"),
		viper.GetString("email.user"),
		viper.GetString("email.password"),
	)

	err = emailClient.SendMail("", []string{recipient}, "subject", "<h1>body</h1>")

	if err != nil {
		log.Error(err.Error(), nil)
		return
	}

	log.Info("Successfully sent the test email", nil)
}
