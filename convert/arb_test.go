package convert

import (
	"testing"

	"github.com/stretchr/testify/assert"
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

func TestArbMessageAttributesIsEmpty(t *testing.T) {
	type testCase struct {
		Name       string
		Attributes ARBMessageAttributes
		Expected   bool
	}

	nonEmptyMap := orderedmap.New[string, *ARBPlaceholder]()
	nonEmptyMap.Set("foo", &ARBPlaceholder{Name: "foo"})

	testCases := []testCase{
		{
			"all empty",
			ARBMessageAttributes{},
			true,
		},
		{
			"empty placeholders",
			ARBMessageAttributes{
				Placeholders: orderedmap.New[string, *ARBPlaceholder](),
			},
			true,
		},
		{
			"non-empty description",
			ARBMessageAttributes{
				Description: "foo",
			},
			false,
		},
		{
			"non-empty placeholders",
			ARBMessageAttributes{
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
