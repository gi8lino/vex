package fsm

import (
	"testing"

	"github.com/gi8lino/vex/internal/flag"
	"github.com/gi8lino/vex/internal/formatter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpReplace(t *testing.T) {
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
		out, err := e.opReplace("VAR", "/", false /*isSet*/, "ignored", "a/b")
		require.Error(t, err)
		assert.EqualError(t, err, "variable not set: VAR")
		assert.Empty(t, out)
	})

	t.Run("returns missing marker when unset and NoReplaceUnset single", func(t *testing.T) {
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
		out, err := e.opReplace("VAR", "/", false /*isSet*/, "ignored", "a/b")
		require.NoError(t, err)
		assert.Equal(t, "${VAR/a/b}", out)
	})

	t.Run("returns missing marker when unset and NoReplaceUnset global", func(t *testing.T) {
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
		out, err := e.opReplace("VAR", "//", false /*isSet*/, "ignored", "x/y")
		require.NoError(t, err)
		assert.Equal(t, "${VAR//x/y}", out)
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
		out, err := e.opReplace("VAR", "/", false /*isSet*/, "ignored", "a/b")
		require.NoError(t, err)
		assert.Equal(t, "", out)
	})

	t.Run("replace first with single slash", func(t *testing.T) {
		t.Parallel()

		e := &Engine{
			Format: formatter.NewFormatter(false),
			Label:  label,
			Opts:   flag.Options{},
		}
		// only the first "aa" becomes "X"
		out, err := e.opReplace("VAR", "/", true /*isSet*/, "aa-aa-aa", "aa/X")
		require.NoError(t, err)
		assert.Equal(t, "X-aa-aa", out)
	})

	t.Run("replace all with double slash", func(t *testing.T) {
		t.Parallel()

		e := &Engine{
			Format: formatter.NewFormatter(false),
			Label:  label,
			Opts:   flag.Options{},
		}
		out, err := e.opReplace("VAR", "//", true /*isSet*/, "aa-aa-aa", "aa/X")
		require.NoError(t, err)
		assert.Equal(t, "X-X-X", out)
	})

	t.Run("pattern not found returns original", func(t *testing.T) {
		t.Parallel()

		e := &Engine{
			Format: formatter.NewFormatter(false),
			Label:  label,
			Opts:   flag.Options{},
		}
		out, err := e.opReplace("VAR", "//", true /*isSet*/, "hello", "zzz/X")
		require.NoError(t, err)
		assert.Equal(t, "hello", out)
	})

	t.Run("empty pattern yields missing literal", func(t *testing.T) {
		t.Parallel()

		e := &Engine{
			Format: formatter.NewFormatter(false),
			Label:  label,
			Opts:   flag.Options{},
		}
		// spec starts with '/', so pat == ""
		out, err := e.opReplace("VAR", "/", true /*isSet*/, "value", "/repl")
		require.NoError(t, err)
		assert.Equal(t, "${VAR//repl}", out) // op + spec preserved exactly
	})

	t.Run("error when result empty and FailOnEmpty single", func(t *testing.T) {
		t.Parallel()

		e := &Engine{
			Format: formatter.NewFormatter(false),
			Label:  label,
			Opts: flag.Options{
				ErrorEmpty: true,
			},
		}
		// replace the only char → result becomes empty → error
		out, err := e.opReplace("VAR", "/", true /*isSet*/, "a", "a/")
		require.Error(t, err)
		assert.EqualError(t, err, "substitution empty: VAR")
		assert.Empty(t, out)
	})

	t.Run("error when result empty and FailOnEmpty global", func(t *testing.T) {
		t.Parallel()

		e := &Engine{
			Format: formatter.NewFormatter(false),
			Label:  label,
			Opts: flag.Options{
				ErrorEmpty: true,
			},
		}
		// remove all occurrences → empty
		out, err := e.opReplace("VAR", "//", true /*isSet*/, "aaa", "a/")
		require.Error(t, err)
		assert.EqualError(t, err, "substitution empty: VAR")
		assert.Empty(t, out)
	})
}
