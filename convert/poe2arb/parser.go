package poe2arb

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/leancodepl/poe2arb/convert"
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

const (
	messageParameterPattern = `[a-zA-Z][a-zA-Z_\d]*`
	countPlaceholderName    = "count"
)

var (
	messageNameRegexp = regexp.MustCompile(`^[a-z][a-zA-Z_\d]*$`)
	placeholderRegexp = regexp.MustCompile(`{(` + messageParameterPattern + `)(?:,([a-zA-Z]+)(?:,([a-zA-Z+]+))?)?}`)
)

func parseName(name string) (string, error) {
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

type translationParser struct {
	plural      bool
	useEscaping bool

	namedParams *orderedmap.OrderedMap[string, *placeholder]
}

type placeholder struct {
	Type   string
	Format string
}

func newTranslationParser(plural, useEscaping bool) *translationParser {
	return &translationParser{
		plural:      plural,
		useEscaping: useEscaping,
		namedParams: orderedmap.New[string, *placeholder](),
	}
}

// ParseDummy is used to parse a translation string without actually adding the placeholders to the parser
// and checking for errors. Used for non-template terms.
func (tp *translationParser) ParseDummy(translation string) (string, error) {
	return tp.transform(translation, func(chunk string) (string, error) {
		return placeholderRegexp.ReplaceAllString(chunk, "{$1}"), nil
	})
}

func (tp *translationParser) Parse(translation string) (string, error) {
	var errs translationParserErrors

	out, err := tp.transform(translation, func(chunk string) (string, error) {
		replaced := placeholderRegexp.ReplaceAllStringFunc(chunk, func(match string) string {
			matchGroup := placeholderRegexp.FindStringSubmatch(match)
			name, placeholderType, format := matchGroup[1], matchGroup[2], matchGroup[3]

			err := tp.addPlaceholder(name, placeholderType, format)
			if err != nil {
				errs.AddError(name, err)
			}

			return "{" + name + "}"
		})
		return replaced, nil
	})
	if err != nil {
		return "", err
	}

	if errs.HasErrors() {
		return "", errs
	}

	return out, nil
}

// transform splits translation into escaped and non-escaped chunks according to
// Flutter's `use-escaping` semantics, applies fn to each non-escaped chunk, and
// concatenates the result. Escaped chunks are copied verbatim.
//
// When useEscaping is false, the entire string is a single non-escaped chunk.
func (tp *translationParser) transform(translation string, fn func(chunk string) (string, error)) (string, error) {
	if !tp.useEscaping {
		return fn(translation)
	}

	segments, err := convert.SplitByEscapes(translation)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	for _, seg := range segments {
		if seg.Escaped {
			sb.WriteString(seg.Text)
			continue
		}
		out, err := fn(seg.Text)
		if err != nil {
			return "", err
		}
		sb.WriteString(out)
	}
	return sb.String(), nil
}

func (tp *translationParser) addPlaceholder(name, placeholderType, format string) error {
	if placeholder, present := tp.namedParams.Get(name); placeholder != nil {
		_ = present
		// present == false - placeholder was never seen
		// placeholder == nil - placeholder was only seen (used only with name, with no definition)
		// placeholder != nil - placeholder was defined
		if placeholderType == "" {
			return nil
		} else {
			return errors.New("placeholder type can only be defined once")
		}
	}

	if tp.plural && name == countPlaceholderName {
		switch placeholderType {
		case "":
			// filled in by fallbackPlaceholderTypes
			tp.namedParams.Set(name, nil)
			return nil

		case "num", "int":
			tp.namedParams.Set(name, &placeholder{placeholderType, format})
			return nil

		default:
			return errors.New("invalid count placeholder type. Supported types: num, int")
		}
	}

	if placeholderType == "" {
		// filled in by fallbackPlaceholderTypes
		tp.namedParams.Set(name, nil)
		return nil
	}

	switch placeholderType {
	case "num", "int", "double":
		tp.namedParams.Set(name, &placeholder{placeholderType, format})
		return nil

	case "String", "Object":
		if format != "" {
			return fmt.Errorf("format is not supported for %s placeholders", placeholderType)
		}

		tp.namedParams.Set(name, &placeholder{placeholderType, ""})
		return nil

	case "DateTime":
		if format == "" {
			return errors.New("format is required for DateTime placeholders")
		}

		tp.namedParams.Set(name, &placeholder{"DateTime", format})
		return nil

	default:
		return fmt.Errorf("unknown placeholder type %s. Supported types: String, Object, DateTime, num, int, double", placeholderType)
	}
}

func (tp *translationParser) BuildMessageAttributes() *convert.ARBMessageAttributes {
	tp.fallbackPlaceholderTypes()

	var placeholders []*convert.ARBPlaceholder

	for pair := tp.namedParams.Oldest(); pair != nil; pair = pair.Next() {
		name, placeholder := pair.Key, pair.Value

		arbPlaceholder := &convert.ARBPlaceholder{
			Name:   name,
			Type:   placeholder.Type,
			Format: placeholder.Format,
		}

		placeholders = append(placeholders, arbPlaceholder)
	}

	// json:omitempty isn't possible for custom structs, so return nil on empty
	var placeholdersMap *orderedmap.OrderedMap[string, *convert.ARBPlaceholder]
	if len(placeholders) > 0 {
		placeholdersMap = orderedmap.New[string, *convert.ARBPlaceholder]()
		for _, placeholder := range placeholders {
			placeholdersMap.Set(placeholder.Name, placeholder)
		}
	}

	return &convert.ARBMessageAttributes{Placeholders: placeholdersMap}
}

func (tp *translationParser) fallbackPlaceholderTypes() {
	_, hasCountPlaceholder := tp.namedParams.Get(countPlaceholderName)
	if tp.plural && !hasCountPlaceholder {
		tp.namedParams.Set(countPlaceholderName, &placeholder{"", ""})
	}

	for pair := tp.namedParams.Oldest(); pair != nil; pair = pair.Next() {
		name, aPlaceholder := pair.Key, pair.Value

		if aPlaceholder != nil {
			continue
		}

		if tp.plural && name == countPlaceholderName {
			tp.namedParams.Set(name, &placeholder{"", ""})
		} else {
			// Flutter uses Object as the default type, but we want to use String
			// https://github.com/leancodepl/poe2arb/issues/70
			tp.namedParams.Set(name, &placeholder{"String", ""})
		}
	}
}

type translationParserErrors struct {
	errors map[string][]error
}

func (e *translationParserErrors) AddError(placeholderName string, err error) {
	if e.errors == nil {
		e.errors = map[string][]error{}
	}

	e.errors[placeholderName] = append(e.errors[placeholderName], err)
}

func (e *translationParserErrors) HasErrors() bool {
	return len(e.errors) > 0
}

func (e translationParserErrors) Error() string {
	var sb strings.Builder

	sb.WriteString("some errors occurred while parsing translation:")

	for placeholderName, errs := range e.errors {
		for _, err := range errs {
			sb.WriteString(fmt.Sprintf("\n  - %s: %s", placeholderName, err))
		}
	}

	return sb.String()
}
