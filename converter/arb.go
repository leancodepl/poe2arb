package converter

import "github.com/leancodepl/poe2arb/utils"

const localeKey = "@@locale"

type arbMessage struct {
	Name        string
	Translation string
	Attributes  *arbMessageAttributes
}

type arbMessageAttributes struct {
	Description  string            `json:"description,omitempty"`
	Placeholders *utils.OrderedMap `json:"placeholders,omitempty"` // value type *arbPlaceholder
}

type arbPlaceholder struct {
	Name   string `json:"-"`
	Type   string `json:"type,omitempty"`
	Format string `json:"format,omitempty"`
}
