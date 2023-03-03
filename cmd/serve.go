package cmd

import (
	"log"
	"net/http"

	"github.com/spf13/cobra"
	"twos.dev/pottytrainer/server"
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

		server := http.Server{
			Addr:    ":8080",
			Handler: server.RootHandler(cmd.Version, db),
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
