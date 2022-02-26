package utils_test

import (
	"testing"

	"github.com/leancodepl/poe2arb/utils"
	"github.com/stretchr/testify/assert"
)

func TestOrderedMap(t *testing.T) {
	m := utils.NewOrderedMap()

	m.Set("b", true)
	m.Set("a", true)
	m.Set("c", true)
	m.Set("d", true)

	m.Set("b", false)

	m.Remove("c")

	var keysIteratedOver []string
	m.ForEach(func(key string, value interface{}) {
		keysIteratedOver = append(keysIteratedOver, key)
	})

	json, err := m.MarshalJSON()

	if assert.NoError(t, err) {
		expected := []byte(`{"b":false,"a":true,"d":true}`)
		assert.Equal(t, expected, json)
	}

	assert.Equal(t, "b", keysIteratedOver[0])
	assert.Equal(t, "a", keysIteratedOver[1])
	assert.Equal(t, "d", keysIteratedOver[2])
}
