package cmd

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewEnvVars(t *testing.T) {
	os.Setenv("POEDITOR_TOKEN", "test token")

	vars, err := newEnvVars()

	assert.NoError(t, err)
	assert.NotNil(t, vars)
	assert.Equal(t, "test token", vars.Token)
}
