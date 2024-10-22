package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/hashicorp/go-version"
	"github.com/leancodepl/poe2arb/flutter"
	"github.com/spf13/cobra"
)

type flutterConfigKey int

const key flutterConfigKey = 1

func contextWithFlutterConfig(ctx context.Context, flutterConfig *flutter.FlutterConfig) context.Context {
	return context.WithValue(ctx, key, flutterConfig)
}

func flutterConfigFromContext(ctx context.Context) (*flutter.FlutterConfig, bool) {
	flutterCfg, ok := ctx.Value(key).(*flutter.FlutterConfig)
	return flutterCfg, ok
}

func ensureSufficientVersion(versionConstraint string) error {
	if versionConstraint == "" {
		return nil
	}

	constraint, err := newConstraintFromString(versionConstraint)
	if err != nil {
		return fmt.Errorf("invalid poe2arb-version format in l10n.yaml: %s", versionConstraint)
	}

	version, _ := version.NewVersion(Version)

	if !constraint.Check(version) {
		return fmt.Errorf("Poe2Arb version %s does not match constraint %s defined in l10n.yaml", version, versionConstraint)
	}

	return nil
}

func newConstraintFromString(versionConstraint string) (version.Constraints, error) {
	return version.NewConstraint(strings.ReplaceAll(versionConstraint, " ", ", "))
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

func getFlutterConfigAndEnsureSufficientVersion(cmd *cobra.Command, _ []string) {
	log := getLogger(cmd)

	logSub := log.Info("loading Flutter config").Sub()

	flutterCfg, err := getFlutterConfig()
	if err != nil {
		logSub.Error("failed: " + err.Error())
		os.Exit(1)
	}

	err = ensureSufficientVersion(flutterCfg.L10n.Poe2ArbVersion)
	if err != nil {
		logSub.Error("failed: " + err.Error())
		os.Exit(1)
	}

	ctx := contextWithFlutterConfig(cmd.Context(), flutterCfg)
	cmd.SetContext(ctx)
}
