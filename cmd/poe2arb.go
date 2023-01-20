// Package cmd provides command-line interface commands.
package cmd

import (
	"context"

	"github.com/leancodepl/poe2arb/log"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "poe2arb",
	Short: "POEditor JSON to Flutter ARB converter",
}

type ctxKey int

const loggerKey = ctxKey(1)

func Execute(logger *log.Logger) {
	rootCmd.AddCommand(convertCmd)
	rootCmd.AddCommand(poeCmd)
	rootCmd.AddCommand(versionCmd)

	ctx := context.WithValue(context.Background(), loggerKey, logger)

	rootCmd.ExecuteContext(ctx)
}

func getLogger(cmd *cobra.Command) *log.Logger {
	return cmd.Context().Value(loggerKey).(*log.Logger)
}
