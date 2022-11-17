// Package cmd provides command-line interface commands.
package cmd

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
	Use:   "poe2arb",
	Short: "POEditor JSON to Flutter ARB converter",
}

func Execute() {
	rootCmd.AddCommand(convertCmd)
	rootCmd.AddCommand(poeCmd)
	rootCmd.AddCommand(versionCmd)

	rootCmd.Execute()
}
