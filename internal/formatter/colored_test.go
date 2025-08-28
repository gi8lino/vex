package formatter

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestColoredFormatter_OkStr(t *testing.T) {
	t.Parallel()
	f := coloredFormatter{}

	t.Run("basic", func(t *testing.T) {
		t.Parallel()
		got := f.OkStr("ok")
		expected := green + "ok" + reset
		assert.True(t, strings.HasPrefix(got, green))
		assert.True(t, strings.HasSuffix(got, reset))
		assert.Equal(t, expected, got)
		assert.Equal(t, 1, strings.Count(got, green))
		assert.Equal(t, 1, strings.Count(got, reset))
	})

	t.Run("empty", func(t *testing.T) {
		t.Parallel()
		got := f.OkStr("")
		assert.Equal(t, green+""+reset, got)
	})
}

func TestColoredFormatter_DefaultStr(t *testing.T) {
	t.Parallel()
	f := coloredFormatter{}

	t.Run("basic", func(t *testing.T) {
		t.Parallel()
		got := f.DefaultStr("def")
		expected := yell + "def" + reset
		assert.True(t, strings.HasPrefix(got, yell))
		assert.True(t, strings.HasSuffix(got, reset))
		assert.Equal(t, expected, got)
		assert.Equal(t, 1, strings.Count(got, yell))
		assert.Equal(t, 1, strings.Count(got, reset))
	})

	t.Run("empty", func(t *testing.T) {
		t.Parallel()
		got := f.DefaultStr("")
		assert.Equal(t, yell+""+reset, got)
	})
}

func TestColoredFormatter_UserErrorStr(t *testing.T) {
	t.Parallel()
	f := coloredFormatter{}

	t.Run("basic", func(t *testing.T) {
		t.Parallel()
		got := f.UserErrorStr("user boom")
		expected := purple + "user boom" + reset
		assert.True(t, strings.HasPrefix(got, purple))
		assert.True(t, strings.HasSuffix(got, reset))
		assert.Equal(t, expected, got)
		assert.Equal(t, 1, strings.Count(got, purple))
		assert.Equal(t, 1, strings.Count(got, reset))
	})

	t.Run("empty", func(t *testing.T) {
		t.Parallel()
		got := f.UserErrorStr("")
		assert.Equal(t, purple+""+reset, got)
	})
}

func TestColoredFormatter_ErrorStr(t *testing.T) {
	t.Parallel()
	f := coloredFormatter{}

	t.Run("basic", func(t *testing.T) {
		t.Parallel()
		got := f.ErrorStr("engine err")
		expected := red + "engine err" + reset
		assert.True(t, strings.HasPrefix(got, red))
		assert.True(t, strings.HasSuffix(got, reset))
		assert.Equal(t, expected, got)
		assert.Equal(t, 1, strings.Count(got, red))
		assert.Equal(t, 1, strings.Count(got, reset))
	})

	t.Run("empty", func(t *testing.T) {
		t.Parallel()
		got := f.ErrorStr("")
		assert.Equal(t, red+""+reset, got)
	})
}

func TestColoredFormatter_EmptyUnsetFilterStr(t *testing.T) {
	t.Parallel()
	f := coloredFormatter{}

	t.Run("EmptyStr basic and empty", func(t *testing.T) {
		t.Parallel()
		got := f.EmptyStr("empty")
		assert.Equal(t, orange+"empty"+reset, got)
		got2 := f.EmptyStr("")
		assert.Equal(t, orange+""+reset, got2)
	})

	t.Run("UnsetStr basic and empty", func(t *testing.T) {
		t.Parallel()
		got := f.UnsetStr("unset")
		assert.Equal(t, magenta+"unset"+reset, got)
		got2 := f.UnsetStr("")
		assert.Equal(t, magenta+""+reset, got2)
	})

	t.Run("FilterStr basic and empty", func(t *testing.T) {
		t.Parallel()
		got := f.FilterStr("filtered")
		assert.Equal(t, gray+"filtered"+reset, got)
		got2 := f.FilterStr("")
		assert.Equal(t, gray+""+reset, got2)
	})
}

func TestANSIConstants_Sanity(t *testing.T) {
	t.Parallel()

	t.Run("non-empty", func(t *testing.T) {
		t.Parallel()
		assert.NotEmpty(t, green)
		assert.NotEmpty(t, yell)
		assert.NotEmpty(t, red)
		assert.NotEmpty(t, orange)
		assert.NotEmpty(t, magenta)
		assert.NotEmpty(t, purple)
		assert.NotEmpty(t, gray)
		assert.NotEmpty(t, reset)
	})

	t.Run("distinct", func(t *testing.T) {
		t.Parallel()
		assert.NotEqual(t, green, yell)
		assert.NotEqual(t, green, red)
		assert.NotEqual(t, green, orange)
		assert.NotEqual(t, green, magenta)
		assert.NotEqual(t, green, purple)
		assert.NotEqual(t, green, gray)

		assert.NotEqual(t, yell, red)
		assert.NotEqual(t, yell, orange)
		assert.NotEqual(t, yell, magenta)
		assert.NotEqual(t, yell, purple)
		assert.NotEqual(t, yell, gray)

		assert.NotEqual(t, red, orange)
		assert.NotEqual(t, red, magenta)
		assert.NotEqual(t, red, purple)
		assert.NotEqual(t, red, gray)

		assert.NotEqual(t, orange, magenta)
		assert.NotEqual(t, orange, purple)
		assert.NotEqual(t, orange, gray)

		assert.NotEqual(t, magenta, purple)
		assert.NotEqual(t, magenta, gray)

		assert.NotEqual(t, purple, gray)
	})

	t.Run("no-reset-inside-colors", func(t *testing.T) {
		t.Parallel()
		assert.False(t, strings.Contains(green, reset))
		assert.False(t, strings.Contains(yell, reset))
		assert.False(t, strings.Contains(red, reset))
		assert.False(t, strings.Contains(orange, reset))
		assert.False(t, strings.Contains(magenta, reset))
		assert.False(t, strings.Contains(purple, reset))
		assert.False(t, strings.Contains(gray, reset))
	})
}
