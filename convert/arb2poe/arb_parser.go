package arb2poe

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/leancodepl/poe2arb/convert"
	"github.com/leancodepl/poe2arb/flutter"
	"github.com/pkg/errors"
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

func parseARB(r io.Reader) (locale flutter.Locale, messages []*convert.ARBMessage, err error) {
	var arb map[string]any
	err = json.NewDecoder(r).Decode(&arb)
	if err != nil {
		err = errors.Wrap(err, "failed to decode ARB")
		return flutter.Locale{}, nil, err
	}

	lang, ok := arb[convert.LocaleKey].(string)
	if !ok {
		err = errors.New("missing locale key")
		return flutter.Locale{}, nil, err
	}

	locale, err = flutter.ParseLocale(lang)
	if err != nil {
		err = errors.Wrap(err, fmt.Sprintf("parsing locale %s", lang))
		return flutter.Locale{}, nil, err
	}

	for key, value := range arb {
		if strings.HasPrefix(key, "@") {
			continue
		}

		var translation string
		if translation, ok = value.(string); !ok {
			err = errors.Errorf("invalid translation value for %s", key)
			return flutter.Locale{}, nil, err
		}

		message := &convert.ARBMessage{
			Name:        key,
			Translation: translation,
		}

		if attrs, ok := arb["@"+key].(map[string]any); ok {
			encoded, err := json.Marshal(attrs)
			if err != nil {
				return flutter.Locale{}, nil,
					errors.Wrap(err, fmt.Sprintf("failed to encode attributes for %s", key))
			}

			var attributes struct {
				Placeholders map[string]struct {
					Type   string `json:"type,omitempty"`
					Format string `json:"format,omitempty"`
				} `json:"placeholders,omitempty"`
			}
			err = json.Unmarshal(encoded, &attributes)
			if err != nil {
				return flutter.Locale{}, nil,
					errors.Wrap(err, fmt.Sprintf("failed to decode attributes for %s", key))
			}

			attrsOm := orderedmap.New[string, *convert.ARBPlaceholder]()
			for placeholderName, placeholder := range attributes.Placeholders {
				attrsOm.Set(placeholderName, &convert.ARBPlaceholder{
					Name:   placeholderName,
					Type:   placeholder.Type,
					Format: placeholder.Format,
				})
			}

			message.Attributes = &convert.ARBMessageAttributes{
				Placeholders: attrsOm,
			}
		}

		messages = append(messages, message)
	}

	return
}
