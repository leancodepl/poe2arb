// Package convert holds structures common to both directions of conversion.
package convert

import (
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

const LocaleKey = "@@locale"

type ARBMessage struct {
	Name        string
	Translation string
	Attributes  *ARBMessageAttributes
}

type ARBMessageAttributes struct {
	Description  string                                          `json:"description,omitempty"`
	Placeholders *orderedmap.OrderedMap[string, *ARBPlaceholder] `json:"placeholders,omitempty"`
}

func (a *ARBMessageAttributes) IsEmpty() bool {
	return a.Description == "" && (a.Placeholders == nil || a.Placeholders.Len() == 0)
}

type ARBPlaceholder struct {
	Name   string `json:"-"`
	Type   string `json:"type,omitempty"`
	Format string `json:"format,omitempty"`
}
