package fsm

import (
	"testing"

	"github.com/gi8lino/vex/internal/flag"
	"github.com/gi8lino/vex/internal/formatter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpCase(t *testing.T) {
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
		out, err := e.opCase("VAR", "^^", false /*isSet*/, "ignored")
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
		out, err := e.opCase("VAR", ",", false /*isSet*/, "ignored")
		require.NoError(t, err)
		assert.Equal(t, "${VAR,}", out)
	})

	t.Run("error when result empty and FailOnEmpty", func(t *testing.T) {
		t.Parallel()

		e := &Engine{
			Format: formatter.NewFormatter(false),
			Label:  label,
			Opts: flag.Options{
				ErrorUnset: false,
				KeepUnset:  false,
				ErrorEmpty: true,
			},
		}
		// Unset + NoReplaceUnset=false => val becomes "" and transformCase keeps it "".
		out, err := e.opCase("EMPTY", "^^", false /*isSet*/, "ignored")
		require.Error(t, err)
		assert.EqualError(t, err, "substitution empty: EMPTY")
		assert.Empty(t, out)
	})

	t.Run("upper first with ^", func(t *testing.T) {
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
		out, err := e.opCase("VAR", "^", true /*isSet*/, "hello")
		require.NoError(t, err)
		assert.Equal(t, "Hello", out, `${VAR^} => "Hello"`)
	})

	t.Run("upper all with ^^", func(t *testing.T) {
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
		out, err := e.opCase("VAR", "^^", true /*isSet*/, "AbcXyZ")
		require.NoError(t, err)
		assert.Equal(t, "ABCXYZ", out, `${VAR^^} => "ABCXYZ"`)
	})

	t.Run("lower first with ,", func(t *testing.T) {
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
		out, err := e.opCase("VAR", ",", true /*isSet*/, "Hello")
		require.NoError(t, err)
		assert.Equal(t, "hello", out, `${VAR,} => "hello"`)
	})

	t.Run("lower all with ,,", func(t *testing.T) {
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
		out, err := e.opCase("VAR", ",,", true /*isSet*/, "AbcXyZ")
		require.NoError(t, err)
		assert.Equal(t, "abcxyz", out, `${VAR,,} => "abcxyz"`)
	})
}
