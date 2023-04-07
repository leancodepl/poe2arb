package converter

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestParseName(t *testing.T) {
	type testCase struct {
		Input        string
		ExpectedName string
		ExpectError  bool
	}

	cases := []testCase{
		{"someName", "someName", false},
		{"some_name", "some_name", false},
		{"some.name", "some_name", false},
		{"some.....name", "some_____name", false},
		{"someName1", "someName1", false},
		{"some/name", "", true},
		{"some-name", "", true},
		{"_someName", "", true},
		{"1someName", "", true},
	}

	for _, c := range cases {
		t.Run(c.Input, func(t *testing.T) {
			name, err := parseName(c.Input)
			assert.Equal(t, c.ExpectedName, name)
			if c.ExpectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTranslationParseDummy(t *testing.T) {
	type testCase struct {
		Input          string
		ExpectedOutput string
	}

	cases := []testCase{
		{"some text", "some text"},
		{"some {placeholder} text", "some {placeholder} text"},
		{"a {placeholder} with {placeholder,DateTime,yMd} {default} and {default}", "a {placeholder} with {placeholder} {default} and {default}"},
	}

	for _, testCase := range cases {
		t.Run(testCase.Input, func(t *testing.T) {
			parser := newTranslationParser(false)

			output := parser.ParseDummy(testCase.Input)

			assert.Equal(t, testCase.ExpectedOutput, output)
		})
	}
}

func TestTranslationParserParseErrors(t *testing.T) {
	type testCase struct {
		TestName             string
		Plural               bool
		Input                string
		ExpectedOutput       string
		ExpectedPlaceholders map[string]*placeholder
		ExpectedError        string
	}

	cases := []testCase{
		{
			TestName:       "no placeholders",
			Input:          "some text",
			ExpectedOutput: "some text",
		},
		{
			TestName:       "one simple placeholder",
			Input:          "some {placeholder} text",
			ExpectedOutput: "some {placeholder} text",
			ExpectedPlaceholders: map[string]*placeholder{
				"placeholder": nil,
			},
		},
		{
			TestName:       "complex placeholders",
			Input:          "a {placeholder} with {placeholder,DateTime,yMd} {default} and {default}",
			ExpectedOutput: "a {placeholder} with {placeholder} {default} and {default}",
			ExpectedPlaceholders: map[string]*placeholder{
				"placeholder": {"DateTime", "yMd"},
				"default":     nil,
			},
		},
		{
			TestName:      "placeholder double definitions",
			Input:         "{placeholder,String} with {placeholder,DateTime,yMd}",
			ExpectedError: "some errors occurred while parsing translation:\n  - placeholder: placeholder type can only be defined once",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.TestName, func(t *testing.T) {
			parser := newTranslationParser(testCase.Plural)

			output, err := parser.Parse(testCase.Input)

			assert.Equal(t, testCase.ExpectedOutput, output)

			if testCase.ExpectedError != "" {
				assert.Error(t, err)
				assert.EqualError(t, err, testCase.ExpectedError)
			} else {
				assert.NoError(t, err)
			}

			for name, placeholder := range testCase.ExpectedPlaceholders {
				expectedPlaceholder, ok := parser.namedParams.Get(name)
				assert.True(t, ok)
				assert.Equal(t, placeholder, expectedPlaceholder)
			}
		})
	}
}

func TestTranslationParserAddPlaceholder(t *testing.T) {
	type testCase struct {
		TestName string

		Plural              bool
		InitialPlaceholders map[string]*placeholder

		Name   string
		Type   string
		Format string

		ExpectedError        string
		ExpectedPlaceholders map[string]*placeholder
	}

	cases := []testCase{
		{
			TestName: "just name",
			Name:     "param",
			ExpectedPlaceholders: map[string]*placeholder{
				"param": nil,
			},
		},
		{
			TestName: "name and type Object",
			Name:     "param",
			Type:     "Object",
			ExpectedPlaceholders: map[string]*placeholder{
				"param": {"Object", ""},
			},
		},
		{
			TestName:      "name, type Object and format",
			Name:          "param",
			Type:          "Object",
			Format:        "format",
			ExpectedError: "format is not supported for Object placeholders",
		},
		{
			TestName:      "name, type String and format",
			Name:          "param",
			Type:          "String",
			Format:        "format",
			ExpectedError: "format is not supported for String placeholders",
		},
		{
			TestName: "name, type DateTime and format",
			Name:     "param",
			Type:     "DateTime",
			Format:   "format",
			ExpectedPlaceholders: map[string]*placeholder{
				"param": {"DateTime", "format"},
			},
		},
		{
			TestName:      "name, type DateTime and no format",
			Name:          "param",
			Type:          "DateTime",
			ExpectedError: "format is required for DateTime placeholders",
		},
		{
			TestName: "name, type num and format",
			Name:     "param",
			Type:     "num",
			Format:   "format",
			ExpectedPlaceholders: map[string]*placeholder{
				"param": {"num", "format"},
			},
		},
		{
			TestName: "name, type int and format",
			Name:     "param",
			Type:     "int",
			Format:   "format",
			ExpectedPlaceholders: map[string]*placeholder{
				"param": {"int", "format"},
			},
		},
		{
			TestName: "name, type double and format",
			Name:     "param",
			Type:     "double",
			Format:   "format",
			ExpectedPlaceholders: map[string]*placeholder{
				"param": {"double", "format"},
			},
		},
		{
			TestName: "name, type double and no format",
			Name:     "param",
			Type:     "double",
			ExpectedPlaceholders: map[string]*placeholder{
				"param": {"double", ""},
			},
		},
		{
			TestName: "name, type int and no format",
			Name:     "param",
			Type:     "int",
			ExpectedPlaceholders: map[string]*placeholder{
				"param": {"int", ""},
			},
		},
		{
			TestName: "name, type num and no format",
			Name:     "param",
			Type:     "num",
			ExpectedPlaceholders: map[string]*placeholder{
				"param": {"num", ""},
			},
		},
		{
			TestName:      "name, and invalid type",
			Name:          "param",
			Type:          "invalid",
			ExpectedError: "unknown placeholder type invalid. Supported types: String, Object, DateTime, num, int, double",
		},
		{
			TestName: "name count, but not in plural",
			Name:     "count",
			Type:     "String",
			ExpectedPlaceholders: map[string]*placeholder{
				"count": {"String", ""},
			},
		},
		//
		// Plural tests
		//
		{
			TestName:      "name count, type String in plural",
			Plural:        true,
			Name:          "count",
			Type:          "String",
			ExpectedError: "invalid count placeholder type. Supported types: num, int",
		},
		{
			TestName: "just name count in plural",
			Plural:   true,
			Name:     "count",
			ExpectedPlaceholders: map[string]*placeholder{
				"count": nil,
			},
		},
		{
			TestName: "name count, type num, without format in plural",
			Plural:   true,
			Name:     "count",
			Type:     "num",
			ExpectedPlaceholders: map[string]*placeholder{
				"count": {"num", ""},
			},
		},
		{
			TestName: "name count, type num, with format in plural",
			Plural:   true,
			Name:     "count",
			Type:     "num",
			Format:   "format",
			ExpectedPlaceholders: map[string]*placeholder{
				"count": {"num", "format"},
			},
		},
		{
			TestName: "name count, type int, without format in plural",
			Plural:   true,
			Name:     "count",
			Type:     "int",
			ExpectedPlaceholders: map[string]*placeholder{
				"count": {"int", ""},
			},
		},
		{
			TestName: "name count, type int, with format in plural",
			Plural:   true,
			Name:     "count",
			Type:     "int",
			Format:   "format",
			ExpectedPlaceholders: map[string]*placeholder{
				"count": {"int", "format"},
			},
		},
		{
			TestName: "just name but already seen",
			InitialPlaceholders: map[string]*placeholder{
				"param": nil,
			},
			Name: "param",
			ExpectedPlaceholders: map[string]*placeholder{
				"param": nil,
			},
		},
		{
			TestName: "just name but already defined",
			InitialPlaceholders: map[string]*placeholder{
				"param": {"String", ""},
			},
			Name: "param",
			ExpectedPlaceholders: map[string]*placeholder{
				"param": {"String", ""},
			},
		},
		{
			TestName: "name and type but already seen",
			InitialPlaceholders: map[string]*placeholder{
				"param": nil,
			},
			Name: "param",
			Type: "String",
			ExpectedPlaceholders: map[string]*placeholder{
				"param": {"String", ""},
			},
		},
		{
			TestName: "name and type but already defined",
			InitialPlaceholders: map[string]*placeholder{
				"param": {"String", ""},
			},
			Name:          "param",
			Type:          "String",
			ExpectedError: "placeholder type can only be defined once",
			ExpectedPlaceholders: map[string]*placeholder{
				"param": {"String", ""},
			},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.TestName, func(t *testing.T) {
			pc := newTranslationParser(testCase.Plural)
			for name, placeholder := range testCase.InitialPlaceholders {
				pc.namedParams.Set(name, placeholder)
			}

			err := pc.addPlaceholder(testCase.Name, testCase.Type, testCase.Format)

			if testCase.ExpectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.EqualError(t, err, testCase.ExpectedError)
			}

			for name, placeholder := range testCase.ExpectedPlaceholders {
				actual, ok := pc.namedParams.Get(name)
				assert.True(t, ok)
				assert.Equal(t, placeholder, actual)
			}
			assert.Equal(t, pc.namedParams.Len(), len(testCase.ExpectedPlaceholders))
		})
	}
}

func TestTranslationParserFallbackPlaceholderTypes(t *testing.T) {
	type testCase struct {
		TestName string
		Plural   bool
		Before   map[string]*placeholder
		After    map[string]*placeholder
	}

	cases := []testCase{
		{
			TestName: "no placeholders",
		},
		{
			TestName: "no placeholders, plural",
			Plural:   true,
			After: map[string]*placeholder{
				"count": {"", ""},
			},
		},
		{
			TestName: "some defined and some seen placeholders",
			Before: map[string]*placeholder{
				"param1": {"String", ""},
				"param2": nil,
				"param3": {"DateTime", "format"},
				"param4": nil,
				"count":  nil,
			},
			After: map[string]*placeholder{
				"param1": {"String", ""},
				"param2": {"String", ""},
				"param3": {"DateTime", "format"},
				"param4": {"String", ""},
				"count":  {"String", ""},
			},
		},
		{
			TestName: "some defined and some seen placeholders, plural",
			Plural:   true,
			Before: map[string]*placeholder{
				"param1": {"String", ""},
				"param2": nil,
				"param3": {"DateTime", "format"},
				"param4": nil,
				"count":  nil,
			},
			After: map[string]*placeholder{
				"param1": {"String", ""},
				"param2": {"String", ""},
				"param3": {"DateTime", "format"},
				"param4": {"String", ""},
				"count":  {"", ""},
			},
		},
		{
			TestName: "count defined, plural",
			Plural:   true,
			Before: map[string]*placeholder{
				"count": {"int", "format"},
			},
			After: map[string]*placeholder{
				"count": {"int", "format"},
			},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.TestName, func(t *testing.T) {
			pc := newTranslationParser(testCase.Plural)

			for name, placeholder := range testCase.Before {
				pc.namedParams.Set(name, placeholder)
			}

			pc.fallbackPlaceholderTypes()

			for name, placeholder := range testCase.After {
				actual, ok := pc.namedParams.Get(name)
				assert.True(t, ok)
				assert.Equal(t, placeholder, actual)
			}
			assert.Equal(t, pc.namedParams.Len(), len(testCase.After))
		})
	}
}

func TestTranslationParserErrors(t *testing.T) {
	var errs translationParserErrors

	assert.False(t, errs.HasErrors())

	errs.AddError("field one", errors.New("error one"))

	assert.True(t, errs.HasErrors())
	assert.ErrorContains(t, errs, "field one")
	assert.ErrorContains(t, errs, "error one")

	errs.AddError("field two", errors.New("error two"))
	errs.AddError("field two", errors.New("error three"))

	assert.True(t, errs.HasErrors())
	assert.ErrorContains(t, errs, "field one")
	assert.ErrorContains(t, errs, "error one")
	assert.ErrorContains(t, errs, "field two")
	assert.ErrorContains(t, errs, "error two")
	assert.ErrorContains(t, errs, "error three")
}
