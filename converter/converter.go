package converter

import (
	"encoding/json"
	"fmt"
	"io"
	"regexp"

	"github.com/leancodepl/poe2arb/utils"
	"github.com/pkg/errors"
)

func Convert(input io.Reader, output io.Writer, lang string) error {
	var jsonContents []*jsonTerm
	err := json.NewDecoder(input).Decode(&jsonContents)
	if err != nil {
		return errors.Wrap(err, "decoding json failed")
	}

	arb := utils.NewOrderedMap()
	arb.Set(localeKey, lang)

	for _, term := range jsonContents {
		message, err := parseTerm(term)
		if err != nil {
			return errors.Wrapf(err, `decoding term "%s" failed`, term.Term)
		}

		arb.Set(message.Name, message.Translation)
		arb.Set("@"+message.Name, message.Attributes)
	}

	encoder := json.NewEncoder(output)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "    ") // 4 spaces

	err = encoder.Encode(arb)
	return errors.Wrap(err, "encoding arb failed")
}

func parseTerm(term *jsonTerm) (*arbMessage, error) {
	var value string
	pc := newParseContext()

	if term.Definition.Value != nil {
		var err error
		value, err = pc.parseTranslation(*term.Definition.Value)
		if err != nil {
			return nil, err
		}
	} else {
		plural, err := term.Definition.Plural.Map(func(s string) (string, error) {
			s, err := pc.parseTranslation(s)
			return s, err
		})
		if err != nil {
			return nil, err
		}

		pc.namedParams["count"] = "num"

		value = plural.ToICUMessageFormat()
	}

	message := &arbMessage{
		Name:        term.Term,
		Translation: value,
		Attributes:  pc.buildMessageAttributes(),
	}

	return message, nil
}

const messageParameterPattern = `[a-zA-Z][a-zA-Z_\d]*`

var (
	posParamRegexp   = regexp.MustCompile("{}")
	namedParamRegexp = regexp.MustCompile("{(" + messageParameterPattern + ")}")
)

type parseContext struct {
	posParamCount int
	namedParams   map[string]string // name to type
}

func newParseContext() *parseContext {
	return &parseContext{
		posParamCount: -1,
		namedParams:   make(map[string]string),
	}
}

func (pc *parseContext) parseTranslation(message string) (string, error) {
	// Positional params. Ex.: This is a {}.
	// Parses translations replacing `{}` parameters with placeholders
	// pos0, pos1, ... This is for a compatibility with easy_localization strings.
	message = posParamRegexp.ReplaceAllStringFunc(message, func(s string) string {
		pc.posParamCount++
		return fmt.Sprintf("{pos%d}", pc.posParamCount)
	})

	// Named params. Ex.: This is a {param}.
	namedMatches := namedParamRegexp.FindAllStringSubmatch(message, -1)
	for _, matchGroup := range namedMatches {
		name := matchGroup[1]
		pc.namedParams[name] = "Object"
	}

	return message, nil
}

func (pc parseContext) buildMessageAttributes() *arbMessageAttributes {
	var placeholders []*arbPlaceholder

	for i := 0; i <= pc.posParamCount; i++ {
		placeholders = append(
			placeholders,
			&arbPlaceholder{
				Name: fmt.Sprintf("pos%d", i),
				Type: "Object",
			},
		)
	}

	for name, pType := range pc.namedParams {
		placeholders = append(placeholders, &arbPlaceholder{name, pType})
	}

	placeholdersMap := make(map[string]*arbPlaceholder, len(placeholders))
	for _, placeholder := range placeholders {
		placeholdersMap[placeholder.Name] = placeholder
	}

	return &arbMessageAttributes{Placeholders: placeholdersMap}
}
