package main

import (
	"log"
	"os"

	"github.com/Syncano/orion/cmd"
)

func main() {
	logger := log.New(os.Stderr, "", 0)

	// Run main cmd.App.
	defer func() {
		if r := recover(); r != nil {
			logger.Panicf("Panic occurred! %q", r)
		}
	}()
	if err := cmd.App.Run(os.Args); err != nil {
		logger.Panicf("Fatal error! %q", err)
	}
}
