package fsm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTransformCase(t *testing.T) {
	t.Parallel()

	t.Run("caret upper first", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "Hello", transformCase("^", "hello"))
	})

	t.Run("double caret upper all", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "ABCXYZ", transformCase("^^", "AbcXyZ"))
	})

	t.Run("comma lower first", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "hello", transformCase(",", "Hello"))
	})

	t.Run("double comma lower all", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "abcxyz", transformCase(",,", "AbcXyZ"))
	})

	t.Run("unknown op returns input", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "Hello", transformCase("~", "Hello"))
	})

	t.Run("empty input safe", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "", transformCase("^", ""))
		assert.Equal(t, "", transformCase(",", ""))
	})
}

func TestSubstr(t *testing.T) {
	t.Parallel()

	t.Run("offset only", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "cdef", substr("2", "abcdef"))
	})

	t.Run("offset and len", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "bcd", substr("1:3", "abcdef"))
	})

	t.Run("negative offset from end", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "ef", substr("-2", "abcdef"))
	})

	t.Run("negative offset with len", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "de", substr("-3:2", "abcdef"))
	})

	t.Run("offset beyond end clamps to end", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "", substr("100:5", "abc"))
	})

	t.Run("offset before start clamps to zero", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "abc", substr("-100:10", "abc"))
	})

	t.Run("zero length allowed", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "", substr("0:0", "abc"))
	})

	t.Run("negative length results empty", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "", substr("2:-5", "abcdef"))
	})
}

func TestParseOffsetLen(t *testing.T) {
	t.Parallel()

	t.Run("offset only", func(t *testing.T) {
		t.Parallel()
		off, length, hasLen := parseOffsetLen("2")
		assert.Equal(t, 2, off)
		assert.Equal(t, 0, length)
		assert.False(t, hasLen)
	})

	t.Run("offset and len", func(t *testing.T) {
		t.Parallel()
		off, length, hasLen := parseOffsetLen("1:3")
		assert.Equal(t, 1, off)
		assert.Equal(t, 3, length)
		assert.True(t, hasLen)
	})

	t.Run("negative offset and len", func(t *testing.T) {
		t.Parallel()
		off, length, hasLen := parseOffsetLen("-2:4")
		assert.Equal(t, -2, off)
		assert.Equal(t, 4, length)
		assert.True(t, hasLen)
	})

	t.Run("empty len is zero but present", func(t *testing.T) {
		t.Parallel()
		off, length, hasLen := parseOffsetLen("10:")
		assert.Equal(t, 10, off)
		assert.Equal(t, 0, length)
		assert.True(t, hasLen)
	})

	t.Run("both empty defaults to zero with len present", func(t *testing.T) {
		t.Parallel()
		off, length, hasLen := parseOffsetLen(":")
		assert.Equal(t, 0, off)
		assert.Equal(t, 0, length)
		assert.True(t, hasLen)
	})
}

func TestAtoiSafe(t *testing.T) {
	t.Parallel()

	t.Run("positive", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, 123, atoiSafe("123"))
	})

	t.Run("negative", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, -45, atoiSafe("-45"))
	})

	t.Run("stops on non digit", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, 12, atoiSafe("12x34"))
	})

	t.Run("empty string zero", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, 0, atoiSafe(""))
	})

	t.Run("just minus zero", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, 0, atoiSafe("-"))
	})
}

func TestTrimPrefixAll(t *testing.T) {
	t.Parallel()

	t.Run("removes repeated prefixes", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "d", trimPrefixAll("aaad", "a"))
	})

	t.Run("empty pattern no change", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "abc", trimPrefixAll("abc", ""))
	})
}

func TestTrimSuffixAll(t *testing.T) {
	t.Parallel()

	t.Run("removes repeated suffixes", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "b", trimSuffixAll("baaa", "a"))
	})

	t.Run("empty pattern no change", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "abc", trimSuffixAll("abc", ""))
	})
}
