package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/leancodepl/poe2arb/convert"
	"github.com/leancodepl/poe2arb/poeditor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTermInPrefixScope(t *testing.T) {
	cases := []struct {
		term, prefix string
		want         bool
	}{
		{"foo", "", true},
		{"foo:bar", "", false},
		{"foo:bar", "foo", true},
		{"other:bar", "foo", false},
		{"bar", "foo", false},
		{"foo:bar:baz", "foo", true}, // colon only counts the first segment
	}

	for _, c := range cases {
		t.Run(c.term+"|"+c.prefix, func(t *testing.T) {
			assert.Equal(t, c.want, termInPrefixScope(c.term, c.prefix))
		})
	}
}

func TestTermsToDelete(t *testing.T) {
	t.Run("empty prefix only considers unprefixed remote terms", func(t *testing.T) {
		remote := []poeditor.Term{
			{Term: "appTitle"},
			{Term: "obsolete"},
			{Term: "other:foo"}, // belongs to a different package
		}
		local := map[string]struct{}{"appTitle": {}}

		got := termsToDelete(remote, local, "")

		require.Len(t, got, 1)
		assert.Equal(t, "obsolete", got[0].Term)
	})

	t.Run("non-empty prefix only considers matching remote terms", func(t *testing.T) {
		remote := []poeditor.Term{
			{Term: "app:foo"},
			{Term: "app:bar"},
			{Term: "other:zzz"},
			{Term: "noPrefix"},
		}
		local := map[string]struct{}{"app:foo": {}}

		got := termsToDelete(remote, local, "app")

		require.Len(t, got, 1)
		assert.Equal(t, "app:bar", got[0].Term)
	})

	t.Run("returns nothing when all remote terms are present locally", func(t *testing.T) {
		remote := []poeditor.Term{{Term: "x"}, {Term: "y"}}
		local := map[string]struct{}{"x": {}, "y": {}}

		got := termsToDelete(remote, local, "")
		assert.Empty(t, got)
	})

	t.Run("output is sorted by term name", func(t *testing.T) {
		remote := []poeditor.Term{{Term: "c"}, {Term: "a"}, {Term: "b"}}
		got := termsToDelete(remote, map[string]struct{}{}, "")
		require.Len(t, got, 3)
		assert.Equal(t, "a", got[0].Term)
		assert.Equal(t, "b", got[1].Term)
		assert.Equal(t, "c", got[2].Term)
	})
}

func TestReadLocalTermNames(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "app_en.arb")
	require.NoError(t, os.WriteFile(path, []byte(`{
		"@@locale": "en",
		"appTitle": "Hello",
		"@appTitle": {"description": "title"},
		"helloWorld": "world"
	}`), 0o644))

	t.Run("without prefix", func(t *testing.T) {
		got, err := readLocalTermNames(path, "")
		require.NoError(t, err)
		assert.Contains(t, got, "appTitle")
		assert.Contains(t, got, "helloWorld")
		assert.NotContains(t, got, "@appTitle")
		assert.NotContains(t, got, "@@locale")
	})

	t.Run("with prefix", func(t *testing.T) {
		got, err := readLocalTermNames(path, "myapp")
		require.NoError(t, err)
		assert.Contains(t, got, "myapp:appTitle")
		assert.Contains(t, got, "myapp:helloWorld")
	})
}

func TestFindARBFiles(t *testing.T) {
	dir := t.TempDir()

	mustWrite := func(name string) {
		require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte("{}"), 0o644))
	}
	mustWrite("app_en.arb")
	mustWrite("app_pl.arb")
	mustWrite("app_de.arb")
	mustWrite("not_an_arb.txt")
	mustWrite("other_en.arb") // wrong prefix

	files, template, err := findARBFiles(dir, "app_", "en")
	require.NoError(t, err)
	require.Len(t, files, 3)
	assert.Equal(t, filepath.Join(dir, "app_en.arb"), template)

	// Sorted
	assert.Equal(t, filepath.Join(dir, "app_de.arb"), files[0])
	assert.Equal(t, filepath.Join(dir, "app_en.arb"), files[1])
	assert.Equal(t, filepath.Join(dir, "app_pl.arb"), files[2])
}

func TestOrderTemplateFirst(t *testing.T) {
	files := []string{"a", "b", "c"}
	got := orderTemplateFirst(files, "b")
	assert.Equal(t, []string{"b", "a", "c"}, got)
}

func TestTermsToAdd(t *testing.T) {
	remote := []poeditor.Term{
		{Term: "existing"},
		{Term: "alsoExisting"},
	}
	local := map[string]struct{}{
		"existing":     {},
		"alsoExisting": {},
		"newTerm":      {},
		"anotherNew":   {},
	}

	got := termsToAdd(remote, local, nil, "")
	assert.Equal(t, []string{"anotherNew", "newTerm"}, got)
}

func ptr(s string) *string { return &s }

func TestEqualTranslation(t *testing.T) {
	t.Run("string equal", func(t *testing.T) {
		local := convert.POETermDefinition{Value: ptr("Hello")}
		remote := json.RawMessage(`"Hello"`)
		assert.True(t, equalTranslation(local, remote))
	})

	t.Run("string not equal", func(t *testing.T) {
		local := convert.POETermDefinition{Value: ptr("Hello")}
		remote := json.RawMessage(`"Hi"`)
		assert.False(t, equalTranslation(local, remote))
	})

	t.Run("nil and empty are treated as equal", func(t *testing.T) {
		local := convert.POETermDefinition{Value: ptr("")}
		remote := json.RawMessage(`null`)
		assert.True(t, equalTranslation(local, remote))
	})

	t.Run("plural shape mismatch is unequal", func(t *testing.T) {
		local := convert.POETermDefinition{Value: ptr("Hello")}
		remote := json.RawMessage(`{"one":"a","other":"b"}`)
		assert.False(t, equalTranslation(local, remote))
	})

	t.Run("plural equal regardless of key order", func(t *testing.T) {
		local := convert.POETermDefinition{
			IsPlural: true,
			Plural: &convert.POETermPluralDefinition{
				One:   ptr("1 apple"),
				Other: "{count} apples",
			},
		}
		remote := json.RawMessage(`{"other":"{count} apples","one":"1 apple"}`)
		assert.True(t, equalTranslation(local, remote))
	})

	t.Run("plural unequal in 'other' branch", func(t *testing.T) {
		local := convert.POETermDefinition{
			IsPlural: true,
			Plural: &convert.POETermPluralDefinition{
				One:   ptr("1 apple"),
				Other: "{count} apples",
			},
		}
		remote := json.RawMessage(`{"one":"1 apple","other":"{count} oranges"}`)
		assert.False(t, equalTranslation(local, remote))
	})
}

func TestShouldSkipLang(t *testing.T) {
	cases := []struct {
		lang      string
		overrides []string
		want      bool
	}{
		{"en", nil, false},
		{"en", []string{}, false},
		{"en", []string{"en", "pl"}, false},
		{"de", []string{"en", "pl"}, true},
		{"EN", []string{"en"}, false}, // case-insensitive
	}
	for _, c := range cases {
		assert.Equal(t, c.want, shouldSkipLang(c.lang, c.overrides), "lang=%s overrides=%v", c.lang, c.overrides)
	}
}

func TestLangInProject(t *testing.T) {
	avail := []poeditor.Language{{Code: "en"}, {Code: "pl"}}
	assert.True(t, langInProject("en", avail))
	assert.True(t, langInProject("PL", avail)) // case-insensitive
	assert.False(t, langInProject("de", avail))
}
