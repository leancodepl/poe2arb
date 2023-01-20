package main

import (
	"os"

	"github.com/leancodepl/poe2arb/cmd"
	"github.com/leancodepl/poe2arb/log"
)

func main() {
	logger := log.New(os.Stdout)

	cmd.Execute(logger)
}
