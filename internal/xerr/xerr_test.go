package xerr

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnset(t *testing.T) {
	t.Parallel()
	t.Run("Wraps Subst error and preserves message", func(t *testing.T) {
		t.Parallel()

		msg := "${FOO} is missing"
		err := Unset(msg)
		require.Error(t, err)

		assert.ErrorIs(t, err, ErrSubst)
		assert.NotErrorIs(t, err, ErrEmpty)
		assert.EqualError(t, err, "variable not set: "+msg)
	})

	t.Run("Unset is not empty", func(t *testing.T) {
		t.Parallel()

		errUnset := Unset("${X}")
		require.Error(t, errUnset)

		assert.ErrorIs(t, errUnset, ErrSubst)
		assert.NotErrorIs(t, errUnset, ErrEmpty)
	})

	t.Run("Wrapped Unset error", func(t *testing.T) {
		t.Parallel()

		wrapped := fmt.Errorf("wrap: %w", Unset("$FOO not set"))
		require.Error(t, wrapped)

		assert.ErrorIs(t, wrapped, ErrSubst)
		assert.NotErrorIs(t, wrapped, ErrEmpty)
	})
}

func TestEmpty(t *testing.T) {
	t.Parallel()

	t.Run("Wraps Empty error and preserves message", func(t *testing.T) {
		msg := "${BAR} expanded to empty"
		err := Empty(msg)
		require.Error(t, err)

		assert.ErrorIs(t, err, ErrEmpty)
		assert.NotErrorIs(t, err, ErrSubst)
		assert.EqualError(t, err, "substitution empty: "+msg)
	})

	t.Run("Empty is not unset", func(t *testing.T) {
		t.Parallel()

		errEmpty := Empty("${Y}")
		require.Error(t, errEmpty)

		assert.ErrorIs(t, errEmpty, ErrEmpty)
		assert.NotErrorIs(t, errEmpty, ErrSubst)
	})

	t.Run("Wrapped Empty error", func(t *testing.T) {
		t.Parallel()

		wrapped := fmt.Errorf("wrap: %w", Empty("${BAR} empty"))
		require.Error(t, wrapped)

		assert.ErrorIs(t, wrapped, ErrEmpty)
		assert.NotErrorIs(t, wrapped, ErrSubst)
	})
}
