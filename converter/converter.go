// Package converter handles coversion from POEditor's JSON to Flutter's ARB.
package converter

import (
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/pkg/errors"
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

type Converter struct {
	EasyLocalizationCompat bool
}

func NewConverter(easyLocalizationCompat bool) *Converter {
	return &Converter{
		EasyLocalizationCompat: easyLocalizationCompat,
	}
}

func (c *Converter) Convert(input io.Reader, output io.Writer, lang string) error {
	var jsonContents []*jsonTerm
	err := json.NewDecoder(input).Decode(&jsonContents)
	if err != nil {
		return errors.Wrap(err, "decoding json failed")
	}

	arb := orderedmap.New[string, any]()
	arb.Set(localeKey, lang)

	for _, term := range jsonContents {
		message, err := c.parseTerm(term)
		if err != nil {
			return errors.Wrapf(err, `decoding term "%s" failed`, term.Term)
		}

		if message != nil {
			arb.Set(message.Name, message.Translation)
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
	pc := c.newParseContext(term.Definition.IsPlural)

	name, err := pc.parseName(term.Term)
	if err != nil {
		return nil, err
	}

	if !term.Definition.IsPlural {
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

		pc.namedParams.Set("count", "num")

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
	plural   bool
	elCompat bool

	posParamCount int
	namedParams   *orderedmap.OrderedMap[string, string] // name to type
}

func (c *Converter) newParseContext(plural bool) *parseContext {
	return &parseContext{
		plural:        plural,
		elCompat:      c.EasyLocalizationCompat,
		posParamCount: -1,
		namedParams:   orderedmap.New[string, string](),
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
	if pc.elCompat && !pc.plural {
		// Positional params. Ex.: This is a {}.
		// Parses translations replacing `{}` parameters with placeholders
		// pos0, pos1, ... This is for a compatibility with easy_localization strings.
		// Positional params are not supported in plurals.
		message = posParamRegexp.ReplaceAllStringFunc(message, func(s string) string {
			pc.posParamCount++
			return fmt.Sprintf("{pos%d}", pc.posParamCount)
		})
	}

	// Named params. Ex.: This is a {param}.
	namedMatches := namedParamRegexp.FindAllStringSubmatch(message, -1)
	for _, matchGroup := range namedMatches {
		name := matchGroup[1]
		pc.namedParams.Set(name, "Object")
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

	for pair := pc.namedParams.Oldest(); pair != nil; pair = pair.Next() {
		name, placeholderType := pair.Key, pair.Value

		placeholder := &arbPlaceholder{
			Name: name,
			Type: placeholderType,
		}

		if placeholderType == "num" {
			placeholder.Format = "decimalPattern"
		}

		placeholders = append(placeholders, placeholder)
	}

	// json:omitempty isn't possible for custom structs, so return nil on empty
	var placeholdersMap *orderedmap.OrderedMap[string, *arbPlaceholder]
	if len(placeholders) > 0 {
		placeholdersMap = orderedmap.New[string, *arbPlaceholder]()
		for _, placeholder := range placeholders {
			placeholdersMap.Set(placeholder.Name, placeholder)
		}
	}

	return &arbMessageAttributes{Placeholders: placeholdersMap}
}
