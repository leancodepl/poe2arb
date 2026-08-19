package convert

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSplitByEscapes(t *testing.T) {
	cases := []struct {
		name    string
		input   string
		want    []EscapeSegment
		wantErr error
	}{
		{
			name:  "empty",
			input: "",
			want:  nil,
		},
		{
			name:  "plain text",
			input: "hello world",
			want:  []EscapeSegment{{Text: "hello world"}},
		},
		{
			name:  "escaped braces",
			input: "'{terms}'",
			want:  []EscapeSegment{{Text: "'{terms}'", Escaped: true}},
		},
		{
			name:  "double apostrophe",
			input: "it''s",
			want: []EscapeSegment{
				{Text: "it"},
				{Text: "''", Escaped: true},
				{Text: "s"},
			},
		},
		{
			name:  "escape mixed with plain placeholder",
			input: "read the '{terms}' before {agreeing}",
			want: []EscapeSegment{
				{Text: "read the "},
				{Text: "'{terms}'", Escaped: true},
				{Text: " before {agreeing}"},
			},
		},
		{
			name:  "multiple escapes",
			input: "'{a}' and '{b}'",
			want: []EscapeSegment{
				{Text: "'{a}'", Escaped: true},
				{Text: " and "},
				{Text: "'{b}'", Escaped: true},
			},
		},
		{
			name:    "unmatched trailing apostrophe",
			input:   "hello '{world",
			wantErr: ErrUnmatchedQuote,
		},
		{
			name:    "single quote at end of string",
			input:   "hello'",
			wantErr: ErrUnmatchedQuote,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := SplitByEscapes(c.input)
			if c.wantErr != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, c.wantErr), "want %v, got %v", c.wantErr, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, c.want, got)
		})
	}
}
