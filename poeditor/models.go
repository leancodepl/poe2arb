package poeditor

import "strings"

type Locale string

func NewLocale(lang string) Locale {
	return Locale(lang)
}

func (l Locale) String() string {
	return string(l)
}

func (l Locale) StringUnderscores() string {
	return strings.ReplaceAll(string(l), "-", "_")
}

type Language struct {
	Name string
	Code Locale
}
