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
	langFlag       = "lang"
	noTemplateFlag = "no-template"
)

func init() {
	convertCmd.PersistentFlags().StringP(langFlag, "l", "", "Language of the input")
	convertCmd.MarkPersistentFlagRequired(langFlag)

	convertCmd.PersistentFlags().StringP(termPrefixFlag, "t", "", "POEditor term prefix")
	convertCmd.PersistentFlags().Bool(noTemplateFlag, false, "Whether the output should NOT be generated as a template ARB")

	convertCmd.AddCommand(convertIoCmd)
}

func runConvertIo(cmd *cobra.Command, args []string) error {
	lang, _ := cmd.Flags().GetString(langFlag)
	noTemplate, _ := cmd.Flags().GetBool(noTemplateFlag)
	termPrefix, _ := cmd.Flags().GetString(termPrefixFlag)

	conv := converter.NewConverter(os.Stdin, &converter.ConverterOptions{
		Lang:                      lang,
		Template:                  !noTemplate,
		RequireResourceAttributes: true,
		TermPrefix:                termPrefix,
	})

	return conv.Convert(os.Stdout)
}
