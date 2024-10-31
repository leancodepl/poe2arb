// Package cmd provides command-line interface commands.
package cmd

import (
	"context"
	"os"

	"github.com/leancodepl/poe2arb/log"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "poe2arb",
	Short: "POEditor JSON to Flutter ARB converter",
}

type loggerKey struct{}

func Execute(logger *log.Logger) {
	rootCmd.AddCommand(convertCmd)
	rootCmd.AddCommand(poeCmd)
	rootCmd.AddCommand(seedCmd)
	rootCmd.AddCommand(versionCmd)

	ctx := context.WithValue(context.Background(), loggerKey{}, logger)

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
}

func getLogger(cmd *cobra.Command) *log.Logger {
	return cmd.Context().Value(loggerKey{}).(*log.Logger)
}
