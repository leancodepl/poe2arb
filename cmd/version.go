package cmd

import (
	"errors"
	"runtime/debug"

	"github.com/spf13/cobra"
)

// Version is the version of the application. It is set during the build process using ldflags.
var Version string

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of poe2arb",
	RunE:  runVersion,
}

func runVersion(cmd *cobra.Command, args []string) error {
	log := getLogger(cmd)

	revision, time, modified, err := getVcsInfo()
	if err != nil {
		return err
	}

	msg := "poe2arb"
	if Version != "" {
		msg += " version " + Version
	} else {
		msg += " built from source"
	}

	msg += ", commit " + revision

	if modified {
		msg += " (with local modifications)"
	}

	msg += ", built at " + time

	log.Info(msg)

	return nil
}

func getVcsInfo() (revision, time string, modified bool, err error) {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		err = errors.New("error reading build info")
		return
	}

	for _, setting := range info.Settings {
		if setting.Key == "vcs.revision" {
			revision = setting.Value
		} else if setting.Key == "vcs.time" {
			time = setting.Value
		} else if setting.Key == "vcs.modified" {
			modified = setting.Value == "true"
		}
	}

	return
}
