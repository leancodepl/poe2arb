package flutter_test

import (
	"testing"

	"github.com/leancodepl/poe2arb/flutter"
	"github.com/stretchr/testify/assert"
)

func TestParseLocale(t *testing.T) {
	testCases := []struct {
		Input          string
		ExpectedOutput flutter.Locale
		ExpectedError  bool
	}{
		{"en", flutter.Locale{Language: "en"}, false},
		{"en-US", flutter.Locale{Language: "en", Country: "US"}, false},
		{"en-us", flutter.Locale{Language: "en", Country: "US"}, false},
		{"en_US", flutter.Locale{Language: "en", Country: "US"}, false},
		{"zh-Hans", flutter.Locale{Language: "zh", Script: "Hans"}, false},
		{"zh-hans", flutter.Locale{Language: "zh", Script: "Hans"}, false},
		{"zh-Hant", flutter.Locale{Language: "zh", Script: "Hant"}, false},
		{"zh-Hans-CN", flutter.Locale{Language: "zh", Script: "Hans", Country: "CN"}, false},
		{"zh-Hant-CN", flutter.Locale{Language: "zh", Script: "Hant", Country: "CN"}, false},
		{"zh-TW", flutter.Locale{Language: "zh", Country: "TW"}, false},
		{"en-unknown", flutter.Locale{Language: "en", Country: "UNKNOWN"}, false},
		{"es-419", flutter.Locale{Language: "es", Country: "419"}, false},
		{"sr-Cyrl", flutter.Locale{Language: "sr", Script: "Cyrl"}, false},
		{"en-Wrong-GB", flutter.Locale{}, true},
		{"a-b-c-d", flutter.Locale{}, true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Input, func(t *testing.T) {
			actual, err := flutter.ParseLocale(testCase.Input)

			assert.Equal(t, testCase.ExpectedOutput, actual)

			if testCase.ExpectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestLocaleString(t *testing.T) {
	testCases := []struct {
		Input    flutter.Locale
		Expected string
	}{
		{flutter.Locale{Language: "en"}, "en"},
		{flutter.Locale{Language: "en", Country: "US"}, "en_US"},
		{flutter.Locale{Language: "es", Country: "419"}, "es_419"},
		{flutter.Locale{Language: "zh", Script: "Hans"}, "zh_Hans"},
		{flutter.Locale{Language: "zh", Script: "Hant", Country: "CN"}, "zh_Hant_CN"},
		{flutter.Locale{Language: "sr", Script: "Cyrl"}, "sr_Cyrl"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Expected, func(t *testing.T) {
			actual := testCase.Input.String()

			assert.Equal(t, testCase.Expected, actual)
		})
	}
}

func TestLocaleStringHyphen(t *testing.T) {
	testCases := []struct {
		Input    flutter.Locale
		Expected string
	}{
		{flutter.Locale{Language: "en"}, "en"},
		{flutter.Locale{Language: "en", Country: "US"}, "en-us"},
		{flutter.Locale{Language: "es", Country: "419"}, "es-419"},
		{flutter.Locale{Language: "zh", Script: "Hans"}, "zh-hans"},
		{flutter.Locale{Language: "zh", Script: "Hant", Country: "CN"}, "zh-hant-cn"},
		{flutter.Locale{Language: "sr", Script: "Cyrl"}, "sr-cyrl"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Expected, func(t *testing.T) {
			actual := testCase.Input.StringHyphen()

			assert.Equal(t, testCase.Expected, actual)
		})
	}
}
