package cmd

import (
	"fmt"
	"os"

	"github.com/Dobefu/go-web-starter/internal/config"
	"github.com/Dobefu/go-web-starter/internal/logger"
	"github.com/Dobefu/go-web-starter/internal/server"
	"github.com/spf13/cobra"
)

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
		log.Error("Failed to get port", logger.Fields{"error": err.Error()})
		return
	}

	log.Debug("Creating new server instance", logger.Fields{"port": port})
	srv := server.New(port)
	err = runServer(srv, port)

	if err != nil {
		log.Error("Failed to start server", logger.Fields{"error": err.Error()})
	}
}

func runServer(srv server.ServerInterface, port int) error {
	log := logger.New(config.GetLogLevel(), os.Stdout)
	log.Info(fmt.Sprintf("Starting server on port %d", port), nil)

	return srv.Start()
}
