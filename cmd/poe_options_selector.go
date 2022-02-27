package cmd

import (
	"path/filepath"
	"strings"

	"github.com/leancodepl/poe2arb/flutter"
	"github.com/pkg/errors"
	"github.com/spf13/pflag"
	"golang.org/x/text/language"
)

// poeOptionsSelector decides on the correct values for given options
// depending on the available sources.
type poeOptionsSelector struct {
	flags *pflag.FlagSet
	l10n  *flutter.L10n
	env   *envVars
}

// poeOptions describes options passed or otherwise obtained to the poe command.
type poeOptions struct {
	ProjectID     string
	Token         string
	ARBPrefix     string
	OutputDir     string
	ElCompat      bool
	OverrideLangs []string
}

// SelectOptions selects all the options used for the poe command.
func (s *poeOptionsSelector) SelectOptions() (*poeOptions, error) {
	projectID, err := s.SelectProjectID()
	if err != nil {
		return nil, err
	}

	token, err := s.SelectToken()
	if err != nil {
		return nil, err
	}

	arbPrefix, err := s.SelectARBPrefix()
	if err != nil {
		return nil, err
	}

	outputDir, err := s.SelectOutputDir()
	if err != nil {
		return nil, err
	}

	elCompat, err := s.SelectElCompat()
	if err != nil {
		return nil, err
	}

	overrideLangs, err := s.SelectOverrideLangs()
	if err != nil {
		return nil, err
	}

	return &poeOptions{
		ProjectID:     projectID,
		Token:         token,
		ARBPrefix:     arbPrefix,
		OutputDir:     outputDir,
		ElCompat:      elCompat,
		OverrideLangs: overrideLangs,
	}, nil
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
	return prefixFromTemplateFileName(s.l10n.TemplateArbFile)
}

// see Flutter gen-l10n implementation:
// https://github.com/flutter/flutter/blob/61a0add2865c51bfee33939c1820709d1115c77d/packages/flutter_tools/lib/src/localizations/gen_l10n_types.dart#L454-L460

func prefixFromTemplateFileName(templateFile string) (string, error) {
	filename := strings.TrimSuffix(templateFile, filepath.Ext(templateFile))

	for i := 0; i < len(filename)-1; i++ {
		if filename[i] != '_' {
			continue
		}

		locale := filename[i+1:]
		_, err := language.Parse(locale)
		if err == nil {
			return filename[:i+1], nil
		}
	}

	return "", errors.New(
		"invalid template-arb-file. Should be a filename with prefix ending " +
			"with an underscore followed by a valid BCP-47 locale.",
	)
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
