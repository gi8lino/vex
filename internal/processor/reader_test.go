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

func TestProcessStream(t *testing.T) {
	t.Parallel()

	t.Run("expands with lookup and flushes", func(t *testing.T) {
		t.Parallel()

		p := &Processor{
			Opts:      flag.Options{Colored: false},
			Formatter: formatter.NewFormatter(false),
			Lookup: func(name string) (string, bool) {
				if name == "NAME" {
					return "Ada", true
				}
				return "", false
			},
		}

		var out bytes.Buffer
		w := bufio.NewWriter(&out)
		r := strings.NewReader("hello ${NAME}")

		err := p.ProcessStream("test.txt", r, w)
		require.NoError(t, err)
		assert.Equal(t, "hello Ada", out.String())
	})

	t.Run("propagates error from engine", func(t *testing.T) {
		t.Parallel()

		p := &Processor{
			Formatter: formatter.NewFormatter(false),
			Opts:      flag.Options{Colored: false},
			Lookup: func(name string) (string, bool) {
				// unset so ${VAR?boom} errors
				return "", false
			},
		}

		var out bytes.Buffer
		w := bufio.NewWriter(&out)
		r := strings.NewReader("${VAR?boom}")

		err := p.ProcessStream("EngineLabel", r, w)
		require.Error(t, err)
		assert.EqualError(t, err, "VAR: boom")
		// no flush on error path; formatter may be empty
	})

	t.Run("calls Setenv on assign null and writes result", func(t *testing.T) {
		t.Parallel()

		var calls int
		var gotName, gotVal string
		p := &Processor{
			Formatter: formatter.NewFormatter(false),
			Opts:      flag.Options{Colored: false},
			Lookup: func(name string) (string, bool) {
				// set but empty -> triggers := assignment
				return "", true
			},
			Setenv: func(name, val string) error {
				calls++
				gotName, gotVal = name, val
				return nil
			},
		}

		var out bytes.Buffer
		w := bufio.NewWriter(&out)
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

		p := &Processor{
			Formatter: formatter.NewFormatter(false),
			Opts: flag.Options{
				NoOps:   true,
				Colored: false,
			},
			Lookup: func(name string) (string, bool) {
				return "", false
			},
		}

		input := "${NAME:-x}"
		var out bytes.Buffer
		w := bufio.NewWriter(&out)
		r := strings.NewReader(input)

		err := p.ProcessStream("label", r, w)
		require.NoError(t, err)
		assert.Equal(t, input, out.String())
	})
}
