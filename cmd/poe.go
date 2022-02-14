package cmd

import (
	"fmt"
	"net/http"
	"os"
	"path"

	"github.com/leancodepl/poe2arb/converter"
	"github.com/leancodepl/poe2arb/poeditor"
	"github.com/spf13/cobra"
)

var poeCmd = &cobra.Command{
	Use:   "poe",
	Short: "Exports POEditor terms and converts them to ARB",
	RunE:  runPoe,
}

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

	poeCmd.Flags().StringP(outputDirFlag, "o", ".", "Output directory")

	addElCompatFlag(poeCmd)
}

func runPoe(cmd *cobra.Command, args []string) error {
	projectID, _ := cmd.Flags().GetString(projectIDFlag)
	token, _ := cmd.Flags().GetString(tokenFlag)
	arbPrefix, _ := cmd.Flags().GetString(arbPrefixFlag)
	outputDir, _ := cmd.Flags().GetString(outputDirFlag)
	elCompat := getElCompatFlag(cmd)

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
