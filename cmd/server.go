package cmd

import (
	"fmt"
	"os"

	"github.com/Dobefu/go-web-starter/internal/config"
	"github.com/Dobefu/go-web-starter/internal/logger"
	"github.com/Dobefu/go-web-starter/internal/server"
	"github.com/spf13/cobra"
)

var osExit = os.Exit

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Run the website on a local server",
	Run:   ServerCmd,
}

func init() {
	rootCmd.AddCommand(serverCmd)

	serverCmd.Flags().IntP("port", "p", 4000, "The port to run the server on")
}

func ServerCmd(cmd *cobra.Command, args []string) {
	log := logger.New(config.GetLogLevel(), os.Stdout)

	port, err := cmd.Flags().GetInt("port")

	if err != nil {
		log.Error("Failed to get port flag", logger.Fields{"error": err.Error()})
		osExit(1)
	}

	log.Debug("Creating new server instance", logger.Fields{"port": port})
	srv, err := server.New(port)

	if err != nil {
		log.Error("Failed to initialize server", logger.Fields{"error": err.Error()})
		osExit(1)
	}

	if err := runServer(srv, port); err != nil {
		log.Error("Server runtime error", logger.Fields{"error": err.Error()})
		osExit(1)
	}
}

func runServer(srv server.ServerInterface, port int) error {
	log := logger.New(config.GetLogLevel(), os.Stdout)
	log.Info(fmt.Sprintf("Starting server on port %d", port), nil)

	return srv.Start()
}
