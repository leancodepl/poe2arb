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
