package formatter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewFormatter_Colored(t *testing.T) {
	t.Parallel()

	t.Run("returns colored impl", func(t *testing.T) {
		t.Parallel()
		got := NewFormatter(true)
		// Not an "error" return, but ensure type is as expected.
		assert.IsType(t, coloredFormatter{}, got)
	})
}

func TestNewFormatter_Plain(t *testing.T) {
	t.Parallel()
	t.Run("returns plain impl", func(t *testing.T) {
		t.Parallel()
		got := NewFormatter(false)
		assert.IsType(t, plainFormatter{}, got)
	})
}
