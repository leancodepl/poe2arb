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

var (
	poeCmd = &cobra.Command{
		Use:  "poe",
		RunE: runPoe,
	}
)

func init() {
	poeCmd.Flags().StringP("token", "t", "", "POEditor API token")
	poeCmd.MarkFlagRequired("token")

	poeCmd.Flags().StringP("project-id", "p", "", "POEditor project ID")
	poeCmd.MarkFlagRequired("project-id")

	poeCmd.Flags().StringP("arb-prefix", "", "app_", "ARB file names prefix")

	poeCmd.Flags().StringP("output-dir", "o", ".", "Output directory")
}

func runPoe(cmd *cobra.Command, args []string) error {
	projectID, _ := cmd.Flags().GetString("project-id")
	token, _ := cmd.Flags().GetString("token")
	arbPrefix, _ := cmd.Flags().GetString("arb-prefix")
	outputDir, _ := cmd.Flags().GetString("output-dir")

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

		err = converter.NewConverter(resp.Body, file, lang.Code).Convert()
		if err != nil {
			return err
		}

		fmt.Printf("Success converting JSON to ARB for %s (%s).\n", lang.Name, lang.Code)
	}

	fmt.Println("\nDone!")

	return nil
}
