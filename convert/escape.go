package convert

import (
	"errors"
	"strings"
)

// EscapeSegment describes one span of a translation string as classified by
// the ICU single-quote escape tokenizer. Escaped=true means the text is a
// matched-quote span (like two consecutive apostrophes, or 'literal') and its
// contents must NOT be scanned for placeholders.
type EscapeSegment struct {
	Text    string
	Escaped bool
}

// ErrUnmatchedQuote is returned by SplitByEscapes when an ICU escape span
// is opened but never closed.
var ErrUnmatchedQuote = errors.New("unmatched single quote in escaping sequence")

// SplitByEscapes tokenizes a translation string into escape and non-escape
// spans following Flutter's use-escaping semantics
// (packages/flutter_tools/lib/src/localizations/message_parser.dart):
//
//   - Two consecutive apostrophes are one escape span rendering a single
//     literal apostrophe.
//   - 'X', where X contains no apostrophe, is one escape span rendering X
//     literally (its contents are not scanned for placeholders).
//   - An unmatched apostrophe is a hard error, matching flutter_tools.
//
// Everything outside matched quotes is returned as non-escaped text.
func SplitByEscapes(s string) ([]EscapeSegment, error) {
	var out []EscapeSegment
	i := 0
	for i < len(s) {
		if s[i] != '\'' {
			j := i
			for j < len(s) && s[j] != '\'' {
				j++
			}
			out = append(out, EscapeSegment{Text: s[i:j]})
			i = j
			continue
		}

		close := strings.IndexByte(s[i+1:], '\'')
		if close == -1 {
			return nil, ErrUnmatchedQuote
		}
		end := i + 1 + close + 1
		out = append(out, EscapeSegment{Text: s[i:end], Escaped: true})
		i = end
	}
	return out, nil
}
