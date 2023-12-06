// Package arb2poe provides a converter from ARB to POEditor's JSON with params support.
package arb2poe

import (
	"encoding/json"
	"io"

	"github.com/leancodepl/poe2arb/convert"
	"github.com/leancodepl/poe2arb/flutter"
	"github.com/pkg/errors"
)

type Converter struct {
	input io.Reader

	templateLocale flutter.Locale
	termPrefix     string
}

func NewConverter(input io.Reader, templateLocale flutter.Locale, termPrefix string) *Converter {
	return &Converter{
		input:          input,
		templateLocale: templateLocale,
		termPrefix:     termPrefix,
	}
}

var NoTermsError = errors.New("no terms to convert")

func (c *Converter) Convert(output io.Writer) (lang flutter.Locale, err error) {
	lang, messages, err := parseARB(c.input)
	if err != nil {
		return flutter.Locale{}, errors.Wrap(err, "failed to parse ARB")
	}

	template := c.templateLocale == lang

	var poeTerms []*convert.POETerm
	for _, message := range messages {
		poeTerm, err := arbMessageToPOETerm(message, !template, c.termPrefix)
		if err != nil {
			return flutter.Locale{}, errors.Wrapf(err, "decoding term %q failed", message.Name)
		}

		poeTerms = append(poeTerms, poeTerm)
	}

	if len(poeTerms) == 0 {
		return flutter.Locale{}, NoTermsError
	}

	err = json.NewEncoder(output).Encode(poeTerms)
	if err != nil {
		return flutter.Locale{}, errors.Wrap(err, "failed to encode POEditor JSON")
	}

	return lang, nil
}
