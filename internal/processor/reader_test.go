package processor

import (
	"bufio"
	"bytes"
	"strings"
	"testing"

	"github.com/gi8lino/vex/internal/flag"
	"github.com/gi8lino/vex/internal/formatter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testBufSize = 64 << 10 // 64 KiB is plenty for tests

func TestProcessStream(t *testing.T) {
	t.Parallel()

	t.Run("expands with lookup and flushes", func(t *testing.T) {
		t.Parallel()

		p := NewProcessor(
			flag.Options{Colored: false},
			func(name string) (string, bool) {
				if name == "NAME" {
					return "Ada", true
				}
				return "", false
			},
			nil, // no Setenv needed
			formatter.NewFormatter(false),
			testBufSize,
		)

		var out bytes.Buffer
		w := bufio.NewWriterSize(&out, testBufSize)
		r := strings.NewReader("hello ${NAME}")

		err := p.ProcessStream("test.txt", r, w)
		require.NoError(t, err)
		assert.Equal(t, "hello Ada", out.String())
	})

	t.Run("propagates error from engine", func(t *testing.T) {
		t.Parallel()

		p := NewProcessor(
			flag.Options{Colored: false},
			func(name string) (string, bool) {
				// unset so ${VAR?boom} errors
				return "", false
			},
			nil,
			formatter.NewFormatter(false),
			testBufSize,
		)

		var out bytes.Buffer
		w := bufio.NewWriterSize(&out, testBufSize)
		r := strings.NewReader("${VAR?boom}")

		err := p.ProcessStream("EngineLabel", r, w)
		require.Error(t, err)
		// exact message depends on your xerr formatting; keep if stable:
		assert.EqualError(t, err, "VAR: boom")
		// no flush on error path; output should be empty
		assert.Equal(t, "", out.String())
	})

	t.Run("calls Setenv on assign null and writes result", func(t *testing.T) {
		t.Parallel()

		var calls int
		var gotName, gotVal string

		p := NewProcessor(
			flag.Options{Colored: false},
			func(name string) (string, bool) {
				// set but empty -> triggers := assignment
				return "", true
			},
			func(name, val string) error {
				calls++
				gotName, gotVal = name, val
				return nil
			},
			formatter.NewFormatter(false),
			testBufSize,
		)

		var out bytes.Buffer
		w := bufio.NewWriterSize(&out, testBufSize)
		r := strings.NewReader("${NEW:=value}")

		err := p.ProcessStream("label", r, w)
		require.NoError(t, err)
		assert.Equal(t, "value", out.String())
		assert.Equal(t, 1, calls)
		assert.Equal(t, "NEW", gotName)
		assert.Equal(t, "value", gotVal)
	})

	t.Run("respects NoOps and keeps literal", func(t *testing.T) {
		t.Parallel()

		p := NewProcessor(
			flag.Options{
				NoOps:   true,
				Colored: false,
			},
			func(name string) (string, bool) {
				return "", false
			},
			nil,
			formatter.NewFormatter(false),
			testBufSize,
		)

		input := "${NAME:-x}"
		var out bytes.Buffer
		w := bufio.NewWriterSize(&out, testBufSize)
		r := strings.NewReader(input)

		err := p.ProcessStream("label", r, w)
		require.NoError(t, err)
		assert.Equal(t, input, out.String())
	})
}
