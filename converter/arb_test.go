package converter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

func TestArbMessageAttributesIsEmpty(t *testing.T) {
	type testCase struct {
		Name       string
		Attributes arbMessageAttributes
		Expected   bool
	}

	nonEmptyMap := orderedmap.New[string, *arbPlaceholder]()
	nonEmptyMap.Set("foo", &arbPlaceholder{Name: "foo"})

	testCases := []testCase{
		{
			"all empty",
			arbMessageAttributes{},
			true,
		},
		{
			"empty placeholders",
			arbMessageAttributes{
				Placeholders: orderedmap.New[string, *arbPlaceholder](),
			},
			true,
		},
		{
			"non-empty description",
			arbMessageAttributes{
				Description: "foo",
			},
			false,
		},
		{
			"non-empty placeholders",
			arbMessageAttributes{
				Placeholders: nonEmptyMap,
			},
			false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			assert.Equal(t, tc.Expected, tc.Attributes.IsEmpty())
		})
	}
}
