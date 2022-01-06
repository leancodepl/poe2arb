package converter

import (
	"encoding/json"
	"fmt"
	"io"
	"regexp"

	"github.com/leancodepl/poe2arb/utils"
	"github.com/pkg/errors"
)

type Converter struct {
	Input  io.Reader
	Output io.Writer
	Lang   string

	posParamCount int
	namedParams   map[string]string // name to type
}

func NewConverter(input io.Reader, output io.Writer, lang string) *Converter {
	return &Converter{
		Input:         input,
		Output:        output,
		Lang:          lang,
		posParamCount: -1,
		namedParams:   make(map[string]string),
	}
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
		message, err := c.parseTerm(term)
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

func (c *Converter) parseTerm(term *jsonTerm) (*arbMessage, error) {
	var value string
	if term.Definition.Value != nil {
		var err error
		value, err = c.parseTranslation(*term.Definition.Value)
		if err != nil {
			return nil, err
		}
	} else {
		plural, err := term.Definition.Plural.Map(func(s string) (string, error) {
			s, err := c.parseTranslation(s)
			return s, err
		})
		if err != nil {
			return nil, err
		}

		c.namedParams["count"] = "num"

		value = plural.ToICUMessageFormat()
	}

	message := &arbMessage{
		Name:        term.Term,
		Translation: value,
		Attributes:  c.buildMessageAttributes(),
	}

	return message, nil
}

const messageParameterPattern = `[a-zA-Z][a-zA-Z_\d]*`

var (
	posParamRegexp   = regexp.MustCompile("{}")
	namedParamRegexp = regexp.MustCompile("{(" + messageParameterPattern + ")}")
)

func (c *Converter) parseTranslation(message string) (string, error) {
	// Positional params. Ex.: This is a {}.
	// Parses translations replacing `{}` parameters with placeholders
	// pos0, pos1, ... This is for a compatibility with easy_localization strings.
	message = posParamRegexp.ReplaceAllStringFunc(message, func(s string) string {
		c.posParamCount++
		return fmt.Sprintf("{pos%d}", c.posParamCount)
	})

	// Named params. Ex.: This is a {param}.
	namedMatches := namedParamRegexp.FindAllStringSubmatch(message, -1)
	for _, matchGroup := range namedMatches {
		name := matchGroup[1]
		c.namedParams[name] = "Object"
	}

	return message, nil
}

func (c Converter) buildMessageAttributes() *arbMessageAttributes {
	var placeholders []*arbPlaceholder

	for i := 0; i <= c.posParamCount; i++ {
		placeholders = append(
			placeholders,
			&arbPlaceholder{
				Name: fmt.Sprintf("pos%d", i),
				Type: "Object",
			},
		)
	}

	for name, pType := range c.namedParams {
		placeholders = append(placeholders, &arbPlaceholder{name, pType})
	}

	placeholdersMap := make(map[string]*arbPlaceholder, len(placeholders))
	for _, placeholder := range placeholders {
		placeholdersMap[placeholder.Name] = placeholder
	}

	return &arbMessageAttributes{Placeholders: placeholdersMap}
}
