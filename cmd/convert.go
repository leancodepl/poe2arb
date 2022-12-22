package cmd

import (
	"os"

	"github.com/leancodepl/poe2arb/converter"
	"github.com/spf13/cobra"
)

var (
	convertCmd = &cobra.Command{
		Use:   "convert",
		Short: "Converts POEditor JSON to Flutter ARB",
	}

	convertIoCmd = &cobra.Command{
		Use:   "io",
		Short: "Converts from stdin to stdout",
		RunE:  runConvertIo,
	}
)

const langFlag = "lang"

func init() {
	convertCmd.PersistentFlags().StringP(langFlag, "l", "", "Language of the input file")
	convertCmd.MarkPersistentFlagRequired(langFlag)

	convertCmd.AddCommand(convertIoCmd)
}

func runConvertIo(cmd *cobra.Command, args []string) error {
	lang, _ := cmd.Flags().GetString(langFlag)

	conv := converter.NewConverter()

	return conv.Convert(os.Stdin, os.Stdout, lang)
}
