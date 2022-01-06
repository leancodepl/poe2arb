package cmd

import "github.com/spf13/cobra"

var (
	lang string

	rootCmd = &cobra.Command{
		Use:   "poe2arb",
		Short: "POEditor JSON to Flutter ARB converter",
	}
)

func Execute() {
	rootCmd.PersistentFlags().StringVarP(&lang, "lang", "l", "", "Language of the input file")
	rootCmd.MarkFlagRequired("lang")

	rootCmd.AddCommand(ioCmd)

	rootCmd.Execute()
}
