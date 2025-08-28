package fsm

import (
	"testing"

	"github.com/gi8lino/vex/internal/flag"
	"github.com/gi8lino/vex/internal/formatter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpDefault(t *testing.T) {
	t.Parallel()
	const label = "EngineLabel"

	t.Run("returns default when unset", func(t *testing.T) {
		t.Parallel()
		e := &Engine{
			Format: formatter.NewFormatter(false),
			Label:  label,
			Opts:   flag.Options{},
		}
		out, err := e.opDefault(false /*isSet*/, "ignored", "fallback")
		require.NoError(t, err)
		assert.Equal(t, "fallback", out)
	})

	t.Run("returns value when set", func(t *testing.T) {
		t.Parallel()
		e := &Engine{
			Format: formatter.NewFormatter(false),
			Label:  label,
			Opts:   flag.Options{},
		}
		out, err := e.opDefault(true /*isSet*/, "value", "fallback")
		require.NoError(t, err)
		assert.Equal(t, "value", out)
	})
}

func TestOpDefaultNull(t *testing.T) {
	t.Parallel()
	const label = "EngineLabel"

	t.Run("returns default when null or unset", func(t *testing.T) {
		t.Parallel()
		e := &Engine{
			Format: formatter.NewFormatter(false),
			Label:  label,
			Opts:   flag.Options{},
		}
		out, err := e.opDefaultNull(false /*notNull*/, "" /*val*/, "fallback")
		require.NoError(t, err)
		assert.Equal(t, "fallback", out)
	})

	t.Run("returns value when notNull", func(t *testing.T) {
		t.Parallel()
		e := &Engine{
			Format: formatter.NewFormatter(false),
			Label:  label,
			Opts:   flag.Options{},
		}
		out, err := e.opDefaultNull(true /*notNull*/, "value", "fallback")
		require.NoError(t, err)
		assert.Equal(t, "value", out)
	})
}

func TestOpAssign(t *testing.T) {
	t.Parallel()
	const label = "EngineLabel"

	t.Run("assigns and returns default when unset", func(t *testing.T) {
		t.Parallel()
		e := &Engine{
			Format: formatter.NewFormatter(false),
			Label:  label,
			Opts:   flag.Options{},
			Setenv: func(name, val string) error { return nil },
		}
		out, err := e.opAssign("VAR", false /*isSet*/, "ignored", "fallback")
		require.NoError(t, err)
		assert.Equal(t, "fallback", out)
		// (Optional) If Engine exposes Getenv, you could assert the env was set here.
	})

	t.Run("returns value when set", func(t *testing.T) {
		t.Parallel()
		e := &Engine{
			Format: formatter.NewFormatter(false),
			Label:  label,
			Opts:   flag.Options{},
		}
		out, err := e.opAssign("VAR", true /*isSet*/, "value", "fallback")
		require.NoError(t, err)
		assert.Equal(t, "value", out)
	})
}

func TestOpAssignNull(t *testing.T) {
	t.Parallel()
	const label = "EngineLabel"

	t.Run("assigns and returns default when null or unset", func(t *testing.T) {
		t.Parallel()
		e := &Engine{
			Format: formatter.NewFormatter(false),
			Label:  label,
			Opts:   flag.Options{},
			Setenv: func(name, val string) error { return nil },
		}
		out, err := e.opAssignNull("VAR", false /*notNull*/, "" /*val*/, "fallback")
		require.NoError(t, err)
		assert.Equal(t, "fallback", out)
	})

	t.Run("returns value when notNull", func(t *testing.T) {
		t.Parallel()
		e := &Engine{
			Format: formatter.NewFormatter(false),
			Label:  label,
			Opts:   flag.Options{},
		}
		out, err := e.opAssignNull("VAR", true /*notNull*/, "value", "fallback")
		require.NoError(t, err)
		assert.Equal(t, "value", out)
	})
}

func TestOpAlt(t *testing.T) {
	t.Parallel()
	const label = "EngineLabel"

	t.Run("returns word when set", func(t *testing.T) {
		t.Parallel()
		e := &Engine{
			Format: formatter.NewFormatter(false),
			Label:  label,
			Opts:   flag.Options{},
		}
		out, err := e.opAlt("VAR", true /*isSet*/, "alt")
		require.NoError(t, err)
		assert.Equal(t, "VAR: alt", out)
	})

	t.Run("returns empty when unset", func(t *testing.T) {
		t.Parallel()
		e := &Engine{
			Format: formatter.NewFormatter(false),
			Label:  label,
			Opts:   flag.Options{},
		}
		out, err := e.opAlt("VAR", false /*isSet*/, "alt")
		require.NoError(t, err)
		assert.Equal(t, "", out)
	})
}

func TestOpAltNull(t *testing.T) {
	t.Parallel()
	const label = "EngineLabel"

	t.Run("returns word when notNull", func(t *testing.T) {
		t.Parallel()
		e := &Engine{
			Format: formatter.NewFormatter(false),
			Label:  label,
			Opts:   flag.Options{},
		}
		out, err := e.opAltNull("VAR", true /*notNull*/, "alt")
		require.NoError(t, err)
		assert.Equal(t, "VAR: alt", out)
	})

	t.Run("returns empty when null or unset", func(t *testing.T) {
		t.Parallel()
		e := &Engine{
			Format: formatter.NewFormatter(false),
			Label:  label,
			Opts:   flag.Options{},
		}
		out, err := e.opAltNull("VAR", false /*notNull*/, "alt")
		require.NoError(t, err)
		assert.Equal(t, "", out)
	})
}

func TestOpErrorUnset(t *testing.T) {
	t.Parallel()
	const label = "EngineLabel"

	t.Run("errors when unset", func(t *testing.T) {
		t.Parallel()
		e := &Engine{
			Format: formatter.NewFormatter(false),
			Label:  label,
			Opts:   flag.Options{},
		}
		out, err := e.opErrorUnset("VAR", false /*isSet*/, "boom")
		require.Error(t, err)
		assert.EqualError(t, err, "VAR: boom")
		assert.Empty(t, out)
	})

	t.Run("ok when set returns empty string", func(t *testing.T) {
		t.Parallel()
		e := &Engine{
			Format: formatter.NewFormatter(false),
			Label:  label,
			Opts:   flag.Options{},
		}
		out, err := e.opErrorUnset("VAR", true /*isSet*/, "boom")
		require.NoError(t, err)
		assert.Equal(t, "", out)
	})
}

func TestOpErrorNull(t *testing.T) {
	t.Parallel()
	const label = "EngineLabel"

	t.Run("errors when null or unset", func(t *testing.T) {
		t.Parallel()
		e := &Engine{
			Format: formatter.NewFormatter(false),
			Label:  label,
			Opts:   flag.Options{},
		}
		out, err := e.opErrorNull("VAR", false /*notNull*/, "boom")
		require.Error(t, err)
		assert.EqualError(t, err, "VAR: boom")
		assert.Empty(t, out)
	})

	t.Run("ok when notNull returns empty string", func(t *testing.T) {
		t.Parallel()
		e := &Engine{
			Format: formatter.NewFormatter(false),
			Label:  label,
			Opts:   flag.Options{},
		}
		out, err := e.opErrorNull("VAR", true /*notNull*/, "boom")
		require.NoError(t, err)
		assert.Equal(t, "", out)
	})
}
