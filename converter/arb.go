package converter

const localeKey = "@@locale"

type arbMessage struct {
	Name        string
	Translation string
	Attributes  *arbMessageAttributes
}

type arbMessageAttributes struct {
	Description  string                     `json:"description,omitempty"`
	Placeholders map[string]*arbPlaceholder `json:"placeholders,omitempty"`
}

type arbPlaceholder struct {
	Name   string `json:"-"`
	Type   string `json:"type,omitempty"`
	Format string `json:"format,omitempty"`
}
