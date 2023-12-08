package cmd

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/leancodepl/poe2arb/convert/poe2arb"
	"github.com/leancodepl/poe2arb/flutter"
	"github.com/leancodepl/poe2arb/log"
	"github.com/leancodepl/poe2arb/poeditor"
	"github.com/spf13/cobra"
)

var (
	poeCmd = &cobra.Command{
		Use: "poe",
		Short: "Exports POEditor terms and converts them to ARB. " +
			"Must be run from the Flutter project root directory or its subdirectory.",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE:          runPoe,
	}
	termPrefixRegexp = regexp.MustCompile("[a-zA-Z]*")
)

const (
	projectIDFlag     = "project-id"
	tokenFlag         = "token"
	termPrefixFlag    = "term-prefix"
	outputDirFlag     = "output-dir"
	overrideLangsFlag = "langs"
)

func init() {
	poeCmd.Flags().StringP(projectIDFlag, "p", "", "POEditor project ID")
	poeCmd.Flags().StringP(tokenFlag, "t", "", "POEditor API token")
	poeCmd.Flags().StringP(termPrefixFlag, "", "", "POEditor term prefix")
	poeCmd.Flags().StringP(outputDirFlag, "o", "", `Output directory [default: "."]`)
	poeCmd.Flags().StringSliceP(overrideLangsFlag, "", []string{}, "Override downloaded languages")
}

func runPoe(cmd *cobra.Command, args []string) error {
	log := getLogger(cmd)

	logSub := log.Info("loading options").Sub()

	sel, err := getOptionsSelector(cmd)
	if err != nil {
		logSub.Error("failed: " + err.Error())
		return err
	}

	options, err := sel.SelectOptions()
	if err != nil {
		logSub.Error("failed: " + err.Error())
		return err
	}

	poeCmd, err := NewPoeCommand(options, log)
	if err != nil {
		logSub.Error(err.Error())
		return err
	}

	log.Info("fetching project languages")
	langs, err := poeCmd.GetExportLanguages()
	if err != nil {
		logSub.Error(err.Error())
		return err
	}

	if err := poeCmd.EnsureOutputDirectory(); err != nil {
		return err
	}

	for _, lang := range langs {
		flutterLocale, err := flutter.ParseLocale(lang.Code)
		if err != nil {
			return fmt.Errorf("parsing %s language code: %w", lang.Code, err)
		}

		template := options.TemplateLocale == flutterLocale

		err = poeCmd.ExportLanguage(lang, flutterLocale, template)
		if err != nil {
			return fmt.Errorf("exporting %s (%s) language: %w", lang.Name, lang.Code, err)
		}
	}

	log.Success("done")

	return nil
}

func getOptionsSelector(cmd *cobra.Command) (*poeOptionsSelector, error) {
	envVars, err := newEnvVars()
	if err != nil {
		return nil, err
	}

	flutterCfg, err := getFlutterConfig()
	if err != nil {
		return nil, err
	}

	return &poeOptionsSelector{
		flags: cmd.Flags(),
		l10n:  flutterCfg.L10n,
		env:   envVars,
	}, nil
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

type poeCommand struct {
	options *poeOptions
	client  *poeditor.Client
	log     *log.Logger
}

func NewPoeCommand(options *poeOptions, log *log.Logger) (*poeCommand, error) {
	if errs := validatePoeOptions(options); len(errs) > 0 {
		msg := ""
		for _, err := range errs {
			msg += err.Error() + "\n"
		}
		return nil, errors.New(msg)
	}

	client := poeditor.NewClient(options.Token)

	return &poeCommand{
		options: options,
		client:  client,
		log:     log,
	}, nil
}

func validatePoeOptions(options *poeOptions) []error {
	errs := []error{}

	if options.ProjectID == "" {
		errs = append(errs, errors.New("no POEditor project id provided"))
	}

	if options.Token == "" {
		errs = append(errs, errors.New("no POEditor API token provided"))
	}

	if !termPrefixRegexp.MatchString(options.TermPrefix) {
		errs = append(errs, errors.New("term prefix must contain only letters or be empty"))
	}

	return errs
}

func (c *poeCommand) GetExportLanguages() ([]poeditor.Language, error) {
	langs, err := c.client.GetProjectLanguages(c.options.ProjectID)
	if err != nil {
		return nil, err
	}

	// Use only overriden langs
	if len(c.options.OverrideLangs) > 0 {
		var filteredLangs []poeditor.Language
		for _, lang := range langs {
			for _, overridenLang := range c.options.OverrideLangs {
				if strings.EqualFold(lang.Code, overridenLang) {
					filteredLangs = append(filteredLangs, lang)
					break
				}
			}
		}

		if len(filteredLangs) == 0 {
			langsWord := "lang"
			if len(overrideLangsFlag) > 1 {
				langsWord = "langs"
			}
			var available []string
			for _, lang := range langs {
				available = append(available, lang.Code)
			}
			return nil, fmt.Errorf(
				`--%s specified %d %s, but none of them were available in the POEditor project. Available langs: %s`,
				overrideLangsFlag, len(c.options.OverrideLangs), langsWord, strings.Join(available, ", "),
			)
		}

		return filteredLangs, nil
	}

	return langs, nil
}

func (c *poeCommand) EnsureOutputDirectory() error {
	dir := c.options.OutputDir
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		logSub := c.log.Info("creating directory %s", dir).Sub()
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			logSub.Error("failed: " + err.Error())
			return err
		}
	}

	return nil
}

func (c *poeCommand) ExportLanguage(lang poeditor.Language, flutterLocale flutter.Locale, template bool) error {
	logSub := c.log.Info("fetching JSON export for %s (%s)", lang.Name, lang.Code).Sub()
	url, err := c.client.GetExportURL(c.options.ProjectID, lang.Code)
	if err != nil {
		return err
	}

	resp, err := http.Get(url)
	if err != nil {
		logSub.Error("making HTTP request failed: " + err.Error())
		return fmt.Errorf("making HTTP request for export: %w", err)
	}

	filePath := path.Join(c.options.OutputDir, fmt.Sprintf("%s%s.arb", c.options.ARBPrefix, flutterLocale.StringFilename()))
	file, err := os.Create(filePath)
	if err != nil {
		logSub.Error("creating file failed: " + err.Error())
		return fmt.Errorf("creating ARB file: %w", err)
	}
	defer file.Close()

	convertLogSub := logSub.Info("converting JSON to ARB").Sub()

	conv := poe2arb.NewConverter(resp.Body, &poe2arb.ConverterOptions{
		Locale:                    flutterLocale,
		Template:                  template,
		RequireResourceAttributes: c.options.RequireResourceAttributes,
		TermPrefix:                c.options.TermPrefix,
	})
	err = conv.Convert(file)
	if err != nil {
		convertLogSub.Error(err.Error())
		return err
	}

	logSub.Success("saved to %s", filePath)

	return nil
}
