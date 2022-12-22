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

			goldenfile := filepath.Join("testdata", testname+".golden")
			golden, err := os.ReadFile(goldenfile)
			if err != nil {
				t.Fatal(err)
			}

			expect := string(golden)

			// Actual test
			out := new(bytes.Buffer)
			conv := converter.NewConverter()
			err = conv.Convert(strings.NewReader(string(source)), out, "en")

			actual := out.String()

			assert.NoError(t, err)
			assert.Equal(t, expect, actual)
		})
	}
}
