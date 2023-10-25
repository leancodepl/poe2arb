package convert

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func ptr(str string) *string { return &str }

func TestJSONTermDefinitionUnmarshalJSON(t *testing.T) {
	type testCase struct {
		Name                 string
		Input                string
		ExpectedValueExists  bool
		ExpectedValue        string
		ExpectedPluralExists bool
		ExpectedPlural       POETermPluralDefinition
	}

	cases := []testCase{
		{
			Name:                "simple string",
			Input:               `"some string"`,
			ExpectedValueExists: true,
			ExpectedValue:       "some string",
		},
		{
			Name:                "empty string",
			Input:               `""`,
			ExpectedValueExists: true,
			ExpectedValue:       "",
		},
		{
			Name:                 "plural with only other",
			Input:                `{"other": "Something"}`,
			ExpectedPluralExists: true,
			ExpectedPlural: POETermPluralDefinition{
				Other: "Something",
			},
		},
		{
			Name: "plural with all categories",
			Input: `{"zero": "Zero", "one": "One", "two": "Two",
            "few": "Few", "many": "Many", "other": "Other"}`,
			ExpectedPluralExists: true,
			ExpectedPlural: POETermPluralDefinition{
				Zero:  ptr("Zero"),
				One:   ptr("One"),
				Two:   ptr("Two"),
				Few:   ptr("Few"),
				Many:  ptr("Many"),
				Other: "Other",
			},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.Name, func(t *testing.T) {
			var d POETermDefinition
			err := json.Unmarshal([]byte(testCase.Input), &d)

			assert.NoError(t, err)

			if testCase.ExpectedValueExists {
				assert.NotNil(t, d.Value)
				assert.Equal(t, testCase.ExpectedValue, *d.Value)
			} else {
				assert.Nil(t, d.Value)
			}

			if testCase.ExpectedPluralExists {
				assert.NotNil(t, d.Plural)
				assert.Equal(t, testCase.ExpectedPlural, *d.Plural)
			} else {
				assert.Nil(t, d.Plural)
			}
		})
	}
}

func TestJSONTermPluralDefinitionToICUMessageFormat(t *testing.T) {
	type testCase struct {
		Name           string
		Input          POETermPluralDefinition
		ExpectedOutput string
	}

	cases := []testCase{
		{
			"only other",
			POETermPluralDefinition{Other: "test"},
			"{count, plural, other {test}}",
		},
		{
			"one and other",
			POETermPluralDefinition{One: ptr("foobar"), Other: "baz"},
			"{count, plural, =1 {foobar} other {baz}}",
		},
		{
			"all",
			POETermPluralDefinition{
				Zero: ptr("zero"), One: ptr("one"),
				Two: ptr("two"), Few: ptr("few"),
				Many: ptr("many"), Other: "other",
			},
			"{count, plural, =0 {zero} =1 {one} =2 {two} few {few} many {many} other {other}}",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.Name, func(t *testing.T) {
			output := testCase.Input.ToICUMessageFormat()
			assert.Equal(t, testCase.ExpectedOutput, output)
		})
	}
}
