package fsm

import (
	"testing"

	"github.com/gi8lino/vex/internal/flag"
	"github.com/gi8lino/vex/internal/formatter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpTrimPrefix(t *testing.T) {
	t.Parallel()
	const label = "EngineLabel"

	t.Run("error when unset and FailOnUnset", func(t *testing.T) {
		t.Parallel()
		e := &Engine{
			Format: formatter.NewFormatter(false),
			Label:  label,
			Opts: flag.Options{
				ErrorUnset: true,
			},
		}
		out, err := e.opTrimPrefix("VAR", "#", false /*isSet*/, "ignored", "a")
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
				KeepUnset: true,
			},
		}
		out, err := e.opTrimPrefix("VAR", "#", false /*isSet*/, "ignored", "a")
		require.NoError(t, err)
		assert.Equal(t, "${VAR#a}", out)
	})

	t.Run("returns missing marker when unset and NoReplaceUnset double", func(t *testing.T) {
		t.Parallel()
		e := &Engine{
			Format: formatter.NewFormatter(false),
			Label:  label,
			Opts: flag.Options{
				KeepUnset: true,
			},
		}
		out, err := e.opTrimPrefix("VAR", "##", false /*isSet*/, "ignored", "a")
		require.NoError(t, err)
		assert.Equal(t, "${VAR##a}", out)
	})

	t.Run("unset without NoReplaceUnset returns empty string", func(t *testing.T) {
		t.Parallel()
		e := &Engine{
			Format: formatter.NewFormatter(false),
			Label:  label,
			Opts:   flag.Options{},
		}
		out, err := e.opTrimPrefix("VAR", "#", false /*isSet*/, "ignored", "x")
		require.NoError(t, err)
		assert.Equal(t, "", out)
	})

	t.Run("empty pattern yields missing literal single", func(t *testing.T) {
		t.Parallel()
		e := &Engine{
			Format: formatter.NewFormatter(false),
			Label:  label,
			Opts:   flag.Options{},
		}
		out, err := e.opTrimPrefix("VAR", "#", true /*isSet*/, "value", "")
		require.NoError(t, err)
		assert.Equal(t, "${VAR#}", out)
	})

	t.Run("empty pattern yields missing literal double", func(t *testing.T) {
		t.Parallel()
		e := &Engine{
			Format: formatter.NewFormatter(false),
			Label:  label,
			Opts:   flag.Options{},
		}
		out, err := e.opTrimPrefix("VAR", "##", true /*isSet*/, "value", "")
		require.NoError(t, err)
		assert.Equal(t, "${VAR##}", out)
	})

	t.Run("trim once when matches prefix", func(t *testing.T) {
		t.Parallel()
		e := &Engine{
			Label:  label,
			Format: formatter.NewFormatter(false),
		}
		out, err := e.opTrimPrefix("VAR", "#", true /*isSet*/, "aaab", "a")
		require.NoError(t, err)
		assert.Equal(t, "aab", out)
	})

	t.Run("trim all when matches prefix repeatedly", func(t *testing.T) {
		t.Parallel()
		e := &Engine{
			Label:  label,
			Format: formatter.NewFormatter(false),
		}
		out, err := e.opTrimPrefix("VAR", "##", true /*isSet*/, "aaab", "a")
		require.NoError(t, err)
		assert.Equal(t, "b", out)
	})

	t.Run("no change when pattern not at prefix", func(t *testing.T) {
		t.Parallel()
		e := &Engine{
			Label:  label,
			Format: formatter.NewFormatter(false),
		}
		out, err := e.opTrimPrefix("VAR", "##", true /*isSet*/, "baaa", "a")
		require.NoError(t, err)
		assert.Equal(t, "baaa", out)
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
		out, err := e.opTrimPrefix("VAR", "##", true /*isSet*/, "aaa", "a")
		require.Error(t, err)
		assert.EqualError(t, err, "substitution empty: VAR")
		assert.Empty(t, out)
	})
}

