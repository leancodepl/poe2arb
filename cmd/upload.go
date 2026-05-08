package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/leancodepl/poe2arb/convert/arb2poe"
	"github.com/leancodepl/poe2arb/poeditor"
	"github.com/spf13/cobra"
)

const forceFlag = "force"

var uploadCmd = &cobra.Command{
	Use: "upload",
	Short: "Synchronizes POEditor with local ARB files, overwriting existing terms and translations. " +
		"Use this when you've made changes to your ARB files and want to push them to POEditor.",
	SilenceErrors: true,
	SilenceUsage:  true,
	RunE:          runUpload,
	PreRunE:       versionGuard.GetFlutterConfigAndEnsureSufficientVersion,
}

func init() {
	uploadCmd.Flags().StringP(projectIDFlag, "p", "", "POEditor project ID")
	uploadCmd.Flags().StringP(tokenFlag, "t", "", "POEditor API token")
	uploadCmd.Flags().StringP(termPrefixFlag, "", "", "POEditor term prefix")
	uploadCmd.Flags().StringP(outputDirFlag, "o", "", `Output directory [default: "."]`)
	uploadCmd.Flags().StringSliceP(overrideLangsFlag, "", []string{}, "Override uploaded languages")
	uploadCmd.Flags().Bool(forceFlag, false, "Allow deleting terms in POEditor that are no longer in local ARB files")
}

func runUpload(cmd *cobra.Command, _ []string) error {
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

	force, err := cmd.Flags().GetBool(forceFlag)
	if err != nil {
		return err
	}

	logSub = log.Info("reading ARB files in %s", options.OutputDir).Sub()

	files, templateFile, err := findARBFiles(options.OutputDir, options.ARBPrefix, options.TemplateLocale.StringFilename())
	if err != nil {
		logSub.Error("failed: " + err.Error())
		return err
	}

	if len(files) == 0 {
		err = errors.New("no ARB files found")
		logSub.Error(err.Error())
		return err
	}
	logSub.Info("found %d ARB files", len(files))

	if templateFile == "" {
		err = fmt.Errorf("template ARB file for locale %s not found", options.TemplateLocale)
		logSub.Error(err.Error())
		return err
	}

	poeClient := poeditor.NewClient(options.Token)

	log.Info("fetching project terms from POEditor")
	remoteTerms, err := poeClient.ListTerms(options.ProjectID, options.TemplateLocale.StringHyphen())
	if err != nil {
		log.Error("failed fetching terms: " + err.Error())
		return err
	}

	log.Info("computing terms diff")
	localTerms, err := readLocalTermNames(templateFile, options.TermPrefix)
	if err != nil {
		log.Error("failed: " + err.Error())
		return err
	}

	toDelete := termsToDelete(remoteTerms, localTerms, options.TermPrefix)

	if len(toDelete) > 0 {
		if !force {
			logSub := log.Error("the following %d term(s) exist in POEditor but not in local ARB files:", len(toDelete)).Sub()
			for _, t := range toDelete {
				if t.Translation != "" {
					logSub.Info("- %s = %q", t.Term, t.Translation)
				} else {
					logSub.Info("- %s", t.Term)
				}
			}
			log.Error("aborting. Re-run with --%s to delete these terms from POEditor and continue the upload.", forceFlag)
			return fmt.Errorf("%d term(s) would be deleted from POEditor; pass --%s to confirm", len(toDelete), forceFlag)
		}

		deleteSub := log.Info("deleting %d term(s) in POEditor", len(toDelete)).Sub()
		refs := make([]poeditor.TermRef, len(toDelete))
		for i, t := range toDelete {
			refs[i] = poeditor.TermRef{Term: t.Term, Context: t.Context}
			deleteSub.Info("- %s", t.Term)
		}
		if err := poeClient.DeleteTerms(options.ProjectID, refs); err != nil {
			deleteSub.Error("failed: " + err.Error())
			return err
		}
	}

	availableLangs, err := poeClient.GetProjectLanguages(options.ProjectID)
	if err != nil {
		log.Error("failed fetching languages: " + err.Error())
		return err
	}

	first := true
	freeAccountRateLimit := false
	// Upload the template language first so terms get their definitions.
	for _, filePath := range orderTemplateFirst(files, templateFile) {
		fileLog := log.Info("uploading %s", filepath.Base(filePath)).Sub()
		fileLog.Info("converting ARB to JSON")

		file, err := os.Open(filePath)
		if err != nil {
			fileLog.Error("failed: " + err.Error())
			return err
		}

		converter := arb2poe.NewConverter(file, options.TemplateLocale, options.TermPrefix)

		var b bytes.Buffer
		flutterLocale, err := converter.Convert(&b)
		_ = file.Close()
		if err != nil {
			if errors.Is(err, arb2poe.ErrNoTerms) {
				fileLog.Info("no terms to convert")
				continue
			}

			fileLog.Error("failed: " + err.Error())
			return err
		}
		lang := flutterLocale.StringHyphen()

		if len(options.OverrideLangs) > 0 {
			langFound := false
			for _, overridenLang := range options.OverrideLangs {
				if strings.EqualFold(lang, overridenLang) {
					langFound = true
					break
				}
			}

			if !langFound {
				fileLog.Info("skipping language %s", lang)
				continue
			}
		}

		availableLangFound := false
		for _, availableLang := range availableLangs {
			if strings.EqualFold(lang, availableLang.Code) {
				availableLangFound = true
				break
			}
		}

		if !availableLangFound {
			langLog := fileLog.Info("adding language %s to project", flutterLocale).Sub()

			err = poeClient.AddLanguage(options.ProjectID, lang)
			if err != nil {
				langLog.Error("failed: " + err.Error())
				return err
			}
		}

		if !first {
			rateLimitTimeout := poeditor.PaidAccountUploadRateLimit
			rateLimitName := "(paid account)"
			if freeAccountRateLimit {
				rateLimitTimeout = poeditor.FreeAccountUploadRateLimit
				rateLimitName = "(free account)"
			}

			fileLog.Info("waiting %v %s to avoid rate limiting", rateLimitTimeout, rateLimitName)
			time.Sleep(rateLimitTimeout)
		}

		uploadLog := fileLog.Info("uploading JSON to POEditor").Sub()

		uploadFileReader := bytes.NewReader(b.Bytes())
		for {
			err = poeClient.Upload(
				options.ProjectID, lang, uploadFileReader,
				poeditor.UploadOptions{Overwrite: true},
			)
			if err != nil {
				var poeErr *poeditor.Error
				if errors.As(err, &poeErr) && poeErr.Code == poeditor.RateLimitErrorCode && !freeAccountRateLimit {
					freeAccountRateLimit = true

					freeRateLimit := poeditor.FreeAccountUploadRateLimit
					uploadLog.Info("paid account rate limit was not enough, retrying with free account rate limit (%v)", freeRateLimit)
					uploadFileReader = bytes.NewReader(b.Bytes())
					time.Sleep(freeRateLimit)

					continue
				}

				uploadLog.Error("failed: " + err.Error())
				return err
			}

			fileLog.Success("done")
			break
		}

		first = false
	}

	log.Success("upload complete")

	return nil
}

