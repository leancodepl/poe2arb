package flutter_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/leancodepl/poe2arb/flutter"
	"github.com/stretchr/testify/assert"
)

func TestNewFromDirectory(t *testing.T) {
	t.Run("without pubspec.yaml anywhere", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "")
		assert.NoError(t, err)
		defer os.RemoveAll(dir)

		cfg, err := flutter.NewFromDirectory(dir)

		assert.NoError(t, err)
		assert.Nil(t, cfg)
	})

	t.Run("with pubspec.yaml, without l10n.yaml", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "")
		assert.NoError(t, err)
		defer os.RemoveAll(dir)

		err = os.WriteFile(filepath.Join(dir, "pubspec.yaml"), []byte{}, 0o666)
		assert.NoError(t, err)

		cfg, err := flutter.NewFromDirectory(dir)

		assert.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.Equal(t, dir, cfg.RootDir)
	})

	t.Run("with pubspec.yaml in parent dir, without l10n.yaml", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "")
		assert.NoError(t, err)
		defer os.RemoveAll(dir)

		err = os.WriteFile(filepath.Join(dir, "pubspec.yaml"), []byte{}, 0o666)
		assert.NoError(t, err)

		childDir := filepath.Join(dir, "child-dir")
		err = os.Mkdir(childDir, 0o777)
		assert.NoError(t, err)

		cfg, err := flutter.NewFromDirectory(childDir)

		assert.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.Equal(t, dir, cfg.RootDir)
	})

	t.Run("with pubspec.yaml, with l10n.yaml", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "")
		assert.NoError(t, err)
		defer os.RemoveAll(dir)

		err = os.WriteFile(filepath.Join(dir, "pubspec.yaml"), []byte{}, 0o666)
		assert.NoError(t, err)

		l10nContents := `arb-dir: this-is/arb-dir/test
poeditor-project-id: 123123`
		err = os.WriteFile(filepath.Join(dir, "l10n.yaml"), []byte(l10nContents), 0o666)
		assert.NoError(t, err)

		cfg, err := flutter.NewFromDirectory(dir)

		assert.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.Equal(t, dir, cfg.RootDir)

		assert.Equal(t, "this-is/arb-dir/test", cfg.L10n.ARBDir)
		assert.Equal(t, "123123", cfg.L10n.POEditorProjectID)
	})
}
