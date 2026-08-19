package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/leancodepl/poe2arb/convert"
	"github.com/leancodepl/poe2arb/convert/arb2poe"
	"github.com/leancodepl/poe2arb/flutter"
	"github.com/leancodepl/poe2arb/log"
	"github.com/leancodepl/poe2arb/poeditor"
	"github.com/spf13/cobra"
)

const (
	forceFlag   = "force"
	dryRunFlag  = "dry-run"
	maxListItem = 20 // truncate long term lists in output
)

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
	uploadCmd.Flags().Bool(dryRunFlag, false, "Show what would change without modifying POEditor")
}

func runUpload(cmd *cobra.Command, _ []string) error {
	logger := getLogger(cmd)

	logSub := logger.Info("loading options").Sub()

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

	dryRun, err := cmd.Flags().GetBool(dryRunFlag)
	if err != nil {
		return err
	}

	logSub = logger.Info("reading ARB files in %s", options.OutputDir).Sub()

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

	parsedFiles, err := parseAllARBs(orderTemplateFirst(files, templateFile), options.TemplateLocale, options.TermPrefix)
	if err != nil {
		logger.Error("failed parsing ARB files: " + err.Error())
		return err
	}

	poeClient := poeditor.NewClient(options.Token)

	logger.Info("fetching project terms from POEditor")
	remoteTerms, err := poeClient.ListTerms(options.ProjectID, options.TemplateLocale.StringHyphen())
	if err != nil {
		logger.Error("failed fetching terms: " + err.Error())
		return err
	}

	availableLangs, err := poeClient.GetProjectLanguages(options.ProjectID)
	if err != nil {
		logger.Error("failed fetching languages: " + err.Error())
		return err
	}

	localTermSet := localTermNames(parsedFiles, templateFile, options.TermPrefix)
	toDelete := termsToDelete(remoteTerms, localTermSet, options.TermPrefix)

	if dryRun {
		return runUploadDryRun(logger, poeClient, options, parsedFiles, templateFile, remoteTerms, availableLangs, toDelete)
	}

	if len(toDelete) > 0 {
		if !force {
			subLog := logger.Error("the following %d term(s) exist in POEditor but not in local ARB files:", len(toDelete)).Sub()
			for _, t := range toDelete {
				if t.Translation != "" {
					subLog.Info("- %s = %q", t.Term, t.Translation)
				} else {
					subLog.Info("- %s", t.Term)
				}
			}
			logger.Error("aborting. Re-run with --%s to delete these terms from POEditor and continue the upload, or --%s to preview without changes.", forceFlag, dryRunFlag)
			return fmt.Errorf("%d term(s) would be deleted from POEditor; pass --%s to confirm", len(toDelete), forceFlag)
		}

		deleteSub := logger.Info("deleting %d term(s) in POEditor", len(toDelete)).Sub()
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

	first := true
	freeAccountRateLimit := false
	for _, pf := range parsedFiles {
		fileLog := logger.Info("uploading %s", filepath.Base(pf.Path)).Sub()
		lang := pf.Locale.StringHyphen()

		if shouldSkipLang(lang, options.OverrideLangs) {
			fileLog.Info("skipping language %s", lang)
			continue
		}

		if !langInProject(lang, availableLangs) {
			langLog := fileLog.Info("adding language %s to project", pf.Locale).Sub()
			if err := poeClient.AddLanguage(options.ProjectID, lang); err != nil {
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

		uploadFileReader := bytes.NewReader(pf.JSONBytes)
		for {
			err := poeClient.Upload(
				options.ProjectID, lang, uploadFileReader,
				poeditor.UploadOptions{Overwrite: true},
			)
			if err != nil {
				var poeErr *poeditor.Error
				if errors.As(err, &poeErr) && poeErr.Code == poeditor.RateLimitErrorCode && !freeAccountRateLimit {
					freeAccountRateLimit = true

					freeRateLimit := poeditor.FreeAccountUploadRateLimit
					uploadLog.Info("paid account rate limit was not enough, retrying with free account rate limit (%v)", freeRateLimit)
					uploadFileReader = bytes.NewReader(pf.JSONBytes)
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

	logger.Success("upload complete")
	return nil
}

func runUploadDryRun(
	logger *log.Logger,
	poeClient *poeditor.Client,
	options *poeOptions,
	parsedFiles []*parsedARB,
	templateFile string,
	remoteTerms []poeditor.Term,
	availableLangs []poeditor.Language,
	toDelete []poeditor.Term,
) error {
	logger.Info("DRY RUN — no changes will be made")

	templatePF := findTemplateFile(parsedFiles, templateFile)
	localTermSet := localTermNames(parsedFiles, templateFile, options.TermPrefix)
	toAdd := termsToAdd(remoteTerms, localTermSet, templatePF, options.TermPrefix)

	logger.Info("term additions: %d", len(toAdd))
	if len(toAdd) > 0 {
		printTermList(logger.Sub(), toAdd, "+")
	}

	logger.Info("term deletions: %d", len(toDelete))
	if len(toDelete) > 0 {
		sub := logger.Sub()
		for i, t := range toDelete {
			if i >= maxListItem {
				sub.Info("... and %d more", len(toDelete)-maxListItem)
				break
			}
			if t.Translation != "" {
				sub.Info("- %s = %q", t.Term, t.Translation)
			} else {
				sub.Info("- %s", t.Term)
			}
		}
		sub.Info("(--%s required to actually delete these)", forceFlag)
	}

	logger.Info("per-language plan:")
	langSub := logger.Sub()

	for _, pf := range parsedFiles {
		lang := pf.Locale.StringHyphen()

		if shouldSkipLang(lang, options.OverrideLangs) {
			langSub.Info("- %s: skipped (not in --%s)", lang, overrideLangsFlag)
			continue
		}

		exists := langInProject(lang, availableLangs)

		if !exists {
			langSub.Info("- %s: language MISSING in POEditor — would be created and uploaded (%d terms)", lang, len(pf.Terms))
			continue
		}

		summary, err := diffLanguage(poeClient, options.ProjectID, lang, pf.Terms)
		if err != nil {
			langSub.Error("- %s: failed to fetch translations: %s", lang, err.Error())
			return err
		}

		if summary.changed() == 0 {
			langSub.Info("- %s: no changes (%d terms unchanged)", lang, summary.Unchanged)
			continue
		}

		langSub.Info("- %s: %d added, %d updated, %d unchanged", lang, len(summary.Added), len(summary.Updated), summary.Unchanged)
		detailSub := langSub.Sub()
		printTermList(detailSub, summary.Added, "+")
		printTermList(detailSub, summary.Updated, "~")
	}

	logger.Info("dry run complete")
	return nil
}

// parsedARB is a local ARB file that has been parsed and converted to POE
// terms ready for upload.
type parsedARB struct {
	Path      string
	Locale    flutter.Locale
	Terms     []*convert.POETerm
	JSONBytes []byte
}

// parseAllARBs parses each ARB file and converts it to POE terms in memory.
// Files that produce no terms (e.g. empty translations file with non-template
// locale and term prefix filtering) are skipped silently.
func parseAllARBs(files []string, templateLocale flutter.Locale, termPrefix string) ([]*parsedARB, error) {
	var out []*parsedARB
	for _, p := range files {
		f, err := os.Open(p)
		if err != nil {
			return nil, fmt.Errorf("opening %s: %w", p, err)
		}

		var b bytes.Buffer
		converter := arb2poe.NewConverter(f, templateLocale, termPrefix)
		locale, err := converter.Convert(&b)
		_ = f.Close()
		if err != nil {
			if errors.Is(err, arb2poe.ErrNoTerms) {
				continue
			}
			return nil, fmt.Errorf("converting %s: %w", p, err)
		}

		jsonBytes := append([]byte(nil), b.Bytes()...)

		var terms []*convert.POETerm
		if err := json.Unmarshal(jsonBytes, &terms); err != nil {
			return nil, fmt.Errorf("re-parsing converted %s: %w", p, err)
		}

		out = append(out, &parsedARB{
			Path:      p,
			Locale:    locale,
			Terms:     terms,
			JSONBytes: jsonBytes,
		})
	}

	return out, nil
}

func findTemplateFile(parsed []*parsedARB, templatePath string) *parsedARB {
	for _, pf := range parsed {
		if pf.Path == templatePath {
			return pf
		}
	}
	return nil
}

// localTermNames returns the set of POE term names (with prefix applied)
// from the template ARB file (which is the canonical source of all term names).
func localTermNames(parsed []*parsedARB, templatePath, termPrefix string) map[string]struct{} {
	_ = termPrefix // already applied by the converter
	template := findTemplateFile(parsed, templatePath)
	names := map[string]struct{}{}
	if template == nil {
		return names
	}
	for _, t := range template.Terms {
		names[t.Term] = struct{}{}
	}
	return names
}

// readLocalTermNames is preserved for tests that exercise the template ARB
// parsing path directly without going through the converter.
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

// termsToAdd returns names of terms that exist locally (in the template) but
// not in POEditor, sorted alphabetically.
func termsToAdd(remote []poeditor.Term, local map[string]struct{}, _ *parsedARB, _ string) []string {
	remoteSet := map[string]struct{}{}
	for _, t := range remote {
		remoteSet[t.Term] = struct{}{}
	}

	var out []string
	for name := range local {
		if _, ok := remoteSet[name]; !ok {
			out = append(out, name)
		}
	}
	sort.Strings(out)
	return out
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

// langDiffSummary reports per-language differences between local and remote
// translations of the terms in a single ARB file.
type langDiffSummary struct {
	Added     []string // term in local, missing/empty in remote
	Updated   []string // term exists in both, translation differs
	Unchanged int
}

func (s langDiffSummary) changed() int {
	return len(s.Added) + len(s.Updated)
}

// diffLanguage compares the local POE terms for a language against POEditor's
// current state for that language. It performs one /terms/list call.
func diffLanguage(client *poeditor.Client, projectID, lang string, local []*convert.POETerm) (langDiffSummary, error) {
	remote, err := client.ListTerms(projectID, lang)
	if err != nil {
		return langDiffSummary{}, err
	}

	remoteByName := map[string]poeditor.Term{}
	for _, t := range remote {
		remoteByName[t.Term] = t
	}

	var s langDiffSummary
	for _, l := range local {
		r, ok := remoteByName[l.Term]
		if !ok || len(r.TranslationRaw) == 0 {
			s.Added = append(s.Added, l.Term)
			continue
		}

		if equalTranslation(l.Definition, r.TranslationRaw) {
			s.Unchanged++
		} else {
			s.Updated = append(s.Updated, l.Term)
		}
	}

	sort.Strings(s.Added)
	sort.Strings(s.Updated)
	return s, nil
}

// equalTranslation reports whether the local POE definition is structurally
// identical to the remote translation JSON.
func equalTranslation(local convert.POETermDefinition, remoteRaw []byte) bool {
	var remote convert.POETermDefinition
	if err := json.Unmarshal(remoteRaw, &remote); err != nil {
		return false
	}

	if local.IsPlural != remote.IsPlural {
		return false
	}

	if local.IsPlural {
		return reflect.DeepEqual(local.Plural, remote.Plural)
	}

	return strPtrEqual(local.Value, remote.Value)
}

func strPtrEqual(a, b *string) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		// Treat nil and "" as equivalent — POEditor sometimes returns one or
		// the other for blank translations.
		return derefOrEmpty(a) == derefOrEmpty(b)
	}
	return *a == *b
}

func derefOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func shouldSkipLang(lang string, overrides []string) bool {
	if len(overrides) == 0 {
		return false
	}
	for _, o := range overrides {
		if strings.EqualFold(lang, o) {
			return false
		}
	}
	return true
}

func langInProject(lang string, available []poeditor.Language) bool {
	for _, l := range available {
		if strings.EqualFold(lang, l.Code) {
			return true
		}
	}
	return false
}

func printTermList(logger *log.Logger, names []string, prefix string) {
	for i, name := range names {
		if i >= maxListItem {
			logger.Info("... and %d more", len(names)-maxListItem)
			return
		}
		logger.Info("%s %s", prefix, name)
	}
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
