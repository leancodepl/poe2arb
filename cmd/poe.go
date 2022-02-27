package cmd

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/leancodepl/poe2arb/converter"
	"github.com/leancodepl/poe2arb/flutter"
	"github.com/leancodepl/poe2arb/poeditor"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var poeCmd = &cobra.Command{
	Use:   "poe",
	Short: "Exports POEditor terms and converts them to ARB",
	RunE:  runPoe,
}

const (
	projectIDFlag     = "project-id"
	tokenFlag         = "token"
	arbPrefixFlag     = "arb-prefix"
	outputDirFlag     = "output-dir"
	overrideLangsFlag = "langs"
)

func init() {
	poeCmd.Flags().StringP(projectIDFlag, "p", "", "POEditor project ID")
	poeCmd.Flags().StringP(tokenFlag, "t", "", "POEditor API token")
	poeCmd.Flags().StringP(arbPrefixFlag, "", "app_", "ARB file names prefix")
	poeCmd.Flags().StringP(outputDirFlag, "o", "", `Output directory [default: "."]`)
	poeCmd.Flags().StringSliceP(overrideLangsFlag, "", []string{}, "Override downloaded languages")

	addElCompatFlag(poeCmd)
}

func runPoe(cmd *cobra.Command, args []string) error {
	envVars, err := newEnvVars()
	if err != nil {
		return err
	}

	flutterCfg, err := getFlutterConfig()
	if err != nil {
		return err
	}

	sel := poeOptionsSelector{cmd.Flags(), flutterCfg.L10n, envVars}

	projectID, err := sel.SelectProjectID()
	if err != nil {
		return err
	}
	if projectID == "" {
		return errors.New("no POEditor project id provided")
	}

	token, err := sel.SelectToken()
	if err != nil {
		return err
	}
	if token == "" {
		return errors.New("no POEditor token provided")
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
	overrideLangs, err := sel.SelectOverrideLangs()
	if err != nil {
		return err
	}

	client := poeditor.NewClient(token)

	fmt.Println("Fetching project languages...")
	langs, err := client.GetProjectLanguages(projectID)
	if err != nil {
		return err
	}

	// Use only overriden langs
	if len(overrideLangs) > 0 {
		var tmp []poeditor.Language
		for _, lang := range langs {
			for _, overridenLang := range overrideLangs {
				if lang.Code == overridenLang {
					tmp = append(tmp, lang)
					break
				}
			}
		}

		if len(tmp) == 0 {
			langsWord := "lang"
			if len(overrideLangsFlag) > 1 {
				langsWord = "langs"
			}
			var available []string
			for _, lang := range langs {
				available = append(available, lang.Code)
			}
			return fmt.Errorf(
				`--%s specified %d %s, but none of them were available in the POEditor project. Available langs: %s`,
				overrideLangsFlag, len(overrideLangs), langsWord, strings.Join(available, ", "),
			)
		}

		langs = tmp
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
			return errors.Wrap(err, "making HTTP request for export")
		}

		filePath := path.Join(outputDir, fmt.Sprintf("%s%s.arb", arbPrefix, lang.Code))
		file, err := os.Create(filePath)
		if err != nil {
			return errors.Wrap(err, "creating ARB file")
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
