package formatter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPlainFormatter_OkStr(t *testing.T) {
	t.Parallel()
	f := plainFormatter{}

	t.Run("basic", func(t *testing.T) {
		t.Parallel()
		got := f.OkStr("ok")
		assert.Equal(t, "ok", got)
	})

	t.Run("empty", func(t *testing.T) {
		t.Parallel()
		got := f.OkStr("")
		assert.Equal(t, "", got)
	})
}

func TestPlainFormatter_DefaultStr(t *testing.T) {
	t.Parallel()
	f := plainFormatter{}

	t.Run("basic", func(t *testing.T) {
		t.Parallel()
		got := f.DefaultStr("def")
		assert.Equal(t, "def", got)
	})

	t.Run("empty", func(t *testing.T) {
		t.Parallel()
		got := f.DefaultStr("")
		assert.Equal(t, "", got)
	})
}

func TestPlainFormatter_ErrorStr(t *testing.T) {
	t.Parallel()
	f := plainFormatter{}

	t.Run("basic", func(t *testing.T) {
		t.Parallel()
		got := f.ErrorStr("miss")
		assert.Equal(t, "miss", got)
	})

	t.Run("empty", func(t *testing.T) {
		t.Parallel()
		got := f.ErrorStr("")
		assert.Equal(t, "", got)
	})
}

func TestPlainFormatter_UserErrorStr(t *testing.T) {
	t.Parallel()
	f := plainFormatter{}

	t.Run("basic", func(t *testing.T) {
		t.Parallel()
		got := f.UserErrorStr("err")
		assert.Equal(t, "err", got)
	})

	t.Run("empty", func(t *testing.T) {
		t.Parallel()
		got := f.UserErrorStr("")
		assert.Equal(t, "", got)
	})
}

func TestPlainFormatter_FilterStr(t *testing.T) {
	t.Parallel()
	f := plainFormatter{}

	t.Run("basic", func(t *testing.T) {
		t.Parallel()
		got := f.FilterStr("filter")
		assert.Equal(t, "filter", got)
	})

	t.Run("empty", func(t *testing.T) {
		t.Parallel()
		got := f.FilterStr("")
		assert.Equal(t, "", got)
	})
}

func TestPlainFormatter_UnsetStr(t *testing.T) {
	t.Parallel()
	f := plainFormatter{}

	t.Run("basic", func(t *testing.T) {
		t.Parallel()
		got := f.UnsetStr("unset")
		assert.Equal(t, "unset", got)
	})

	t.Run("empty", func(t *testing.T) {
		t.Parallel()
		got := f.UnsetStr("")
		assert.Equal(t, "", got)
	})
}

func TestPlainFormatter_EmptyStr(t *testing.T) {
	t.Parallel()
	f := plainFormatter{}

	t.Run("basic", func(t *testing.T) {
		t.Parallel()
		got := f.EmptyStr("empty")
		assert.Equal(t, "empty", got)
	})

	t.Run("empty", func(t *testing.T) {
		t.Parallel()
		got := f.EmptyStr("")
		assert.Equal(t, "", got)
	})
}
