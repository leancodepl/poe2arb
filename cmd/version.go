package cmd

import "github.com/spf13/cobra"

var (
	// Version is the version of the application. It is set during the build process using ldflags.
	Version = "dev"
	// Commit is the commit hash of the application. It is set during the build process using ldflags.
	Commit = "none"
	// Date is the date of the build. It is set during the build process using ldflags.
	BuiltDate = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of poe2arb",
	RunE:  runVersion,
}

func runVersion(cmd *cobra.Command, args []string) error {
	log := GetLogger(cmd)

	log.Info("poe2arb version %s, commit %s, built at %s", Version, Commit, BuiltDate)

	return nil
}
