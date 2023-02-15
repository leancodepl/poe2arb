package converter

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	orderedmap "github.com/wk8/go-ordered-map/v2"
)

const (
	messageParameterPattern = `[a-zA-Z][a-zA-Z_\d]*`
	countPlaceholderName    = "count"
)

var (
	messageNameRegexp = regexp.MustCompile(`^[a-z][a-zA-Z_\d]*$`)
	placeholderRegexp = regexp.MustCompile(`{(` + messageParameterPattern + `)(?:,([a-zA-Z]+)(?:,([a-zA-Z]+))?)?}`)
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
	plural bool

	namedParams *orderedmap.OrderedMap[string, *placeholder]
}

type placeholder struct {
	Type   string
	Format string
}

func newTranslationParser(plural bool) *translationParser {
	return &translationParser{
		plural:      plural,
		namedParams: orderedmap.New[string, *placeholder](),
	}
}

func (tpc *translationParser) Parse(message string) (string, error) {
	var errors translationParserErrors

	replaced := placeholderRegexp.ReplaceAllStringFunc(message, func(match string) string {
		matchGroup := placeholderRegexp.FindStringSubmatch(match)
		name, placeholderType, format := matchGroup[1], matchGroup[2], matchGroup[3]

		err := tpc.addPlaceholder(name, placeholderType, format)
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

func (tpc *translationParser) addPlaceholder(name, placeholderType, format string) error {
	if placeholder, _ := tpc.namedParams.Get(name); placeholder != nil {
		// doesn't exist - placeholder was never seen
		// exists but nil - placeholder was only seen (used only with name, with no definition)
		// exists and not nil - placeholder was defined
		if placeholderType == "" {
			return nil
		} else {
			return errors.New("placeholder type can only be defined once")
		}
	}

	if tpc.plural && name == countPlaceholderName {
		if placeholderType == "" {
			// filled in by fallbackPlaceholderTypes
			tpc.namedParams.Set(name, nil)
			return nil
		} else if placeholderType == "num" && format == "" {
			// Special edge-case, when plural variable doesn't have a type defined, it falls back to num
			// and because no actual type in ARB is specified, requires no format.
			// https://github.com/flutter/flutter/blob/1faa95009e947c66e8139903e11b1866365f282c/packages/flutter_tools/lib/src/localizations/gen_l10n_types.dart#L507-L512

			tpc.namedParams.Set(name, &placeholder{"", ""})
			return nil
		} else if placeholderType == "num" {
			tpc.namedParams.Set(name, &placeholder{"num", format})
			return nil
		} else if placeholderType == "int" {
			if format == "" {
				return errors.New("format is required for int plural placeholders")
			}

			tpc.namedParams.Set(name, &placeholder{"int", format})
			return nil
		}

		return errors.New("unknown placeholder type. Supported types: num, int")
	}

	if placeholderType == "" {
		// filled in by fallbackPlaceholderTypes
		tpc.namedParams.Set(name, nil)
		return nil
	}

	if format != "" {
		if placeholderType == "DateTime" {
			tpc.namedParams.Set(name, &placeholder{"DateTime", format})
			return nil
		} else if placeholderType == "num" || placeholderType == "int" || placeholderType == "double" {
			tpc.namedParams.Set(name, &placeholder{placeholderType, format})
			return nil
		} else {
			return errors.New("format is only supported for DateTime and int, num or double placeholders")
		}
	}

	if placeholderType == "String" || placeholderType == "Object" {
		tpc.namedParams.Set(name, &placeholder{placeholderType, ""})
		return nil
	}

	return errors.New("unknown placeholder type. Supported types: String, Object, DateTime, num, int, double")
}

func (tpc *translationParser) BuildMessageAttributes() *arbMessageAttributes {
	tpc.fallbackPlaceholderTypes()

	var placeholders []*arbPlaceholder

	for pair := tpc.namedParams.Oldest(); pair != nil; pair = pair.Next() {
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

func (tpc *translationParser) fallbackPlaceholderTypes() {
	_, hasCountPlaceholder := tpc.namedParams.Get(countPlaceholderName)
	if tpc.plural && !hasCountPlaceholder {
		tpc.namedParams.Set(countPlaceholderName, &placeholder{"", ""})
	}

	for pair := tpc.namedParams.Oldest(); pair != nil; pair = pair.Next() {
		name, aPlaceholder := pair.Key, pair.Value

		if aPlaceholder != nil {
			continue
		}

		if tpc.plural && name == countPlaceholderName {
			tpc.namedParams.Set(name, &placeholder{"", ""})
		} else {
			tpc.namedParams.Set(name, &placeholder{"Object", ""})
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

	sb.WriteString("some errors occurred while parsing translation:\n")

	for placeholderName, errs := range e.errors {
		for _, err := range errs {
			sb.WriteString(fmt.Sprintf("  - %s: %s\n", placeholderName, err))
		}
	}

	return sb.String()
}
