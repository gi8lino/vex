package fsm

import (
	"testing"

	"github.com/gi8lino/vex/internal/flag"
	"github.com/gi8lino/vex/internal/formatter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpLen(t *testing.T) {
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
		out, err := e.opLen("VAR", false /*isSet*/, "ignored")
		require.Error(t, err)
		assert.EqualError(t, err, "variable not set: VAR")
		assert.Empty(t, out)
	})

	t.Run("returns zero when unset and not FailOnUnset", func(t *testing.T) {
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
		out, err := e.opLen("VAR", false /*isSet*/, "ignored")
		require.NoError(t, err)
		assert.Equal(t, "0", out)
	})

	t.Run("returns length for ascii", func(t *testing.T) {
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
		out, err := e.opLen("VAR", true /*isSet*/, "abcde")
		require.NoError(t, err)
		assert.Equal(t, "5", out)
	})

	t.Run("returns rune length for unicode", func(t *testing.T) {
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
		// 你好 (2 runes) + 😊 (1 rune) => total 3
		out, err := e.opLen("VAR", true /*isSet*/, "你好😊")
		require.NoError(t, err)
		assert.Equal(t, "3", out)
	})

	t.Run("returns zero for empty string when set", func(t *testing.T) {
		t.Parallel()

		e := &Engine{
			Format: formatter.NewFormatter(false),
			Label:  label,
			Opts: flag.Options{
				ErrorUnset: false,
				KeepUnset:  false,
				ErrorEmpty: true, // irrelevant for opLen; included to mirror template
			},
		}
		out, err := e.opLen("EMPTY", true /*isSet*/, "")
		require.NoError(t, err)
		assert.Equal(t, "0", out)
	})
}
