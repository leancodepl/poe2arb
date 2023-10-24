// Package arb2poe provides a converter from ARB to POEditor's JSON with params support.
package arb2poe

import (
	"encoding/json"
	"io"

	"github.com/leancodepl/poe2arb/convert"
	"github.com/pkg/errors"
)

type Converter struct {
	input io.Reader

	template bool
}

func NewConverter(input io.Reader, template bool) *Converter {
	return &Converter{
		input:    input,
		template: template,
	}
}

func (c *Converter) Convert(output io.Writer) (lang string, err error) {
	lang, messages, err := parseARB(c.input)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse ARB")
	}

	var poeTerms []*convert.POETerm
	for _, message := range messages {
		poeTerm, err := arbMessageToPOETerm(message, !c.template)
		if err != nil {
			return "", errors.Wrapf(err, "decoding term %q failed", message.Name)
		}

		poeTerms = append(poeTerms, poeTerm)
	}

	err = json.NewEncoder(output).Encode(poeTerms)
	if err != nil {
		return "", errors.Wrap(err, "failed to encode POEditor JSON")
	}

	return lang, nil
}
