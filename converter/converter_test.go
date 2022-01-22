package converter_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/leancodepl/poe2arb/converter"
	"github.com/stretchr/testify/assert"
)

func TestConverterConvert(t *testing.T) {
	type testCase struct {
		Name     string
		ElCompat bool
		Input    string
		Output   string
	}

	testCases := []testCase{
		{
			"Just text", false,
			`[
    {
        "term": "justText",
        "definition": "This is text",
        "context": "",
        "term_plural": "",
        "reference": "",
        "comment": ""
    }
]`,
			`{
    "@@locale": "en",
    "justText": "This is text",
    "@justText": {}
}
`,
		},
		{
			"Text with positional placeholder w/o elCompat", false,
			`[
    {
        "term": "textWithPositionalPlaceholder",
        "definition": "This is {}.",
        "context": "",
        "term_plural": "",
        "reference": "",
        "comment": ""
    }
]`,
			`{
    "@@locale": "en",
    "textWithPositionalPlaceholder": "This is {}.",
    "@textWithPositionalPlaceholder": {}
}
`,
		},
		{
			"Text with positional placeholder w/ elCompat", true,
			`[
    {
        "term": "textWithPositionalPlaceholder",
        "definition": "This is {}.",
        "context": "",
        "term_plural": "",
        "reference": "",
        "comment": ""
    }
]`,
			`{
    "@@locale": "en",
    "textWithPositionalPlaceholder": "This is {pos0}.",
    "@textWithPositionalPlaceholder": {
        "placeholders": {
            "pos0": {
                "type": "Object"
            }
        }
    }
}
`,
		},
		{
			"Text with few positional placeholders w/ elCompat", true,
			`[
    {
        "term": "textWithPositionalPlaceholders",
        "definition": "So {} is a {}.",
        "context": "",
        "term_plural": "",
        "reference": "",
        "comment": ""
    }
]`,
			`{
    "@@locale": "en",
    "textWithPositionalPlaceholders": "So {pos0} is a {pos1}.",
    "@textWithPositionalPlaceholders": {
        "placeholders": {
            "pos0": {
                "type": "Object"
            },
            "pos1": {
                "type": "Object"
            }
        }
    }
}
`,
		},
		{
			"Text with named placeholder", false,
			`[
    {
        "term": "textWithNamedPlaceholder",
        "definition": "This is {text}.",
        "context": "",
        "term_plural": "",
        "reference": "",
        "comment": ""
    }
]`,
			`{
    "@@locale": "en",
    "textWithNamedPlaceholder": "This is {text}.",
    "@textWithNamedPlaceholder": {
        "placeholders": {
            "text": {
                "type": "Object"
            }
        }
    }
}
`,
		},
		{
			"Text with unique named placeholders", false,
			`[
    {
        "term": "textWithUniqueNamedPlaceholders",
        "definition": "So {something} is a {somethingElse}.",
        "context": "",
        "term_plural": "",
        "reference": "",
        "comment": ""
    }
]`,
			`{
    "@@locale": "en",
    "textWithUniqueNamedPlaceholders": "So {something} is a {somethingElse}.",
    "@textWithUniqueNamedPlaceholders": {
        "placeholders": {
            "something": {
                "type": "Object"
            },
            "somethingElse": {
                "type": "Object"
            }
        }
    }
}
`,
		},
		{
			"Text with repeated named placeholder", false,
			`[
    {
        "term": "textWithRepeatedNamedPlaceholder",
        "definition": "So {something} is the same thing as {something}.",
        "context": "",
        "term_plural": "",
        "reference": "",
        "comment": ""
    }
]`,
			`{
    "@@locale": "en",
    "textWithRepeatedNamedPlaceholder": "So {something} is the same thing as {something}.",
    "@textWithRepeatedNamedPlaceholder": {
        "placeholders": {
            "something": {
                "type": "Object"
            }
        }
    }
}
`,
		},
		{
			"Text with double quotes", false,
			`[
    {
        "term": "textWithDoubleQuotes",
        "definition": "Those are some \"quotes\".",
        "context": "",
        "term_plural": "",
        "reference": "",
        "comment": ""
    }
]`,
			`{
    "@@locale": "en",
    "textWithDoubleQuotes": "Those are some \"quotes\".",
    "@textWithDoubleQuotes": {}
}
`,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			out := new(bytes.Buffer)
			conv := converter.NewConverter(testCase.ElCompat)
			err := conv.Convert(strings.NewReader(testCase.Input), out, "en")

			assert.NoError(t, err)
			assert.Equal(t, testCase.Output, out.String())
		})
	}
}
