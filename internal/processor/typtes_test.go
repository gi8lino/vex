package processor

import (
	"bufio"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gi8lino/vex/internal/flag"
	"github.com/gi8lino/vex/internal/formatter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProcessFile(t *testing.T) {
	t.Parallel()

	t.Run("expands file content and flushes", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		path := filepath.Join(dir, "greet.txt")
		require.NoError(t, os.WriteFile(path, []byte("hi ${NAME}"), 0o600))

		p := NewProcessor(
			flag.Options{Colored: false},
			func(name string) (string, bool) {
				if name == "NAME" {
					return "Ada", true
				}
				return "", false
			},
			nil,
			formatter.NewFormatter(false),
			testBufSize,
		)

		var out bytes.Buffer
		w := bufio.NewWriterSize(&out, testBufSize)

		err := p.ProcessFile(path, w, testBufSize)
		require.NoError(t, err)
		assert.Equal(t, "hi Ada", out.String())
	})

	t.Run("propagates open error", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		path := filepath.Join(dir, "missing.txt")

		p := NewProcessor(
			flag.Options{},
			nil,
			nil,
			formatter.NewFormatter(false),
			testBufSize,
		)

		var out bytes.Buffer
		w := bufio.NewWriterSize(&out, testBufSize)

		err := p.ProcessFile(path, w, testBufSize)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "open "+path+": ")
		assert.Equal(t, "", out.String())
	})

	t.Run("error during processing is returned", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		path := filepath.Join(dir, "boom.txt")
		require.NoError(t, os.WriteFile(path, []byte("${VAR?boom}"), 0o600))

		p := NewProcessor(
			flag.Options{Colored: false},
			func(string) (string, bool) { return "", false }, // unset -> error
			nil,
			formatter.NewFormatter(false),
			testBufSize,
		)

		var out bytes.Buffer
		w := bufio.NewWriterSize(&out, testBufSize)

		err := p.ProcessFile(path, w, testBufSize)
		require.Error(t, err)
		assert.EqualError(t, err, "VAR: boom")
		assert.Equal(t, "", out.String()) // no flush on error path
	})
}

func TestProcessFiles(t *testing.T) {
	t.Parallel()

	t.Run("concatenates multiple files in order", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		a := filepath.Join(dir, "a.txt")
		b := filepath.Join(dir, "b.txt")

		require.NoError(t, os.WriteFile(a, []byte("A=${A};"), 0o600))
		require.NoError(t, os.WriteFile(b, []byte("B=${B}"), 0o600))

		p := NewProcessor(
			flag.Options{Colored: false},
			func(name string) (string, bool) {
				switch name {
				case "A":
					return "aa", true
				case "B":
					return "bb", true
				default:
					return "", false
				}
			},
			nil,
			formatter.NewFormatter(false),
			testBufSize,
		)

		var out bytes.Buffer
		w := bufio.NewWriterSize(&out, testBufSize)

		err := p.ProcessFiles([]string{a, b}, w, testBufSize)
		require.NoError(t, err)
		assert.Equal(t, "A=aa;B=bb", out.String())
	})

	t.Run("stops at first failing file and returns error; partial output kept", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		ok := filepath.Join(dir, "ok.txt")
		bad := filepath.Join(dir, "bad.txt")
		require.NoError(t, os.WriteFile(ok, []byte("X=${X};"), 0o600))
		require.NoError(t, os.WriteFile(bad, []byte("${VAR?boom}"), 0o600))

		p := NewProcessor(
			flag.Options{Colored: false},
			func(name string) (string, bool) {
				if name == "X" {
					return "x", true
				}
				return "", false
			},
			nil,
			formatter.NewFormatter(false),
			testBufSize,
		)

		var out bytes.Buffer
		w := bufio.NewWriterSize(&out, testBufSize)

		err := p.ProcessFiles([]string{ok, bad}, w, testBufSize)
		require.Error(t, err)
		assert.EqualError(t, err, "VAR: boom")
		// Data from the first file was flushed before the error in the second.
		assert.Equal(t, "X=x;", out.String())
	})
}

func TestProcessStdin(t *testing.T) {
	t.Parallel()

	t.Run("expands from stdin and flushes", func(t *testing.T) {
		t.Parallel()

		p := NewProcessor(
			flag.Options{Colored: false},
			func(string) (string, bool) { return "", false }, // unset -> default path
			nil,
			formatter.NewFormatter(false),
			testBufSize,
		)

		var out bytes.Buffer
		w := bufio.NewWriterSize(&out, testBufSize)
		r := bufio.NewReaderSize(strings.NewReader("v=${V:-def}"), testBufSize)

		err := p.ProcessStdin(r, w)
		require.NoError(t, err)
		assert.Equal(t, "v=def", out.String())
	})
}
