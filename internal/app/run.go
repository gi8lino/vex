package app

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/gi8lino/vex/internal/flag"
	"github.com/gi8lino/vex/internal/formatter"
	"github.com/gi8lino/vex/internal/processor"
	"github.com/gi8lino/vex/internal/utils"

	"github.com/containeroo/tinyflags"
)

// Run parses flags, wires IO, dispatches processing and returns exit code + error.
func Run(
	version, commit string,
	args []string,
	stdout io.Writer,
	stdin io.Reader,
	lookupEnv func(string) (string, bool),
	setEnv func(string, string) error,
) error {
	flags, err := flag.ParseFlags(args, version, commit)
	if err != nil {
		if tinyflags.IsHelpRequested(err) || tinyflags.IsVersionRequested(err) {
			fmt.Fprintln(stdout, err) // nolint:errcheck
			return nil
		}
		return err
	}

	// Merge external vars (multiple files allowed).
	if len(flags.VarsFiles) > 0 {
		lookupEnv, err = utils.MergeVars(flags.VarsFiles, lookupEnv)
		if err != nil {
			return err
		}
	}

	pr := processor.Processor{
		Opts:      flags,
		Stdout:    bufio.NewWriterSize(stdout, 1<<20),
		Stdin:     bufio.NewReaderSize(stdin, 1<<20),
		Lookup:    lookupEnv,
		Setenv:    setEnv,
		Formatter: formatter.NewFormatter(flags.Colored),
	}

	if len(flags.Positional) == 0 {
		// stdin -> stdout
		if err := pr.ProcessStream("<stdin>", pr.Stdin, pr.Stdout); err != nil {
			return err
		}
		return nil
	}

	if flags.InPlace {
		for _, p := range flags.Positional {
			if err := pr.ProcessInPlace(p); err != nil {
				return err
			}
		}
		return nil
	}

	// concatenate files to stdout in order
	for _, p := range flags.Positional {
		f, err := os.Open(p)
		if err != nil {
			return err
		}
		br := bufio.NewReaderSize(f, 1<<20)
		if err := pr.ProcessStream(filepath.Base(p), br, pr.Stdout); err != nil {
			_ = f.Close()
			return err
		}
		f.Close() // nolint:errcheck
	}
	return nil
}
