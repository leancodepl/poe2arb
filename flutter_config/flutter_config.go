package flutter_config

import (
	"os"
	"path"
	"path/filepath"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

type FlutterConfig struct {
	RootDir string
	L10n    *L10n
}

// https://github.com/flutter/flutter/blob/61a0add2865c51bfee33939c1820709d1115c77d/packages/flutter_tools/lib/src/localizations/localizations_utils.dart#L291
type L10n struct {
	ARBDir                    string `yaml:"arb-dir"`
	TemplateArbFile           string `yaml:"template-arb-file"`
	RequireResourceAttributes bool   `yaml:"required-resource-attributes"`
}

func newDefaultL10n() *L10n {
	return &L10n{
		ARBDir:                    "lib/l10n",
		TemplateArbFile:           "app_en.arb",
		RequireResourceAttributes: false,
	}
}

func NewFromDirectory(dir string) (*FlutterConfig, error) {
	pubspec, err := walkUpForPubspec(dir)
	if err != nil {
		return nil, err
	}
	if pubspec == nil {
		// no pubspec found
		return nil, nil
	}

	rootDir := filepath.Dir(pubspec.Name())

	l10nFile, err := getL10nFile(rootDir)
	if err != nil {
		return nil, err
	}

	l10n := newDefaultL10n()
	err = yaml.NewDecoder(l10nFile).Decode(&l10n)
	if err != nil {
		return nil, errors.Wrap(err, "failure decoding l10n.yaml")
	}

	return &FlutterConfig{
		RootDir: rootDir,
		L10n:    l10n,
	}, nil
}

func walkUpForPubspec(dir string) (file *os.File, err error) {
	for {
		if file, err = os.Open(path.Join(dir, "pubspec.yaml")); err == nil {
			return file, nil
		}

		if !errors.Is(err, os.ErrNotExist) {
			return nil, errors.Wrap(err, "failure searching for pubspec.yaml")
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// we tried to up the root directory
			// could not find Dart project root (pubspec.yaml)
			return nil, nil
		}

		dir = parent
	}
}

func getL10nFile(pubspecDir string) (*os.File, error) {
	file, err := os.Open(path.Join(pubspecDir, "l10n.yaml"))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, errors.Wrap(err, "failure reading l10n.yaml")
	}

	return file, nil
}
