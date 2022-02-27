package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrefixFromTemplateFileName(t *testing.T) {
	type testCase struct {
		Filename       string
		ExpectedPrefix string
		ExpectedError  bool
	}

	testCases := []testCase{
		{"app_en.arb", "app_", false},
		{"some-prefix_en.arb", "some-prefix_", false},
		{"app_en-US.arb", "app_", false},
		{"app_zh-Hant-CN.arb", "app_", false},
		{"app.arb", "", true},
		{"app-xd.arb", "", true}, // "xd" is not a valid ISO 639 country
		{"en.arb", "", true},     // must have prefix ending with "_"
	}

	for _, testCase := range testCases {
		t.Run(testCase.Filename, func(t *testing.T) {
			prefix, err := prefixFromTemplateFileName(testCase.Filename)

			assert.Equal(t, testCase.ExpectedPrefix, prefix)
			if testCase.ExpectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
