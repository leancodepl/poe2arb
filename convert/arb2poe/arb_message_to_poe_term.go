package arb2poe

import (
	"errors"
	"regexp"
	"strings"

	"github.com/leancodepl/poe2arb/convert"
)

var (
	countPluralHeaderRegexp = regexp.MustCompile(`^{\s*count\s*,\s*plural\s*,\s*`)
	pluralCategoryRegexp    = regexp.MustCompile(`\s*(=0|=1|=2|zero|one|two|few|many|other)\s*{`)
)

func arbMessageToPOETerm(
	m *convert.ARBMessage,
	skipPlaceholderDefinitions bool,
	termPrefix string,
	useEscaping bool,
) (*convert.POETerm, error) {
	pluralDefinition, err := parseCountPlural(m.Translation, useEscaping)
	if err != nil {
		return nil, err
	}

	var definition convert.POETermDefinition
	// texts are the strings placeholder definitions can be added to, in the
	// order poe2arb parses them back.
	var texts []*string
	if pluralDefinition != nil {
		definition = convert.POETermDefinition{
			IsPlural: true,
			Plural:   pluralDefinition,
		}
		for _, form := range []*string{
			pluralDefinition.Zero, pluralDefinition.One, pluralDefinition.Two,
			pluralDefinition.Few, pluralDefinition.Many, &pluralDefinition.Other,
		} {
			if form != nil {
				texts = append(texts, form)
			}
		}
	} else {
		translation := m.Translation
		definition = convert.POETermDefinition{Value: &translation}
		texts = []*string{&translation}
	}

	if !skipPlaceholderDefinitions && m.Attributes != nil && m.Attributes.Placeholders != nil {
		for pair := m.Attributes.Placeholders.Oldest(); pair != nil; pair = pair.Next() {
			placeholderName, placeholder := pair.Key, pair.Value

			definitionAppend := ""
			if placeholder.Type != "" {
				definitionAppend += "," + placeholder.Type
			}
			if placeholder.Format != "" {
				definitionAppend += "," + placeholder.Format
			}

			// Only do the replacement for the first occurence (defining the same parameter multiple times is illegal)
			for _, text := range texts {
				found, err := indexPlaceholder(*text, placeholderName, useEscaping)
				if err != nil {
					return nil, err
				}
				if found == -1 {
					continue
				}

				index := 1 + found + len(placeholderName)
				*text = (*text)[:index] + definitionAppend + (*text)[index:]
				break
			}
		}
	}

	var termPlural string
	if definition.IsPlural {
		termPlural = "."
	}

	termName := m.Name
	if termPrefix != "" {
		termName = termPrefix + ":" + termName
	}

	return &convert.POETerm{
		Term:       termName,
		TermPlural: termPlural,
		Definition: definition,
	}, nil
}

// parseCountPlural converts a message with a `{count, plural, ...}`
// expression into POEditor plural forms, or returns nil if there is none.
//
// The plural may be embedded in text, e.g. `I have {count, plural, one {an
// apple} other {{count} apples}}`. POEditor plural terms have no notion of
// that, so the surrounding text is copied into every form: `I have an apple`,
// `I have {count} apples`.
func parseCountPlural(translation string, useEscaping bool) (*convert.POETermPluralDefinition, error) {
	start, end := -1, -1

	for i := 0; i < len(translation); i++ {
		switch translation[i] {
		case '\\':
			// escape character, ignore next bracket
			i++
		case '\'':
			if useEscaping {
				close := strings.IndexByte(translation[i+1:], '\'')
				if close == -1 {
					return nil, convert.ErrUnmatchedQuote
				}
				i += 1 + close
			}
		case '{':
			closing := matchingBrace(translation, i, useEscaping)
			if closing == -1 {
				// Not a valid ICU message, leave it for POEditor as is.
				return nil, nil
			}

			if countPluralHeaderRegexp.MatchString(translation[i:]) {
				if start != -1 {
					return nil, errors.New("POEditor plural terms support only one count plural per message")
				}
				start, end = i, closing
			}

			i = closing
		}
	}

	if start == -1 {
		return nil, nil
	}

	prefix, suffix := translation[:start], translation[end+1:]
	header := countPluralHeaderRegexp.FindString(translation[start:])

	pluralDefinition, err := parsePluralCases(translation[start+len(header):end], useEscaping)
	if err != nil {
		return nil, err
	}

	return pluralDefinition.Map(func(s string) (string, error) {
		return prefix + s + suffix, nil
	})
}

// parsePluralCases parses the `category {message}` cases of a plural
// expression.
func parsePluralCases(pluralsString string, useEscaping bool) (*convert.POETermPluralDefinition, error) {
	pluralDefinition := &convert.POETermPluralDefinition{}

	for {
		loc := pluralCategoryRegexp.FindStringSubmatchIndex(pluralsString)
		if loc == nil {
			break
		}

		pluralCategory := pluralsString[loc[2]:loc[3]]

		openingBrace := loc[1] - 1
		closingBrace := matchingBrace(pluralsString, openingBrace, useEscaping)
		if closingBrace == -1 {
			return nil, errors.New("unclosed plural category " + pluralCategory)
		}

		pluralDefinitionValue := pluralsString[openingBrace+1 : closingBrace]
		switch pluralCategory {
		case "=0", "zero":
			if pluralDefinition.Zero != nil {
				return nil, errors.New("multiple definitions for plural category zero")
			}

			pluralDefinition.Zero = &pluralDefinitionValue
		case "=1", "one":
			if pluralDefinition.One != nil {
				return nil, errors.New("multiple definitions for plural category one")
			}

			pluralDefinition.One = &pluralDefinitionValue
		case "=2", "two":
			if pluralDefinition.Two != nil {
				return nil, errors.New("multiple definitions for plural category two")
			}

			pluralDefinition.Two = &pluralDefinitionValue
		case "few":
			pluralDefinition.Few = &pluralDefinitionValue
		case "many":
			pluralDefinition.Many = &pluralDefinitionValue
		case "other":
			pluralDefinition.Other = pluralDefinitionValue
		}

		pluralsString = pluralsString[closingBrace+1:]
	}

	return pluralDefinition, nil
}

// matchingBrace returns the index of the brace closing the one at open, or -1
// if there is none. With useEscaping, braces inside ICU single-quote escape
// spans are ignored.
func matchingBrace(s string, open int, useEscaping bool) int {
	depth := 0
	for i := open; i < len(s); i++ {
		switch s[i] {
		case '\\':
			// escape character, ignore next bracket
			i++
		case '\'':
			if useEscaping {
				close := strings.IndexByte(s[i+1:], '\'')
				if close == -1 {
					return -1
				}
				i += 1 + close
			}
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return i
			}
		}
	}

	return -1
}

// indexPlaceholder returns the byte offset of the first `{name}` occurrence
// in translation, or -1 if not found. When useEscaping is true, matches inside
// ICU single-quote escape spans are skipped.
func indexPlaceholder(translation, name string, useEscaping bool) (int, error) {
	needle := "{" + name + "}"

	if !useEscaping {
		return strings.Index(translation, needle), nil
	}

	segments, err := convert.SplitByEscapes(translation)
	if err != nil {
		return 0, err
	}

	offset := 0
	for _, seg := range segments {
		if !seg.Escaped {
			if i := strings.Index(seg.Text, needle); i != -1 {
				return offset + i, nil
			}
		}
		offset += len(seg.Text)
	}
	return -1, nil
}
