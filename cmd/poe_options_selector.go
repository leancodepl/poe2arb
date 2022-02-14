package cmd

import (
	"github.com/leancodepl/poe2arb/flutter_config"
	"github.com/spf13/pflag"
)

type poeOptionsSelector struct {
	flags *pflag.FlagSet
	l10n  *flutter_config.L10n
}

func (s *poeOptionsSelector) SelectProjectID() (string, error) {
	fromCmd, err := s.flags.GetString(projectIDFlag)
	return fromCmd, err
}

func (s *poeOptionsSelector) SelectToken() (string, error) {
	fromCmd, err := s.flags.GetString(tokenFlag)
	return fromCmd, err
}

func (s *poeOptionsSelector) SelectARBPrefix() (string, error) {
	fromCmd, err := s.flags.GetString(arbPrefixFlag)
	return fromCmd, err
}

func (s *poeOptionsSelector) SelectOutputDir() (string, error) {
	fromCmd, err := s.flags.GetString(outputDirFlag)
	if err != nil {
		return "", err
	}
	if fromCmd != "" {
		return fromCmd, nil
	}

	if s.l10n.ARBDir != "" {
		return s.l10n.ARBDir, nil
	}

	return ".", err
}

func (s *poeOptionsSelector) SelectElCompat() (bool, error) {
	return getElCompatFlag(s.flags), nil
}