func TestOpTrimSuffix(t *testing.T) {
	t.Parallel()
	const label = "EngineLabel"

	t.Run("error when unset and FailOnUnset", func(t *testing.T) {
		t.Parallel()
		e := &Engine{
			Format: formatter.NewFormatter(false),
			Label:  label,
			Opts: flag.Options{
				ErrorUnset: true,
			},
		}
		out, err := e.opTrimSuffix("VAR", "%", false /*isSet*/, "ignored", "a")
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
				KeepUnset: true,
			},
		}
		out, err := e.opTrimSuffix("VAR", "%", false /*isSet*/, "ignored", "a")
		require.NoError(t, err)
		assert.Equal(t, "${VAR%a}", out)
	})

	t.Run("returns missing marker when unset and NoReplaceUnset double", func(t *testing.T) {
		t.Parallel()
		e := &Engine{
			Format: formatter.NewFormatter(false),
			Label:  label,
			Opts: flag.Options{
				KeepUnset: true,
			},
		}
		out, err := e.opTrimSuffix("VAR", "%%", false /*isSet*/, "ignored", "a")
		require.NoError(t, err)
		assert.Equal(t, "${VAR%%a}", out)
	})

	t.Run("unset without NoReplaceUnset returns empty string", func(t *testing.T) {
		t.Parallel()
		e := &Engine{
			Format: formatter.NewFormatter(false),
			Label:  label,
			Opts:   flag.Options{},
		}
		out, err := e.opTrimSuffix("VAR", "%", false /*isSet*/, "ignored", "x")
		require.NoError(t, err)
		assert.Equal(t, "", out)
	})

	t.Run("empty pattern yields missing literal single", func(t *testing.T) {
		t.Parallel()
		e := &Engine{
			Format: formatter.NewFormatter(false),
			Label:  label,
			Opts:   flag.Options{},
		}
		out, err := e.opTrimSuffix("VAR", "%", true /*isSet*/, "value", "")
		require.NoError(t, err)
		assert.Equal(t, "${VAR%}", out)
	})

	t.Run("empty pattern yields missing literal double", func(t *testing.T) {
		t.Parallel()
		e := &Engine{
			Format: formatter.NewFormatter(false),
			Label:  label,
			Opts:   flag.Options{},
		}
		out, err := e.opTrimSuffix("VAR", "%%", true /*isSet*/, "value", "")
		require.NoError(t, err)
		assert.Equal(t, "${VAR%%}", out)
	})

	t.Run("trim once when matches suffix", func(t *testing.T) {
		t.Parallel()
		e := &Engine{
			Label:  label,
			Format: formatter.NewFormatter(false),
		}
		out, err := e.opTrimSuffix("VAR", "%", true /*isSet*/, "baaa", "a")
		require.NoError(t, err)
		assert.Equal(t, "baa", out)
	})

	t.Run("trim all when matches suffix repeatedly", func(t *testing.T) {
		t.Parallel()
		e := &Engine{
			Format: formatter.NewFormatter(false),
			Label:  label,
		}
		out, err := e.opTrimSuffix("VAR", "%%", true /*isSet*/, "baaa", "a")
		require.NoError(t, err)
		assert.Equal(t, "b", out)
	})

	t.Run("no change when pattern not at suffix", func(t *testing.T) {
		t.Parallel()
		e := &Engine{
			Label:  label,
			Format: formatter.NewFormatter(false),
		}
		out, err := e.opTrimSuffix("VAR", "%%", true /*isSet*/, "aaab", "a")
		require.NoError(t, err)
		assert.Equal(t, "aaab", out)
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
		out, err := e.opTrimSuffix("VAR", "%%", true /*isSet*/, "aaa", "a")
		require.Error(t, err)
		assert.EqualError(t, err, "substitution empty: VAR")
		assert.Empty(t, out)
	})
}
