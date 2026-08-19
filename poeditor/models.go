package poeditor

import "encoding/json"

type Language struct {
	Name string
	Code string
}

type Term struct {
	Term      string
	Context   string
	Plural    string
	Reference string
	Tags      []string
	Comment   string
	// Translation is a human-readable preview of the term's translation in the
	// language passed to ListTerms. For plural terms it is the "other" form.
	Translation string
	// TranslationRaw is the unmodified JSON of the translation as returned by
	// POEditor — either a string or a plural object. Use it for exact equality
	// comparisons.
	TranslationRaw json.RawMessage
}
