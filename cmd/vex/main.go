package main

import (
	"fmt"
	"os"

	"github.com/gi8lino/vex/internal/app"
)

var (
	Version = "dev"
	Commit  = "none"
)

// main sets up the application context and runs the main loop.
func main() {
	if err := app.Run(
		Version,
		Commit,
		os.Args[1:],
		os.Stdout,
		os.Stdin,
		os.LookupEnv,
		os.Setenv,
	); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
