package main

import (
	"github.com/leancodepl/poe2arb/cmd"
	"github.com/leancodepl/poe2arb/log"
)

func main() {
	logger := log.NewStdout()

	cmd.Execute(logger)
}
