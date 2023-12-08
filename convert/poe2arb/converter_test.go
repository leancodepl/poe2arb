package poe2arb_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/leancodepl/poe2arb/convert/poe2arb"
	"github.com/leancodepl/poe2arb/flutter"
	"github.com/stretchr/testify/assert"
)

func TestConverterConvert(t *testing.T) {
	paths, err := filepath.Glob(filepath.Join("testdata", "*.input"))
	if err != nil {
		t.Fatal(err)
	}

	for _, path := range paths {
		_, filename := filepath.Split(path)
		testname := filename[:len(filename)-len(filepath.Ext(path))]

		t.Run(testname, func(t *testing.T) {
			source, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			template := !strings.Contains(testname, "-no-template")
			requireResourceAttributes := strings.Contains(testname, "-req-attrs")

			var termPrefix string
			if strings.Contains(testname, "-prefix") {
				termPrefix = "prefix"
			}

			goldenfile := filepath.Join("testdata", testname+".golden")
			golden, err := os.ReadFile(goldenfile)
			if err != nil {
				t.Fatal(err)
			}

			expect := string(golden)

			// Actual test
			actual, err := convert(t, string(source), template, requireResourceAttributes, termPrefix)

			assert.NoError(t, err)
			assert.Equal(t, expect, actual)
		})
	}

	// https://github.com/leancodepl/poe2arb/issues/41
	issue41Source := `[
		{
			"term": "testPlural",
			"definition": {
				"one": "",
				"few": "",
				"many": "",
				"other": ""
			},
			"context": "",
			"term_plural": "plural",
			"reference": "",
			"comment": ""
		}
	]
`

	t.Run("issue 41 template", func(t *testing.T) {
		actual, err := convert(t, issue41Source, true, false, "")

		assert.Error(t, err)
		assert.EqualError(t, err, `decoding term "testPlural" failed: missing "other" plural category`)
		assert.Equal(t, "", actual)
	})

	t.Run("issue 41 non-template", func(t *testing.T) {
		actual, err := convert(t, issue41Source, false, false, "")

		assert.NoError(t, err)
		assert.Equal(t, "{\n    \"@@locale\": \"en\"\n}\n", actual)
	})
}

func flutterMustParseLocale(lang string) flutter.Locale {
	locale, err := flutter.ParseLocale(lang)
	if err != nil {
		panic(err)
	}
	return locale
}

func convert(
	t *testing.T,
	input string,
	template bool,
	requireResourceAttributes bool,
	termPrefix string,
) (converted string, err error) {
	reader := strings.NewReader(input)
	conv := poe2arb.NewConverter(reader, &poe2arb.ConverterOptions{
		Locale:                    flutterMustParseLocale("en"),
		Template:                  template,
		RequireResourceAttributes: requireResourceAttributes,
		TermPrefix:                termPrefix,
	})
	out := new(bytes.Buffer)
	err = conv.Convert(out)

	converted = out.String()

	return
}
