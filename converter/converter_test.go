package converter_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/leancodepl/poe2arb/converter"
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

			goldenfile := filepath.Join("testdata", testname+".golden")
			golden, err := os.ReadFile(goldenfile)
			if err != nil {
				t.Fatal(err)
			}

			expect := string(golden)

			// Actual test
			actual, err := convert(t, string(source), template, requireResourceAttributes)

			assert.NoError(t, err)
			assert.Equal(t, expect, actual)
		})
	}

	// https://github.com/leancodepl/poe2arb/issues/41
	t.Run("issue 41", func(t *testing.T) {
		source := `[
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

		actual, err := convert(t, source, true, false)

		assert.Error(t, err)
		assert.EqualError(t, err, `decoding term "testPlural" failed: missing "other" plural category`)
		assert.Equal(t, "", actual)
	})
}

func convert(t *testing.T, input string, template bool, requireResourceAttributes bool) (converted string, err error) {
	reader := strings.NewReader(input)
	conv := converter.NewConverter(reader, "en", template, requireResourceAttributes)

	out := new(bytes.Buffer)
	err = conv.Convert(out)

	converted = out.String()

	return
}
