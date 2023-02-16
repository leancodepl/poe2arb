// Package log provides beautiful logger for the console interface.
package log

import (
	"fmt"
	"io"
	"strings"

	clr "github.com/TwiN/go-color"
)

type Logger struct {
	writer io.Writer
	depth  int
}

func New(writer io.Writer) *Logger {
	return &Logger{
		writer: writer,
		depth:  0,
	}
}

func (l *Logger) Info(msg string, params ...any) *Logger {
	l.log(clr.Blue, msg, params...)

	return l
}

func (l *Logger) Success(msg string, params ...any) *Logger {
	l.log(clr.Green, msg, params...)

	return l
}

func (l *Logger) Error(msg string, params ...any) *Logger {
	l.log(clr.Red, msg, params...)

	return l
}

func (l *Logger) log(color, msg string, params ...any) {
	prefix := strings.Repeat("  ", l.depth) + color + " • " + clr.Reset

	lines := strings.Split(msg, "\n")
	for i, line := range lines {
		lines[i] = prefix + line + "\n"
	}

	msg = strings.Join(lines, "")

	str := fmt.Sprintf(msg, params...)
	fmt.Fprint(l.writer, str)
}

func (l *Logger) Sub() *Logger {
	return &Logger{
		writer: l.writer,
		depth:  l.depth + 1,
	}
}
