package cmd

import (
	"fmt"

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
	port, _ := cmd.Flags().GetInt("port")
	srv := server.New(port)
	if err := runServer(srv, port); err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
	}
}

func runServer(srv server.ServerInterface, port int) error {
	fmt.Printf("Starting server on port %d...\n", port)
	return srv.Start()
}
