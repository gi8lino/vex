package fsm

import (
	"testing"

	"github.com/gi8lino/vex/internal/flag"
	"github.com/gi8lino/vex/internal/formatter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpSubstr(t *testing.T) {
	t.Parallel()
	const label = "EngineLabel"

	t.Run("error when unset and FailOnUnset", func(t *testing.T) {
		t.Parallel()

		e := &Engine{
			Format: formatter.NewFormatter(false),
			Label:  label,
			Opts: flag.Options{
				ErrorUnset: true,
				KeepUnset:  false,
				ErrorEmpty: false,
			},
		}
		out, err := e.opSubstr("VAR", false /*isSet*/, "ignored", "1:2")
		require.Error(t, err)
		assert.EqualError(t, err, "variable not set: VAR")
		assert.Empty(t, out)
	})

	t.Run("returns missing marker when unset and NoReplaceUnset", func(t *testing.T) {
		t.Parallel()

		e := &Engine{
			Format: formatter.NewFormatter(false),
			Label:  label,
			Opts: flag.Options{
				ErrorUnset: false,
				KeepUnset:  true,
				ErrorEmpty: false,
			},
		}
		out, err := e.opSubstr("VAR", false /*isSet*/, "ignored", "1:2")
		require.NoError(t, err)
		assert.Equal(t, "${VAR:1:2}", out)
	})

	t.Run("unset without NoReplaceUnset returns empty string", func(t *testing.T) {
		t.Parallel()

		e := &Engine{
			Format: formatter.NewFormatter(false),
			Label:  label,
			Opts: flag.Options{
				ErrorUnset: false,
				KeepUnset:  false,
				ErrorEmpty: false,
			},
		}
		out, err := e.opSubstr("VAR", false /*isSet*/, "ignored", "0:3")
		require.NoError(t, err)
		assert.Equal(t, "", out)
	})

	t.Run("substr offset only when set", func(t *testing.T) {
		t.Parallel()

		e := &Engine{
			Format: formatter.NewFormatter(false),
			Label:  label,
			Opts:   flag.Options{},
		}
		// Assuming bash-like 0-based indexing: "abcdef"[2:] -> "cdef"
		out, err := e.opSubstr("VAR", true /*isSet*/, "abcdef", "2")
		require.NoError(t, err)
		assert.Equal(t, "cdef", out)
	})

	t.Run("substr with length when set", func(t *testing.T) {
		t.Parallel()

		e := &Engine{
			Format: formatter.NewFormatter(false),
			Label:  label,
			Opts:   flag.Options{},
		}
		// "abcdef"[1:1+3] -> "bcd"
		out, err := e.opSubstr("VAR", true /*isSet*/, "abcdef", "1:3")
		require.NoError(t, err)
		assert.Equal(t, "bcd", out)
	})

	t.Run("error when result empty and FailOnEmpty", func(t *testing.T) {
		t.Parallel()

		e := &Engine{
			Format: formatter.NewFormatter(false),
			Label:  label,
			Opts: flag.Options{
				ErrorEmpty: true,
			},
		}
		// Offset beyond end → empty result triggers FailOnEmpty
		out, err := e.opSubstr("VAR", true /*isSet*/, "abc", "10:1")
		require.Error(t, err)
		assert.EqualError(t, err, "substitution empty: VAR")
		assert.Empty(t, out)
	})

	t.Run("zero length when set without FailOnEmpty", func(t *testing.T) {
		t.Parallel()

		e := &Engine{
			Format: formatter.NewFormatter(false),
			Label:  label,
			Opts:   flag.Options{},
		}
		out, err := e.opSubstr("VAR", true /*isSet*/, "abc", "0:0")
		require.NoError(t, err)
		assert.Equal(t, "", out)
	})
}
