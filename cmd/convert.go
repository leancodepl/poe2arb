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

const (
	langFlag     = "lang"
	templateFlag = "template"
)

func init() {
	convertCmd.PersistentFlags().StringP(langFlag, "l", "", "Language of the input")
	convertCmd.MarkPersistentFlagRequired(langFlag)

	convertCmd.PersistentFlags().BoolP(templateFlag, "t", false, "Whether the output should be a template ARB")

	convertCmd.AddCommand(convertIoCmd)
}

func runConvertIo(cmd *cobra.Command, args []string) error {
	lang, _ := cmd.Flags().GetString(langFlag)
	template, _ := cmd.Flags().GetBool(templateFlag)

	conv := converter.NewConverter()

	return conv.Convert(os.Stdin, os.Stdout, lang, template)
}
