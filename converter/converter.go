package converter

import (
	"encoding/json"
	"io"

	"github.com/leancodepl/poe2arb/utils"
	"github.com/pkg/errors"
)

type Converter struct {
	Input  io.Reader
	Output io.Writer
	Lang   string
}

func NewConverter(input io.Reader, output io.Writer, lang string) *Converter {
	return &Converter{
		Input:  input,
		Output: output,
		Lang:   lang,
	}
}

type jsonTerm struct {
	Term       string             `json:"term"`
	Definition jsonTermDefinition `json:"definition"`
}

type jsonTermDefinition struct {
	Value  *string
	Plural *jsonTermPluralDefinition
}

func (d *jsonTermDefinition) UnmarshalJSON(data []byte) error {
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	switch v := v.(type) {
	case string:
		d.Value = &v
		return nil
	case map[string]interface{}:
		return json.Unmarshal(data, &d.Plural)
	}

	return errors.New("invalid definition type")
}

type jsonTermPluralDefinition struct {
	Zero  *string `json:"zero"`
	One   *string `json:"one"`
	Two   *string `json:"two"`
	Few   *string `json:"few"`
	Many  *string `json:"many"`
	Other string  `json:"other"`
}

const (
	localeKey = "@@locale"
)

type arbMessage struct {
	Name        string
	Translation string
	Attributes  *arbMessageAttributes
}

type arbMessageAttributes struct {
	Description  string                     `json:"description,omitempty"`
	Placeholders map[string]*arbPlaceholder `json:"placeholders,omitempty"`
}

type arbPlaceholder struct {
	Name string `json:"-"`
	Type string `json:"type,omitempty"`
}

func (c *Converter) Convert() error {
	var jsonContents []*jsonTerm
	err := json.NewDecoder(c.Input).Decode(&jsonContents)
	if err != nil {
		return errors.Wrap(err, "decoding json failed")
	}

	arb := utils.NewOrderedMap()
	arb.Set(localeKey, c.Lang)

	for _, term := range jsonContents {
		message, err := parseTerm(term)
		if err != nil {
			return errors.Wrapf(err, `decoding term "%s" failed`, term.Term)
		}

		arb.Set(message.Name, message.Translation)
		arb.Set("@"+message.Name, message.Attributes)
	}

	encoder := json.NewEncoder(c.Output)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "    ") // 4 spaces

	err = encoder.Encode(arb)
	return errors.Wrap(err, "encoding arb failed")
}

func parseTerm(term *jsonTerm) (*arbMessage, error) {
	// todo: icu support

	var value string
	if term.Definition.Value != nil {
		value = *term.Definition.Value
	} else {
		value = "<empty>"
	}

	message := &arbMessage{
		Name:        term.Term,
		Translation: value,
		Attributes:  &arbMessageAttributes{},
	}

	return message, nil
}
