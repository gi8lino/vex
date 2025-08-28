package fsm

import (
	"testing"

	"github.com/gi8lino/vex/internal/flag"
	"github.com/gi8lino/vex/internal/formatter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFilter(t *testing.T) {
	t.Parallel()
	const label = "EngineLabel"

	t.Run("allows all when lists empty", func(t *testing.T) {
		t.Parallel()

		e := &Engine{
			Format: formatter.NewFormatter(false),
			Label:  label,
			Opts:   flag.Options{}, // Variables/Prefix/Suffix all empty
		}
		require.True(t, e.filter("ANY"))
		require.True(t, e.filter("foo"))
		require.True(t, e.filter("bar baz"))
	})

	t.Run("allows when name in Variables", func(t *testing.T) {
		t.Parallel()

		e := &Engine{
			Format: formatter.NewFormatter(false),
			Label:  label,
			Opts: flag.Options{
				Variables: []string{"FOO", "BAR"},
			},
		}
		assert.True(t, e.filter("FOO"))
		assert.True(t, e.filter("BAR"))
		assert.False(t, e.filter("BAZ"))
	})

	t.Run("allows when name matches any prefix", func(t *testing.T) {
		t.Parallel()

		e := &Engine{
			Format: formatter.NewFormatter(false),
			Label:  label,
			Opts: flag.Options{
				Prefix: []string{"APP ", "SYS "},
			},
		}
		assert.True(t, e.filter("APP CONFIG"))
		assert.True(t, e.filter("SYS PATH"))
		assert.False(t, e.filter("USER NAME"))
	})

	t.Run("allows when name matches any suffix", func(t *testing.T) {
		t.Parallel()

		e := &Engine{
			Format: formatter.NewFormatter(false),
			Label:  label,
			Opts: flag.Options{
				Suffix: []string{" TOKEN", " ID"},
			},
		}
		assert.True(t, e.filter("GITHUB TOKEN"))
		assert.True(t, e.filter("USER ID"))
		assert.False(t, e.filter("SESSION KEY"))
	})

	t.Run("denies when lists nonempty and no match", func(t *testing.T) {
		t.Parallel()

		e := &Engine{
			Format: formatter.NewFormatter(false),
			Label:  label,
			Opts: flag.Options{
				Variables: []string{"ONLY THIS"},
				Prefix:    []string{"ALLOWED "},
				Suffix:    []string{" OK"},
			},
		}
		assert.False(t, e.filter("NOT ALLOWED"))
		assert.False(t, e.filter("PREFIX NOPE"))
		assert.False(t, e.filter("NOPE SUFFIX"))
	})

	t.Run("mixed lists precedence any match allows", func(t *testing.T) {
		t.Parallel()

		e := &Engine{
			Format: formatter.NewFormatter(false),
			Label:  label,
			Opts: flag.Options{
				Variables: []string{"EXACT"},
				Prefix:    []string{"PRE "},
				Suffix:    []string{" SUF"},
			},
		}
		// Exact variable match
		assert.True(t, e.filter("EXACT"))
		// Prefix match (not in Variables)
		assert.True(t, e.filter("PRE NAME"))
		// Suffix match (not in Variables or Prefix)
		assert.True(t, e.filter("NAME SUF"))
		// No match in any list => false
		assert.False(t, e.filter("MIDDLE"))
	})

	t.Run("case sensitive matching", func(t *testing.T) {
		t.Parallel()

		e := &Engine{
			Format: formatter.NewFormatter(false),
			Label:  label,
			Opts: flag.Options{
				Variables: []string{"CaseSensitive"},
				Prefix:    []string{"Pre "},
				Suffix:    []string{" Suf"},
			},
		}
		assert.True(t, e.filter("CaseSensitive"))
		assert.False(t, e.filter("casesensitive")) // different case
		assert.True(t, e.filter("Pre Value"))
		assert.False(t, e.filter("pre Value")) // different case
		assert.True(t, e.filter("Value Suf"))
		assert.False(t, e.filter("Value suf")) // different case
	})
}
