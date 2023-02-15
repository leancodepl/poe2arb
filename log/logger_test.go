package log_test

import (
	"bytes"
	"testing"

	"github.com/leancodepl/poe2arb/log"
	"github.com/stretchr/testify/assert"
)

const (
	blue  = "\x1b[34m"
	green = "\x1b[32m"
	red   = "\x1b[31m"
	reset = "\x1b[0m"
)

func TestLoggerInfo(t *testing.T) {
	var buf bytes.Buffer

	l := log.New(&buf)

	l.Info("Hello, %s!", "world")

	assert.Equal(t, blue+" • "+reset+"Hello, world!\n", buf.String())
}

func TestLoggerSuccess(t *testing.T) {
	var buf bytes.Buffer

	l := log.New(&buf)

	l.Success("Hello, %s!", "world")

	assert.Equal(t, green+" • "+reset+"Hello, world!\n", buf.String())
}

func TestLoggerError(t *testing.T) {
	var buf bytes.Buffer

	l := log.New(&buf)

	l.Error("Hello, %s!", "world")

	assert.Equal(t, red+" • "+reset+"Hello, world!\n", buf.String())
}

func TestLoggerSub(t *testing.T) {
	var buf bytes.Buffer

	l := log.New(&buf)

	sub := l.Sub()
	sub.Info("test")

	subSub := sub.Sub()
	subSub.Info("test2")

	assert.Equal(t, "  "+blue+" • "+reset+"test\n    "+blue+" • "+reset+"test2\n", buf.String())
}

func TestLoggerMultilineInfo(t *testing.T) {
	var buf bytes.Buffer

	l := log.New(&buf)

	l.Info("test one line\ntest second line")

	assert.Equal(t, blue+" • "+reset+"test one line\n"+blue+" • "+reset+"test second line\n", buf.String())
}
