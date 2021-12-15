package converter_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/leancodepl/poe2arb/converter"
	"github.com/stretchr/testify/assert"
)

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
	conv := converter.NewConverter(strings.NewReader(in), out, "pl")
	err := conv.Convert()

	assert.NoError(t, err)
	assert.Equal(t, expectedOut, out.String())
}
