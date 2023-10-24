package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/leancodepl/poe2arb/convert/arb2poe"
	"github.com/leancodepl/poe2arb/poeditor"
	"github.com/spf13/cobra"
)

var seedCmd = &cobra.Command{
	Use:           "seed",
	Short:         "EXPERIMENTAL! Seeds POEditor with data from ARBs. To be used only on empty projects!",
	SilenceErrors: true,
	SilenceUsage:  true,
	RunE:          runSeed,
}

func init() {
	seedCmd.Flags().StringP(projectIDFlag, "p", "", "POEditor project ID")
	seedCmd.Flags().StringP(tokenFlag, "t", "", "POEditor API token")
	seedCmd.Flags().StringP(termPrefixFlag, "", "", "POEditor term prefix")
	seedCmd.Flags().StringP(outputDirFlag, "o", "", `Output directory [default: "."]`)
	seedCmd.Flags().StringSliceP(overrideLangsFlag, "", []string{}, "Override downloaded languages")
}

func runSeed(cmd *cobra.Command, args []string) error {
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

	logSub = log.Info("reading ARB files in %s", options.OutputDir).Sub()

	var files []string
	rawFiles, err := os.ReadDir(options.OutputDir)
	if err != nil {
		logSub.Error("failed: " + err.Error())
		return err
	}
	for _, file := range rawFiles {
		if file.IsDir() {
			continue
		}

		fileName := file.Name()
		if !strings.HasPrefix(fileName, options.ARBPrefix) || filepath.Ext(fileName) != ".arb" {
			continue
		}

		files = append(files, filepath.Join(options.OutputDir, fileName))
	}

	if len(files) == 0 {
		logSub.Error("no ARB files found")
		return err
	} else {
		logSub.Info("found %d ARB files", len(files))
	}

	poeClient := poeditor.NewClient(options.Token)

	availableLangs, err := poeClient.GetProjectLanguages(options.ProjectID)
	if err != nil {
		log.Error("failed fetching languages: " + err.Error())
		return err
	}

	first := true
	for _, filePath := range files {
		logSub = log.Info("seeding %s", filepath.Base(filePath)).Sub()
		logSub.Info("converting ARB to JSON")

		file, err := os.Open(filePath)
		if err != nil {
			logSub.Error("failed: " + err.Error())
			return err
		}

		converter := arb2poe.NewConverter(file, options.TemplateLocale, options.TermPrefix)

		var b bytes.Buffer
		lang, err := converter.Convert(&b)
		if err != nil {
			logSub.Error("failed: " + err.Error())
			return err
		}

		langFound := false
		for _, availableLang := range availableLangs {
			if lang == availableLang.Code {
				langFound = true
				break
			}
		}

		if !langFound {
			langSub := logSub.Info("adding language %s to project", lang).Sub()

			err = poeClient.AddLanguage(options.ProjectID, lang)
			if err != nil {
				langSub.Error("failed: " + err.Error())
				return err
			}
		}

		if !first {
			logSub.Info("waiting 30 seconds to avoid rate limiting")
			time.Sleep(30 * time.Second)
		}

		uploadSub := logSub.Info("uploading JSON to POEditor").Sub()

		err = poeClient.Upload(options.ProjectID, lang, &b)
		if err != nil {
			uploadSub.Error("failed: " + err.Error())
			return err
		} else {
			logSub.Success("done")
		}

		first = false
	}

	return nil
}
