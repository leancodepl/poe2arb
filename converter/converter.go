// Package converter handles coversion from POEditor's JSON to Flutter's ARB.
package converter

import (
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"

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

	name, err := pc.parseName(term.Term)
	if err != nil {
		return nil, err
	}

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
		Name:        name,
		Translation: value,
		Attributes:  pc.buildMessageAttributes(),
	}

	return message, nil
}

const messageParameterPattern = `[a-zA-Z][a-zA-Z_\d]*`

var (
	messageNameRegexp = regexp.MustCompile(`^[a-z][a-zA-Z_\d]*$`)
	posParamRegexp    = regexp.MustCompile("{}")
	namedParamRegexp  = regexp.MustCompile("{(" + messageParameterPattern + ")}")
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

func (parseContext) parseName(name string) (string, error) {
	// lowercase first letter, Flutter gen-l10n doesn't allow first letter uppercase
	// https://github.com/flutter/flutter/blob/fae84f67140cbaa7a07ed5c82ee99f31c7bb1f0e/packages/flutter_tools/lib/src/localizations/gen_l10n.dart#L1056
	name = strings.ToLower(name[:1]) + name[1:]

	// replace dots with an underscore
	name = strings.ReplaceAll(name, ".", "_")

	if !messageNameRegexp.MatchString(name) {
		return "", errors.New("term name must start with lowercase letter followed by any number of anycase letter, digit or underscore")
	}

	return name, nil
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
		placeholder := &arbPlaceholder{
			Name: name,
			Type: pType,
		}

		if pType == "num" {
			placeholder.Format = "decimalPattern"
		}

		placeholders = append(placeholders, placeholder)
	}

	placeholdersMap := make(map[string]*arbPlaceholder, len(placeholders))
	for _, placeholder := range placeholders {
		placeholdersMap[placeholder.Name] = placeholder
	}

	return &arbMessageAttributes{Placeholders: placeholdersMap}
}
