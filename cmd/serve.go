package cmd

import (
	"log"
	"net/http"

	"github.com/spf13/cobra"
	"twos.dev/pottytrainer/server"
)

const (
	// AppleClientID is the environment variable key for the Apple client ID.
	keyAppleClientID = "APPLE_CLIENT_ID"
	// AppleClientSecret is the environment variable key for the Apple client
	// secret.
	keyAppleClientSecret = "APPLE_CLIENT_SECRET"
)

// serveCmd represents the serve command.
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: wrap(`Start the API server`),
	Long: wrap(`
    Start the Potty Trainer API server to accept food and bowel movement logs.
  `),
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := server.NewDB(&server.DBConfig{
			Region:          "us-east-1",
			TableNamePrefix: "pottytrainer",
		})
		if err != nil {
			return err
		}

		appleClient, err := server.NewAppleClient()
		if err != nil {
			return err
		}

		server := http.Server{
			Addr:    ":8080",
			Handler: server.RootHandler(cmd.Version, db, appleClient),
		}

		log.Print("Starting server on :8080")
		if err := server.ListenAndServe(); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}
