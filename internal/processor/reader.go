package processor

import (
	"bufio"
	"io"

	"github.com/gi8lino/vex/internal/fsm"
)

// ProcessStream runs the FSM on the given reader and writer and flushes the writer.
func (p *Processor) ProcessStream(label string, r io.Reader, w *bufio.Writer) error {
	eng := &fsm.Engine{
		Label:  label,
		Opts:   p.opts,
		Lookup: p.lookup,
		Setenv: p.setenv,
		Format: p.formatter,
	}
	if err := eng.Consume(r, w); err != nil {
		return err
	}
	return w.Flush()
}