// findARBFiles enumerates ARB files matching the prefix in a directory,
// and returns the path of the template file (matching templateLocaleFilename) if present.
func findARBFiles(dir, arbPrefix, templateLocaleFilename string) (files []string, templateFile string, err error) {
	rawFiles, err := os.ReadDir(dir)
	if err != nil {
		return nil, "", err
	}

	templateFilename := arbPrefix + templateLocaleFilename + ".arb"

	for _, file := range rawFiles {
		if file.IsDir() {
			continue
		}

		fileName := file.Name()
		if !strings.HasPrefix(fileName, arbPrefix) || filepath.Ext(fileName) != ".arb" {
			continue
		}

		path := filepath.Join(dir, fileName)
		files = append(files, path)
		if fileName == templateFilename {
			templateFile = path
		}
	}

	sort.Strings(files)
	return files, templateFile, nil
}

// orderTemplateFirst returns a copy of files with templateFile moved to the front.
func orderTemplateFirst(files []string, templateFile string) []string {
	out := make([]string, 0, len(files))
	if templateFile != "" {
		out = append(out, templateFile)
	}
	for _, f := range files {
		if f != templateFile {
			out = append(out, f)
		}
	}
	return out
}

// readLocalTermNames returns the set of POE term names (with prefix applied)
// derived from the template ARB file.
func readLocalTermNames(templateFile, termPrefix string) (map[string]struct{}, error) {
	f, err := os.Open(templateFile)
	if err != nil {
		return nil, fmt.Errorf("opening template ARB: %w", err)
	}
	defer f.Close()

	var arb map[string]any
	if err := json.NewDecoder(f).Decode(&arb); err != nil {
		return nil, fmt.Errorf("parsing template ARB: %w", err)
	}

	names := make(map[string]struct{})
	for key := range arb {
		if strings.HasPrefix(key, "@") {
			continue
		}

		full := key
		if termPrefix != "" {
			full = termPrefix + ":" + key
		}
		names[full] = struct{}{}
	}

	return names, nil
}

// termsToDelete returns POE terms that match the current prefix scope but are
// missing from the local ARB.
//
// Scope of the prefix:
//   - empty prefix → consider only POE terms WITHOUT a prefix.
//   - non-empty prefix → consider only POE terms with that exact prefix.
//
// This protects sibling packages sharing one POEditor project from accidental
// deletion.
func termsToDelete(remote []poeditor.Term, localPrefixed map[string]struct{}, termPrefix string) []poeditor.Term {
	var out []poeditor.Term
	for _, t := range remote {
		if !termInPrefixScope(t.Term, termPrefix) {
			continue
		}
		if _, ok := localPrefixed[t.Term]; ok {
			continue
		}
		out = append(out, t)
	}

	sort.Slice(out, func(i, j int) bool { return out[i].Term < out[j].Term })

	return out
}

// termInPrefixScope reports whether the given remote term name is in scope
// of the supplied term prefix (see termsToDelete).
func termInPrefixScope(term, termPrefix string) bool {
	colonIdx := strings.IndexByte(term, ':')
	if termPrefix == "" {
		return colonIdx == -1
	}

	if colonIdx == -1 {
		return false
	}

	return term[:colonIdx] == termPrefix
}
