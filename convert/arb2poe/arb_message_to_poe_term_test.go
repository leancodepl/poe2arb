package arb2poe

import (
	"testing"

	"github.com/leancodepl/poe2arb/convert"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

func TestIndexPlaceholder(t *testing.T) {
	t.Run("without escaping: finds any occurrence including inside quotes", func(t *testing.T) {
		got, err := indexPlaceholder("read '{name}' Hello {name}", "name", false)
		require.NoError(t, err)
		assert.Equal(t, 6, got) // inside the quotes
	})

	t.Run("with escaping: skips occurrence inside quotes", func(t *testing.T) {
		got, err := indexPlaceholder("read '{name}' Hello {name}", "name", true)
		require.NoError(t, err)
		assert.Equal(t, 20, got) // the real placeholder outside quotes
	})

	t.Run("with escaping: unmatched quote propagates the error", func(t *testing.T) {
		_, err := indexPlaceholder("'{oops", "oops", true)
		assert.ErrorIs(t, err, convert.ErrUnmatchedQuote)
	})

	t.Run("with escaping: only escaped occurrence returns -1", func(t *testing.T) {
		got, err := indexPlaceholder("only '{name}' here", "name", true)
		require.NoError(t, err)
		assert.Equal(t, -1, got)
	})
}

func TestArbMessageToPOETermWithEscaping(t *testing.T) {
	placeholders := orderedmap.New[string, *convert.ARBPlaceholder]()
	placeholders.Set("name", &convert.ARBPlaceholder{Name: "name", Type: "String"})

	m := &convert.ARBMessage{
		Name:        "greeting",
		Translation: "read '{name}' Hello {name}",
		Attributes:  &convert.ARBMessageAttributes{Placeholders: placeholders},
	}

	term, err := arbMessageToPOETerm(m, false, "", true)
	require.NoError(t, err)
	// Only the unescaped {name} should be annotated with the ,String type.
	require.NotNil(t, term.Definition.Value)
	assert.Equal(t, "read '{name}' Hello {name,String}", *term.Definition.Value)
}

func TestArbMessageToPOETermCountPlural(t *testing.T) {
	str := func(s string) *string { return &s }

	type testCase struct {
		Name         string
		Translation  string
		Placeholders map[string]*convert.ARBPlaceholder
		UseEscaping  bool

		ExpectedPlural *convert.POETermPluralDefinition
		ExpectedValue  string
		ExpectedError  string
	}

	cases := []testCase{
		{
			Name:           "whole message",
			Translation:    "{count, plural, one {an apple} other {{count} apples}}",
			ExpectedPlural: &convert.POETermPluralDefinition{One: str("an apple"), Other: "{count} apples"},
		},
		{
			// https://github.com/leancodepl/poe2arb/issues/94
			Name:           "embedded in text",
			Translation:    "I have {count, plural, one {an apple} other {{count} apples}}",
			ExpectedPlural: &convert.POETermPluralDefinition{One: str("I have an apple"), Other: "I have {count} apples"},
		},
		{
			Name:        "text and placeholders on both sides",
			Translation: "{user} has {count, plural, =0 {no messages} one {one message} other {{count} messages}} since {date}.",
			Placeholders: map[string]*convert.ARBPlaceholder{
				"user":  {Type: "String"},
				"count": {Type: "int"},
				"date":  {Type: "DateTime", Format: "yMd"},
			},
			ExpectedPlural: &convert.POETermPluralDefinition{
				// each placeholder is defined only once, poe2arb rejects redefinitions
				Zero:  str("{user,String} has no messages since {date,DateTime,yMd}."),
				One:   str("{user} has one message since {date}."),
				Other: "{user} has {count,int} messages since {date}.",
			},
		},
		{
			Name:          "plural with other placeholder than count",
			Translation:   "I have {n, plural, one {an apple} other {{n} apples}}",
			ExpectedValue: "I have {n, plural, one {an apple} other {{n} apples}}",
		},
		{
			Name:          "not closed",
			Translation:   "I have {count, plural, one {an apple} other {{count} apples}",
			ExpectedValue: "I have {count, plural, one {an apple} other {{count} apples}",
		},
		{
			Name:          "two count plurals",
			Translation:   "{count, plural, one {a} other {b}} and {count, plural, one {c} other {d}}",
			ExpectedError: "POEditor plural terms support only one count plural per message",
		},
		{
			Name:          "escaped plural",
			Translation:   "Type '{count, plural, other {x}}' to {count}",
			UseEscaping:   true,
			ExpectedValue: "Type '{count, plural, other {x}}' to {count}",
		},
		{
			Name:           "escaped braces inside a case",
			Translation:    "Items: {count, plural, one {one '{'} other {many '}'}}",
			UseEscaping:    true,
			ExpectedPlural: &convert.POETermPluralDefinition{One: str("Items: one '{'"), Other: "Items: many '}'"},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			m := &convert.ARBMessage{Name: "apples", Translation: c.Translation}
			if c.Placeholders != nil {
				placeholders := orderedmap.New[string, *convert.ARBPlaceholder]()
				for _, name := range []string{"user", "count", "date"} {
					if p, ok := c.Placeholders[name]; ok {
						placeholders.Set(name, p)
					}
				}
				m.Attributes = &convert.ARBMessageAttributes{Placeholders: placeholders}
			}

			term, err := arbMessageToPOETerm(m, false, "", c.UseEscaping)
			if c.ExpectedError != "" {
				assert.EqualError(t, err, c.ExpectedError)
				return
			}
			require.NoError(t, err)

			if c.ExpectedPlural != nil {
				assert.True(t, term.Definition.IsPlural)
				assert.Equal(t, ".", term.TermPlural)
				assert.Equal(t, c.ExpectedPlural, term.Definition.Plural)
			} else {
				assert.False(t, term.Definition.IsPlural)
				require.NotNil(t, term.Definition.Value)
				assert.Equal(t, c.ExpectedValue, *term.Definition.Value)
			}
		})
	}
}
