package fsm

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

// helper to run the FSM on a string and capture formatter.
func runFSM(t *testing.T, e *Engine, in string) (string, error) {
	t.Helper()
	var out bytes.Buffer
	w := bufio.NewWriter(&out)
	err := e.Consume(strings.NewReader(in), w)
	w.Flush() // nolint:errcheck
	return out.String(), err
}

func TestEngineFSM(t *testing.T) {
	t.Parallel()
	const label = "EngineLabel"

	t.Run("empty input emits nothing and flushes", func(t *testing.T) {
		t.Parallel()
		e := &Engine{
			Label:  label,
			Opts:   flag.Options{},
			Format: formatter.NewFormatter(false),
		}
		got, err := runFSM(t, e, "")
		require.NoError(t, err)
		assert.Equal(t, "", got)
	})

	t.Run("text passthrough", func(t *testing.T) {
		t.Parallel()
		e := &Engine{
			Format: formatter.NewFormatter(false),
			Label:  label,
		}
		got, err := runFSM(t, e, "hello world")
		require.NoError(t, err)
		assert.Equal(t, "hello world", got)
	})

	t.Run("escaped dollar formatters literal dollar", func(t *testing.T) {
		t.Parallel()
		e := &Engine{
			Format: formatter.NewFormatter(false),
			Label:  label,
			Opts:   flag.Options{Colored: false},
		}
		got, err := runFSM(t, e, `\$VAR`)
		require.NoError(t, err)
		assert.Equal(t, "$VAR", got)
	})

	t.Run("dangling dollar at eof formatters dollar", func(t *testing.T) {
		t.Parallel()
		e := &Engine{
			Format: formatter.NewFormatter(false),
			Label:  label,
			Opts:   flag.Options{},
		}
		got, err := runFSM(t, e, "$")
		require.NoError(t, err)
		assert.Equal(t, "$", got)
	})

	t.Run("after dollar bare name expands", func(t *testing.T) {
		t.Parallel()
		e := &Engine{
			Label:  label,
			Opts:   flag.Options{},
			Format: formatter.NewFormatter(false),
			Lookup: func(name string) (string, bool) {
				if name == "NAME" {
					return "Ada", true
				}
				return "", false
			},
		}
		got, err := runFSM(t, e, "$NAME")
		require.NoError(t, err)
		assert.Equal(t, "Ada", got)
	})

	t.Run("after dollar non name token keeps dollar and literal", func(t *testing.T) {
		t.Parallel()
		e := &Engine{
			Label:  label,
			Opts:   flag.Options{},
			Format: formatter.NewFormatter(false),
		}
		got, err := runFSM(t, e, "$.")
		require.NoError(t, err)
		assert.Equal(t, "$.", got)
	})

	t.Run("braced simple expands", func(t *testing.T) {
		t.Parallel()
		e := &Engine{
			Label:  label,
			Opts:   flag.Options{},
			Format: formatter.NewFormatter(false),
			Lookup: func(name string) (string, bool) {
				if name == "NAME" {
					return "Ada", true
				}
				return "", false
			},
		}
		got, err := runFSM(t, e, "${NAME}")
		require.NoError(t, err)
		assert.Equal(t, "Ada", got)
	})

	t.Run("braced len form hash len", func(t *testing.T) {
		t.Parallel()
		e := &Engine{
			Label:  label,
			Opts:   flag.Options{},
			Format: formatter.NewFormatter(false),
			Lookup: func(name string) (string, bool) {
				if name == "S" {
					return "你好😊", true // 3 runes
				}
				return "", false
			},
		}
		got, err := runFSM(t, e, "${#S}")
		require.NoError(t, err)
		assert.Equal(t, "3", got)
	})

	t.Run("operators disabled keep literal", func(t *testing.T) {
		t.Parallel()
		e := &Engine{
			Label:  label,
			Opts:   flag.Options{NoOps: true, Colored: false},
			Format: formatter.NewFormatter(false),

			Lookup: func(name string) (string, bool) {
				return "", false
			},
		}
		in := "${NAME:-x}"
		got, err := runFSM(t, e, in)
		require.NoError(t, err)
		assert.Equal(t, in, got)
	})

	t.Run("unexpected token inside braces keeps literal", func(t *testing.T) {
		t.Parallel()
		e := &Engine{
			Label:  label,
			Opts:   flag.Options{},
			Format: formatter.NewFormatter(false),
		}
		in := "${NAME$}"
		got, err := runFSM(t, e, in)
		require.NoError(t, err)
		assert.Equal(t, in, got)
	})

	t.Run("braced op accumulates two chars then word executes", func(t *testing.T) {
		t.Parallel()
		e := &Engine{
			Label:  label,
			Opts:   flag.Options{},
			Format: formatter.NewFormatter(false),
			Lookup: func(name string) (string, bool) {
				if name == "V" {
					return "aaab", true
				}
				return "", false
			},
		}
		got, err := runFSM(t, e, "${V##a}")
		require.NoError(t, err)
		assert.Equal(t, "b", got)
	})

	t.Run("braced op without word calls expandWithOp", func(t *testing.T) {
		t.Parallel()
		e := &Engine{
			Label:  label,
			Opts:   flag.Options{},
			Format: formatter.NewFormatter(false),
			Lookup: func(name string) (string, bool) {
				return "value", true
			},
		}
		// ${VAR#} -> empty pattern => kept literal
		in := "${VAR#}"
		got, err := runFSM(t, e, in)
		require.NoError(t, err)
		assert.Equal(t, in, got)
	})

	t.Run("braced word supports nested braces and expands inside word", func(t *testing.T) {
		t.Parallel()
		e := &Engine{
			Label:  label,
			Opts:   flag.Options{},
			Format: formatter.NewFormatter(false),
			Lookup: func(name string) (string, bool) {
				switch name {
				case "VAR":
					return "", false // unset -> triggers default
				case "X":
					return "Y", true
				default:
					return "", false
				}
			},
		}
		got, err := runFSM(t, e, "${VAR:-a${X}b}")
		require.NoError(t, err)
		assert.Equal(t, "aYb", got)
	})

	t.Run("unterminated brace emits literal until eof", func(t *testing.T) {
		t.Parallel()
		e := &Engine{
			Format: formatter.NewFormatter(false),
			Label:  label,
			Opts:   flag.Options{Colored: false},
		}
		in := "${VAR:abc"
		got, err := runFSM(t, e, in)
		require.NoError(t, err)
		assert.Equal(t, in, got)
	})

	t.Run("unknown operator keeps literal form", func(t *testing.T) {
		t.Parallel()
		e := &Engine{
			Label:  label,
			Opts:   flag.Options{},
			Format: formatter.NewFormatter(false),
			Lookup: func(name string) (string, bool) { return "v", true },
		}
		in := "${VAR~x}"
		got, err := runFSM(t, e, in)
		require.NoError(t, err)
		assert.Equal(t, in, got)
	})

	t.Run("error from expandWithOp propagates and stops", func(t *testing.T) {
		t.Parallel()
		e := &Engine{
			Label:  label,
			Opts:   flag.Options{},
			Format: formatter.NewFormatter(false),
			Lookup: func(name string) (string, bool) { return "", false }, // unset
		}
		got, err := runFSM(t, e, "${VAR?boom}")
		require.Error(t, err)
		assert.EqualError(t, err, "VAR: boom")
		assert.Equal(t, "", got)
	})
}
