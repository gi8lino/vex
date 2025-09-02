package app

import (
	"bufio"
	"fmt"
	"io"

	"github.com/gi8lino/vex/internal/flag"
	"github.com/gi8lino/vex/internal/formatter"
	"github.com/gi8lino/vex/internal/processor"
	"github.com/gi8lino/vex/internal/utils"

	"github.com/containeroo/tinyflags"
)

const ioBufSize = 1 << 20 // 1 MiB

// Run parses flags, wires IO, dispatches processing and returns error.
func Run(
	version, commit string,
	args []string,
	out io.Writer,
	in io.Reader,
	lookupEnv func(string) (string, bool),
	setEnv func(string, string) error,
) error {
	flags, err := flag.ParseFlags(args, version, commit)
	if err != nil {
		// Help/version are represented as errors by tinyflags.
		if tinyflags.IsHelpRequested(err) || tinyflags.IsVersionRequested(err) {
			// Print directly to the provided writer (unbuffered).
			_, _ = fmt.Fprintln(out, err)
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

	// Instantiate processor.
	pr := processor.NewProcessor(
		flags,
		lookupEnv,
		setEnv,
		formatter.NewFormatter(flags.Colored),
		ioBufSize,
	)

	// Prepare buffered writer once; only used in code paths that write to stdout.
	bw := bufio.NewWriterSize(out, ioBufSize)

	// No positional args: stream stdin -> stdout.
	if len(flags.Positional) == 0 {
		br := bufio.NewReaderSize(in, ioBufSize)
		if err := pr.ProcessStdin(br, bw); err != nil {
			return err
		}
		return bw.Flush()
	}

	// In-place editing for positional files.
	if flags.InPlace {
		for _, p := range flags.Positional {
			if err := pr.ProcessInPlace(p, ioBufSize); err != nil {
				return err
			}
		}
		// Nothing buffered to flush in this branch (writes go to files).
		return nil
	}

	// Positional files -> stdout (concatenate in order).
	if err := pr.ProcessFiles(flags.Positional, bw, ioBufSize); err != nil {
		return err
	}

	_ = bw.Flush()

	return nil
}
