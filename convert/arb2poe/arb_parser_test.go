package arb2poe

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseARB(t *testing.T) {
	file, _ := os.Open("testdata/english.arb")
	defer file.Close()

	lang, messages, err := parseARB(file)

	assert.Equal(t, "en", lang)
	assert.NotEmpty(t, messages)
	assert.NoError(t, err)
}
