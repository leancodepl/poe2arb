package cmd

import (
	"os"

	"github.com/leancodepl/poe2arb/converter"
	"github.com/spf13/cobra"
)

var ioCmd = &cobra.Command{
	Use:   "io",
	Short: "Reads JSON from stdin and output to stdout",
	Args:  cobra.ExactArgs(1),
	RunE:  runIo,
}

func runIo(cmd *cobra.Command, args []string) error {
	conv := converter.NewConverter(os.Stdin, os.Stdout, lang)
	return conv.Convert()
}
