package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	elCompatFlag = "el-compat"
)

func addElCompatFlag(cmd *cobra.Command) {
	cmd.PersistentFlags().BoolP(elCompatFlag, "", false, "easy_localization compatibility mode")
}

func getElCompatFlag(flags *pflag.FlagSet) bool {
	elCompat, _ := flags.GetBool(elCompatFlag)
	return elCompat
}
