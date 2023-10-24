package arb2poe

import (
	"regexp"
	"strings"

	"github.com/leancodepl/poe2arb/convert"
	"github.com/pkg/errors"
)

func arbMessageToPOETerm(
	m *convert.ARBMessage,
	skipPlaceholderDefinitions bool,
	termPrefix string,
) (*convert.POETerm, error) {
	translation := m.Translation
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
			found := strings.Index(translation, "{"+placeholderName+"}")
			if found == -1 {
				continue
			}

			index := 1 + found + len(placeholderName)

			translation = translation[:index] + definitionAppend + translation[index:]
		}
	}

	var definition convert.POETermDefinition
	pluralRegexp := regexp.MustCompile(`^{count,\s*plural,\s*(.+)}$`)
	if matches := pluralRegexp.FindStringSubmatch(translation); len(matches) > 0 {
		pluralDefinition := &convert.POETermPluralDefinition{}

		pluralCategoryRegexp := regexp.MustCompile(`\s*(=0|=1|=2|zero|one|two|few|many|other)\s*{`)
		pluralsString := matches[1]

		for {
			match := pluralCategoryRegexp.FindStringSubmatch(pluralsString)
			if len(match) == 0 {
				break
			}

			pluralCategory := match[1]

			pluralsString = pluralsString[len(match[0]):]

			depth := 1
			endLength := 0
			for {
				findString := pluralsString[endLength:]

				if findString[0] == '\\' {
					// escape character, ignore next bracket
					endLength += 2
					continue
				} else if findString[0] == '{' {
					depth++
				} else if findString[0] == '}' {
					depth--
				}

				endLength++

				if depth == 0 {
					break
				}

			}

			pluralDefinitionValue := pluralsString[:endLength-1] // -1 to remove the closing bracket
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

			pluralsString = pluralsString[endLength:]
		}

		definition = convert.POETermDefinition{
			IsPlural: true,
			Plural:   pluralDefinition,
		}
	} else {
		definition = convert.POETermDefinition{Value: &translation}
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
