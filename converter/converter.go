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
	Term       string `json:"term"`
	Definition string `json:"definition"`
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
			return errors.Wrapf(err, "decoding term \"%s\" failed", term.Term)
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

	message := &arbMessage{
		Name:        term.Term,
		Translation: term.Definition,
		Attributes:  &arbMessageAttributes{},
	}

	return message, nil
}
