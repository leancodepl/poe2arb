package cmd

import "github.com/spf13/cobra"

const (
	elCompatFlag = "el-compat"
)

func addElCompatFlag(cmd *cobra.Command) {
	cmd.PersistentFlags().BoolP(elCompatFlag, "", false, "easy_localization compatibility mode")
}

func getElCompatFlag(cmd *cobra.Command) bool {
	elCompat, _ := cmd.Flags().GetBool(elCompatFlag)
	return elCompat
}
