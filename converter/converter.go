// Package converter handles coversion from POEditor's JSON to Flutter's ARB.
package converter

import (
	"encoding/json"
	"io"

	"github.com/pkg/errors"
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

type Converter struct {
	input io.Reader

	lang     string
	template bool

	requireResourceAttributes bool
}

func NewConverter(
	input io.Reader,
	lang string,
	template bool,
	requireResourceAttributes bool,
) *Converter {
	return &Converter{
		input: input,

		lang:     lang,
		template: template,

		requireResourceAttributes: requireResourceAttributes,
	}
}

func (c *Converter) Convert(output io.Writer) error {
	var jsonContents []*jsonTerm
	err := json.NewDecoder(c.input).Decode(&jsonContents)
	if err != nil {
		return errors.Wrap(err, "decoding json failed")
	}

	arb := orderedmap.New[string, any]()
	arb.Set(localeKey, c.lang)

	for _, term := range jsonContents {
		message, err := c.parseTerm(term)
		if err != nil {
			return errors.Wrapf(err, `decoding term "%s" failed`, term.Term)
		}

		if message == nil {
			continue
		}

		if !c.template && message.Translation == "" {
			// Don't generate terms for empty translations if we're not generating a template
			// https://github.com/leancodepl/poe2arb/issues/42
			continue
		}

		arb.Set(message.Name, message.Translation)

		if c.template &&
			(c.requireResourceAttributes ||
				message.Attributes != nil && !message.Attributes.IsEmpty()) {
			arb.Set("@"+message.Name, message.Attributes)
		}
	}

	encoder := json.NewEncoder(output)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "    ") // 4 spaces

	err = encoder.Encode(arb)
	return errors.Wrap(err, "encoding arb failed")
}

func (c Converter) parseTerm(term *jsonTerm) (*arbMessage, error) {
	var value string
	pc := newTranslationParser(term.Definition.IsPlural)

	name, err := parseName(term.Term)
	if err != nil {
		return nil, err
	}

	if !term.Definition.IsPlural {
		var err error
		value, err = pc.Parse(*term.Definition.Value)
		if err != nil {
			return nil, err
		}
	} else {
		plural, err := term.Definition.Plural.Map(func(s string) (string, error) {
			s, err := pc.Parse(s)
			return s, err
		})
		if err != nil {
			return nil, err
		}

		if plural.Other != "" {
			value = plural.ToICUMessageFormat()
		} else {
			return nil, nil
			// TODO: Log note about missing "other" plural
		}
	}

	message := &arbMessage{
		Name:        name,
		Translation: value,
		Attributes:  pc.BuildMessageAttributes(),
	}

	return message, nil
}
