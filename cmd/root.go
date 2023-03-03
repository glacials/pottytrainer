package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "pottytrainer",
	Short: `Discover which foods cause poor bowel movements`,
	Long: wrap(`
    Potty Trainer is a tool to maintain a food journal and bowel movement
    journal, and to find correlations between the two.
`),
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(version string) {
	rootCmd.Version = version
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
