package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/go-version"
	"github.com/leancodepl/poe2arb/flutter"
	"github.com/spf13/cobra"
)

const flutterConfigKey = ctxKey(2)

func flutterConfigFromCommand(cmd *cobra.Command) *flutter.FlutterConfig {
	return cmd.Context().Value(flutterConfigKey).(*flutter.FlutterConfig)
}

// getFlutterConfigAndEnsureSufficientVersion gets Flutter project configuration,
// puts it in the command's context and verifies if poe2arb version matches constraint.
func getFlutterConfigAndEnsureSufficientVersion(cmd *cobra.Command, _ []string) error {
	log := getLogger(cmd)

	logSub := log.Info("loading Flutter config").Sub()

	flutterCfg, err := getFlutterConfig()
	if err != nil {
		logSub.Error("failed: " + err.Error())
		return err
	}

	err = ensureSufficientVersion(flutterCfg.L10n.Poe2ArbVersion)
	if err != nil {
		logSub.Error("failed: " + err.Error())
		return err
	}

	ctx := context.WithValue(cmd.Context(), flutterConfigKey, flutterCfg)
	cmd.SetContext(ctx)

	return nil
}

func ensureSufficientVersion(versionConstraint string) error {
	if versionConstraint == "" {
		return nil
	}

	constraint, err := version.NewConstraint(versionConstraint)
	if err != nil {
		return fmt.Errorf("invalid poe2arb-version format in l10n.yaml: %s", versionConstraint)
	}

	version, err := version.NewVersion(Version)
	if err != nil {
		return fmt.Errorf("poe2arb version format is invalid: %w", err)
	}

	if !constraint.Check(version) {
		return fmt.Errorf("poe2arb version %s does not match constraint %s defined in l10n.yaml", version, versionConstraint)
	}

	return nil
}

func getFlutterConfig() (*flutter.FlutterConfig, error) {
	workDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	flutterCfg, err := flutter.NewFromDirectory(workDir)
	if err != nil {
		return nil, err
	}

	return flutterCfg, nil
}
