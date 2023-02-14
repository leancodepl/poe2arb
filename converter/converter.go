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

		// if the plural did not ever use {count}, add it
		pc.addPlaceholder(countPlaceholderName, "num", "")

		if plural.Other != "" {
			value = plural.ToICUMessageFormat()
		} else {
			return nil, nil
			// TODO: Log note about missing "other" plural
		}
	}

	pc.fallbackPlaceholderTypes()

	message := &arbMessage{
		Name:        name,
		Translation: value,
		Attributes:  pc.buildMessageAttributes(),
	}

	return message, nil
}

const (
	messageParameterPattern = `[a-zA-Z][a-zA-Z_\d]*`
	countPlaceholderName    = "count"
)

var (
	messageNameRegexp = regexp.MustCompile(`^[a-z][a-zA-Z_\d]*$`)
	namedParamRegexp  = regexp.MustCompile("{(" + messageParameterPattern + ")}")
	placeholderRegexp = regexp.MustCompile(`{(` + messageParameterPattern + `)(?:,([a-zA-Z]+)(?:,([a-zA-Z]+))?)?}`)
)

type parseContext struct {
	plural bool

	namedParams *orderedmap.OrderedMap[string, *placeholder]
}

type placeholder struct {
	Type   string
	Format string
}

func (c *Converter) newParseContext(plural bool) *parseContext {
	return &parseContext{
		plural:      plural,
		namedParams: orderedmap.New[string, *placeholder](),
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
	var errors parseTranslationError

	replaced := placeholderRegexp.ReplaceAllStringFunc(message, func(match string) string {
		matchGroup := placeholderRegexp.FindStringSubmatch(match)
		name, placeholderType, format := matchGroup[1], matchGroup[2], matchGroup[3]

		err := pc.addPlaceholder(name, placeholderType, format)
		if err != nil {
			errors.AddError(name, err)
		}

		return "{" + name + "}"
	})

	if errors.HasErrors() {
		return "", errors
	}

	return replaced, nil
}

func (pc *parseContext) addPlaceholder(name, placeholderType, format string) error {
	if _, exists := pc.namedParams.Get(name); exists {
		if placeholderType == "" {
			return nil
		} else {
			return errors.New("placeholder type can only be defined once")
		}
	}

	if pc.plural && name == countPlaceholderName {
		if placeholderType == "" {
			// filled in by fallbackPlaceholderTypes
			pc.namedParams.Set(name, nil)
			return nil
		} else if placeholderType == "num" && format == "" {
			// Special edge-case, when plural variable doesn't have a type defined, it falls back to num
			// and because no actual type in ARB is specified, requires no format.
			// https://github.com/flutter/flutter/blob/1faa95009e947c66e8139903e11b1866365f282c/packages/flutter_tools/lib/src/localizations/gen_l10n_types.dart#L507-L512

			pc.namedParams.Set(name, &placeholder{"", ""})
			return nil
		} else if placeholderType == "num" {
			pc.namedParams.Set(name, &placeholder{"num", format})
			return nil
		} else if placeholderType == "int" {
			if format == "" {
				return errors.New("format is required for int plural placeholders")
			}

			pc.namedParams.Set(name, &placeholder{"int", format})
			return nil
		}

		return errors.New("unknown placeholder type. Supported types: num, int")
	}

	if placeholderType == "" {
		// filled in by fallbackPlaceholderTypes
		pc.namedParams.Set(name, nil)
		return nil
	}

	if format != "" {
		if placeholderType == "DateTime" {
			pc.namedParams.Set(name, &placeholder{"DateTime", format})
			return nil
		} else if placeholderType == "num" || placeholderType == "int" || placeholderType == "double" {
			pc.namedParams.Set(name, &placeholder{placeholderType, format})
			return nil
		} else {
			return errors.New("format is only supported for DateTime and int, num or double placeholders")
		}
	}

	if placeholderType == "String" || placeholderType == "Object" {
		pc.namedParams.Set(name, &placeholder{placeholderType, ""})
		return nil
	}

	return errors.New("unknown placeholder type. Supported types: String, Object, DateTime, num, int, double")
}

func (pc *parseContext) fallbackPlaceholderTypes() {
	for pair := pc.namedParams.Oldest(); pair != nil; pair = pair.Next() {
		name, aPlaceholder := pair.Key, pair.Value

		if aPlaceholder != nil {
			continue
		}

		if pc.plural && name == countPlaceholderName {
			pc.namedParams.Set(name, &placeholder{"", ""})
		} else {
			pc.namedParams.Set(name, &placeholder{"Object", ""})
		}
	}
}

func (pc parseContext) buildMessageAttributes() *arbMessageAttributes {
	var placeholders []*arbPlaceholder

	for pair := pc.namedParams.Oldest(); pair != nil; pair = pair.Next() {
		name, placeholder := pair.Key, pair.Value

		arbPlaceholder := &arbPlaceholder{
			Name:   name,
			Type:   placeholder.Type,
			Format: placeholder.Format,
		}

		placeholders = append(placeholders, arbPlaceholder)
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

type parseTranslationError struct {
	errors map[string][]error
}

func (e *parseTranslationError) AddError(placeholderName string, err error) {
	if e.errors == nil {
		e.errors = map[string][]error{}
	}

	e.errors[placeholderName] = append(e.errors[placeholderName], err)
}

func (e *parseTranslationError) HasErrors() bool {
	return len(e.errors) > 0
}

func (e parseTranslationError) Error() string {
	var sb strings.Builder

	sb.WriteString("some errors occurred while parsing translation:\n")

	for placeholderName, errs := range e.errors {
		for _, err := range errs {
			sb.WriteString(fmt.Sprintf("  - %s: %s\n", placeholderName, err))
		}
	}

	return sb.String()
}
