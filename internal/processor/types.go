package processor

import (
	"bufio"
	"os"
	"path/filepath"

	"github.com/gi8lino/vex/internal/flag"
	"github.com/gi8lino/vex/internal/formatter"
)

// Processor coordinates options, I/O streams, and env access.
type Processor struct {
	opts      flag.Options
	lookup    func(string) (string, bool)
	setenv    func(string, string) error
	formatter formatter.Formatter
}

// NewProcessor creates a Processor with the given options, env lookup, and formatter.
func NewProcessor(
	opts flag.Options,
	lookup func(string) (string, bool),
	setenv func(string, string) error,
	fmt formatter.Formatter,
	ioBufSize int,
) *Processor {
	return &Processor{
		opts:      opts,
		lookup:    lookup,
		setenv:    setenv,
		formatter: fmt,
	}
}

// ProcessFile opens a file and streams it into p.stdout using base name as label.
func (p *Processor) ProcessFile(path string, out *bufio.Writer, bufSize int) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	br := bufio.NewReaderSize(f, bufSize)
	return p.ProcessStream(filepath.Base(path), br, out)
}

// ProcessFiles processes multiple files to p.stdout in order.
func (p *Processor) ProcessFiles(paths []string, out *bufio.Writer, bufSize int) error {
	for _, path := range paths {
		if err := p.ProcessFile(path, out, bufSize); err != nil {
			return err
		}
	}
	return nil
}

// ProcessStdin streams p.stdin to p.stdout with a stable label.
func (p *Processor) ProcessStdin(in *bufio.Reader, out *bufio.Writer) error {
	return p.ProcessStream("<stdin>", in, out)
}
