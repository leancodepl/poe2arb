package converter

import (
	"bytes"
	"encoding/json"
	"strings"
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
		ExpectedPlural       jsonTermPluralDefinition
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
			ExpectedPlural: jsonTermPluralDefinition{
				Other: "Something",
			},
		},
		{
			Name: "plural with all categories",
			Input: `{"zero": "Zero", "one": "One", "two": "Two",
            "few": "Few", "many": "Many", "other": "Other"}`,
			ExpectedPluralExists: true,
			ExpectedPlural: jsonTermPluralDefinition{
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
			var d jsonTermDefinition
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

func TestConverterConvert(t *testing.T) {
	in := `[
    {
        "term": "welcomeCallToAction",
        "definition": "Zarz\u0105dzaj zasobami",
        "context": "",
        "term_plural": "",
        "reference": "",
        "comment": ""
    },
    {
        "term": "welcomeSignIn",
        "definition": "ZALOGUJ SI\u0118",
        "context": "",
        "term_plural": "",
        "reference": "",
        "comment": ""
    },
    {
        "term": "signInAppBarTitle",
        "definition": "Zaloguj si\u0119",
        "context": "",
        "term_plural": "",
        "reference": "",
        "comment": ""
    }
]`
	expectedOut := `{
    "@@locale": "pl",
    "welcomeCallToAction": "Zarządzaj zasobami",
    "@welcomeCallToAction": {},
    "welcomeSignIn": "ZALOGUJ SIĘ",
    "@welcomeSignIn": {},
    "signInAppBarTitle": "Zaloguj się",
    "@signInAppBarTitle": {}
}
`

	out := new(bytes.Buffer)
	conv := NewConverter(strings.NewReader(in), out, "pl")
	err := conv.Convert()

	assert.NoError(t, err)
	assert.Equal(t, expectedOut, out.String())
}
