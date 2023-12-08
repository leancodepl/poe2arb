package flutter

import (
	"slices"
	"strings"

	"github.com/pkg/errors"
)

type Locale struct {
	Language string
	Script   string
	Country  string
}

func ParseLocale(locale string) (Locale, error) {
	knownScripts := []string{
		"cyrl", "hans", "hant", // from https://poeditor.com/docs/languages
		"latn", // latn is not in the docs, but I put it here anyway
	}

	parts := strings.FieldsFunc(locale, func(r rune) bool {
		return r == '_' || r == '-'
	})

	switch len(parts) {
	case 1:
		return Locale{
			Language: parts[0],
		}, nil
	case 2:
		var script, country string
		for _, s := range knownScripts {
			if strings.ToLower(parts[1]) == s {
				script = s
				script = strings.ToUpper(script[:1]) + script[1:]
				break
			}
		}
		if script == "" {
			country = parts[1]
		}

		return Locale{
			Language: parts[0],
			Script:   script,
			Country:  strings.ToUpper(country),
		}, nil
	case 3:
		if !slices.Contains(knownScripts, strings.ToLower(parts[1])) {
			return Locale{}, errors.Errorf("invalid script: %s", parts[1])
		}

		return Locale{
			Language: parts[0],
			Script:   strings.ToUpper(parts[1][:1]) + parts[1][1:],
			Country:  strings.ToUpper(parts[2]),
		}, nil
	default:
		return Locale{}, errors.Errorf("invalid locale: %s", locale)
	}
}

func (l Locale) String() string {
	locale := l.Language
	if l.Script != "" {
		locale += "_" + l.Script
	}
	if l.Country != "" {
		locale += "_" + l.Country
	}
	return locale
}

func (l Locale) StringHyphen() string {
	locale := l.Language
	if l.Script != "" {
		locale += "-" + l.Script
	}
	if l.Country != "" {
		locale += "-" + l.Country
	}
	return strings.ToLower(locale)
}
