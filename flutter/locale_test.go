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
		{"zn-Hans", flutter.Locale{Language: "zn", Script: "Hans"}, false},
		{"zn-hans", flutter.Locale{Language: "zn", Script: "Hans"}, false},
		{"zn-Hant", flutter.Locale{Language: "zn", Script: "Hant"}, false},
		{"zn-Hans-CN", flutter.Locale{Language: "zn", Script: "Hans", Country: "CN"}, false},
		{"zn-Hant-CN", flutter.Locale{Language: "zn", Script: "Hant", Country: "CN"}, false},
		{"zn-TW", flutter.Locale{Language: "zn", Country: "TW"}, false},
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
		{flutter.Locale{Language: "zn", Script: "Hans"}, "zn_Hans"},
		{flutter.Locale{Language: "zn", Script: "Hant", Country: "CN"}, "zn_Hant_CN"},
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
		{flutter.Locale{Language: "zn", Script: "Hans"}, "zn-hans"},
		{flutter.Locale{Language: "zn", Script: "Hant", Country: "CN"}, "zn-hant-cn"},
		{flutter.Locale{Language: "sr", Script: "Cyrl"}, "sr-cyrl"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Expected, func(t *testing.T) {
			actual := testCase.Input.StringHyphen()

			assert.Equal(t, testCase.Expected, actual)
		})
	}
}
