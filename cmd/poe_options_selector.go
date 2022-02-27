package cmd

import (
	"github.com/leancodepl/poe2arb/flutter"
	"github.com/spf13/pflag"
)

// poeOptionsSelector decides on the correct values for given options
// depending on the available sources.
type poeOptionsSelector struct {
	flags *pflag.FlagSet
	l10n  *flutter.L10n
	env   *envVars
}

// SelectProjectID returns POEditor project id from available sources.
func (s *poeOptionsSelector) SelectProjectID() (string, error) {
	fromCmd, err := s.flags.GetString(projectIDFlag)
	if err != nil {
		return "", err
	}
	if fromCmd != "" {
		return fromCmd, nil
	}

	return s.l10n.POEditorProjectID, nil
}

// SelectToken returns POEditor API token option from available sources.
func (s *poeOptionsSelector) SelectToken() (string, error) {
	fromCmd, err := s.flags.GetString(tokenFlag)
	if err != nil {
		return "", err
	}
	if fromCmd != "" {
		return fromCmd, nil
	}

	return s.env.Token, nil
}

// SelectARBPrefix returns ARB files prefix option from available sources.
func (s *poeOptionsSelector) SelectARBPrefix() (string, error) {
	fromCmd, err := s.flags.GetString(arbPrefixFlag)
	return fromCmd, err
}

// SelectOutputDir returns output directory option from available sources.
//
// Defaults to current directory.
func (s *poeOptionsSelector) SelectOutputDir() (string, error) {
	fromCmd, err := s.flags.GetString(outputDirFlag)
	if err != nil {
		return "", err
	}
	if fromCmd != "" {
		return fromCmd, nil
	}

	if s.l10n != nil && s.l10n.ARBDir != "" {
		return s.l10n.ARBDir, nil
	}

	return ".", err
}

// SelectElCompat returns whether the easy_localizations compatibility option
// is enabled from available sources.
//
// Defaults to false.
func (s *poeOptionsSelector) SelectElCompat() (bool, error) {
	return getElCompatFlag(s.flags), nil
}

// SelectOverrideLangs returns a slice of languages that narrow down
// the available languages from POEditor API.
//
// Defaults to empty, which doesn't change the original language list.
func (s *poeOptionsSelector) SelectOverrideLangs() ([]string, error) {
	fromCmd, err := s.flags.GetStringSlice(overrideLangsFlag)
	if err != nil {
		return nil, err
	}
	if len(fromCmd) > 0 {
		return fromCmd, nil
	}

	fromL10n := s.l10n.POEditorLangs
	if fromL10n != nil {
		return fromL10n, nil
	}

	return []string{}, nil
}
