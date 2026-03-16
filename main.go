package main

import (
	"github.com/prokizzle/refactorkit/cmd"
	"github.com/prokizzle/refactorkit/internal/logging"
)

func main() {
	defer logging.RecoverPanic("main", func() {
		logging.ErrorPersist("Application terminated due to unhandled panic")
	})

	cmd.Execute()
}
