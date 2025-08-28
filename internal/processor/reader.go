package processor

import (
	"bufio"
	"io"

	"github.com/gi8lino/vex/internal/flag"
	"github.com/gi8lino/vex/internal/formatter"
	"github.com/gi8lino/vex/internal/fsm"
)

// Processor coordinates options, standard I/O streams, and env access.
// It builds an FSM engine and drives substitution on provided streams.
type Processor struct {
	Opts      flag.Options                // parsed CLI options
	Stdout    *bufio.Writer               // default stdout writer
	Stdin     *bufio.Reader               // default stdin reader
	Lookup    func(string) (string, bool) // env lookup (name → value, ok)
	Setenv    func(string, string) error  // env setter for :=, = operators
	Formatter formatter.Formatter         // formatter for colored formatter
}

// ProcessStream runs the FSM on the given reader and writer.
// It instantiates an Engine with configured options and environment hooks.
func (p *Processor) ProcessStream(label string, r io.Reader, w *bufio.Writer) error {
	eng := &fsm.Engine{
		Label:  label,
		Opts:   p.Opts,
		Lookup: p.Lookup,
		Setenv: p.Setenv,
		Format: p.Formatter,
	}
	if err := eng.Consume(r, w); err != nil {
		return err
	}
	return w.Flush()
}
