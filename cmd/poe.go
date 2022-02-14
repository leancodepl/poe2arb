package cmd

import (
	"fmt"
	"net/http"
	"os"
	"path"

	"github.com/leancodepl/poe2arb/converter"
	"github.com/leancodepl/poe2arb/flutter_config"
	"github.com/leancodepl/poe2arb/poeditor"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	poeCmd = &cobra.Command{
		Use:   "poe",
		Short: "Exports POEditor terms and converts them to ARB",
		RunE:  runPoe,
	}
)

const (
	projectIDFlag = "project-id"
	tokenFlag     = "token"
	arbPrefixFlag = "arb-prefix"
	outputDirFlag = "output-dir"
)

func init() {
	poeCmd.Flags().StringP(projectIDFlag, "p", "", "(required) POEditor project ID")
	poeCmd.MarkFlagRequired(projectIDFlag)

	poeCmd.Flags().StringP(tokenFlag, "t", "", "(required) POEditor API token")
	poeCmd.MarkFlagRequired(tokenFlag)

	poeCmd.Flags().StringP(arbPrefixFlag, "", "app_", "ARB file names prefix")

	poeCmd.Flags().StringP(outputDirFlag, "o", "", "Output directory. Defaults to current directory.")

	addElCompatFlag(poeCmd)
}

func runPoe(cmd *cobra.Command, args []string) error {
	flutterCfg, err := getFlutterConfig()
	if err != nil {
		return err
	}

	sel := poeOptionsSelector{cmd.Flags(), flutterCfg.L10n}
	projectID, err := sel.SelectProjectID()
	if err != nil {
		return err
	}
	token, err := sel.SelectToken()
	if err != nil {
		return err
	}
	arbPrefix, err := sel.SelectARBPrefix()
	if err != nil {
		return err
	}
	outputDir, err := sel.SelectOutputDir()
	if err != nil {
		return err
	}
	elCompat, err := sel.SelectElCompat()
	if err != nil {
		return err
	}

	client := poeditor.NewClient(token)

	fmt.Println("Fetching project languages...")
	langs, err := client.GetProjectLanguages(projectID)
	if err != nil {
		return err
	}

	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		fmt.Printf("Creating directory %s...\n", outputDir)
		if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
			return err
		}
	}

	for _, lang := range langs {
		fmt.Printf("Fetching JSON export for %s (%s)...\n", lang.Name, lang.Code)
		url, err := client.GetExportURL(projectID, lang.Code)
		if err != nil {
			return err
		}

		resp, err := http.Get(url)
		if err != nil {
			return err
		}

		filePath := path.Join(outputDir, fmt.Sprintf("%s%s.arb", arbPrefix, lang.Code))
		file, err := os.Create(filePath)
		if err != nil {
			return err
		}
		defer file.Close()

		conv := converter.NewConverter(elCompat)
		err = conv.Convert(resp.Body, file, lang.Code)
		if err != nil {
			return err
		}

		fmt.Printf("Success converting JSON to ARB for %s (%s).\n", lang.Name, lang.Code)
	}

	fmt.Println("\nDone!")

	return nil
}

func getFlutterConfig() (*flutter_config.FlutterConfig, error) {
	workDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	flutterCfg, err := flutter_config.NewFromDirectory(workDir)
	if err != nil {
		return nil, err
	}

	return flutterCfg, nil
}

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
