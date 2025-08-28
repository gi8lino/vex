package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type errReader struct{}

func (errReader) Read(p []byte) (int, error) {
	return 0, fmt.Errorf("boom")
}

func TestReadVars(t *testing.T) {
	t.Parallel()

	t.Run("Empty input returns empty map", func(t *testing.T) {
		t.Parallel()
		m, err := readVars(strings.NewReader(""))
		require.NoError(t, err)
		assert.Equal(t, map[string]string{}, m)
	})

	t.Run("Ignore comments and blank lines", func(t *testing.T) {
		t.Parallel()
		input := "\n# comment\nFOO=bar\n  \n# another\nBAZ = qux  \n"
		m, err := readVars(strings.NewReader(input))
		require.NoError(t, err)
		assert.Equal(t, map[string]string{"FOO": "bar", "BAZ": "qux"}, m)
	})

	t.Run("Value may contain equals", func(t *testing.T) {
		t.Parallel()
		input := "KEY=a=b=c\n"
		m, err := readVars(strings.NewReader(input))
		require.NoError(t, err)
		assert.Equal(t, map[string]string{"KEY": "a=b=c"}, m)
	})

	t.Run("Invalid line without equals", func(t *testing.T) {
		t.Parallel()
		input := "NOTVALID\n"
		_, err := readVars(strings.NewReader(input))
		require.Error(t, err)
		assert.EqualError(t, err, `invalid var line "NOTVALID" (expected KEY=VALUE)`)
	})

	t.Run("Scanner error is returned", func(t *testing.T) {
		t.Parallel()
		_, err := readVars(errReader{})
		require.Error(t, err)
		assert.EqualError(t, err, "boom")
	})
}

func TestMergeVars(t *testing.T) {
	t.Parallel()

	t.Run("Single file overrides fallback", func(t *testing.T) {
		t.Parallel()
		path := writeTempFile(t, "FOO=from_file\n")

		fallback := func(k string) (string, bool) {
			if k == "FOO" {
				return "from_fallback", true
			}
			return "", false
		}

		lookup, err := MergeVars([]string{path}, fallback)
		require.NoError(t, err)

		v, ok := lookup("FOO")
		require.True(t, ok)
		assert.Equal(t, "from_file", v)

		// Key only in fallback
		v2, ok2 := lookup("ONLY_FALLBACK")
		assert.False(t, ok2)
		assert.Equal(t, "", v2)
	})

	t.Run("Later files override earlier", func(t *testing.T) {
		t.Parallel()
		p1 := writeTempFile(t, "A=1\nB=2\n")
		p2 := writeTempFile(t, "B=22\nC=3\n")

		lookup, err := MergeVars([]string{p1, p2}, nil)
		require.NoError(t, err)

		vA, okA := lookup("A")
		require.True(t, okA)
		assert.Equal(t, "1", vA)

		vB, okB := lookup("B")
		require.True(t, okB)
		assert.Equal(t, "22", vB)

		vC, okC := lookup("C")
		require.True(t, okC)
		assert.Equal(t, "3", vC)
	})

	t.Run("Missing file returns enriched error", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		missing := filepath.Join(dir, "does-not-exist.env")

		// Build expected message by reproducing the underlying os.Open error.
		_, openErr := os.Open(missing)
		require.Error(t, openErr)

		_, err := MergeVars([]string{missing}, nil)
		require.Error(t, err)
		assert.EqualError(t, err, openErr.Error())
	})

	t.Run("Parse error is wrapped with file context", func(t *testing.T) {
		t.Parallel()
		path := writeTempFile(t, "GOOD=ok\nBADLINE\n")

		_, err := MergeVars([]string{path}, nil)
		require.Error(t, err)

		expected := fmt.Sprintf(`parsing vars in %q: invalid var line "BADLINE" (expected KEY=VALUE)`, path)
		assert.EqualError(t, err, expected)
	})

	t.Run("Lookup falls back when not in files", func(t *testing.T) {
		t.Parallel()
		path := writeTempFile(t, "INFILE=1\n")

		fallback := func(k string) (string, bool) {
			if k == "FALLBACK" {
				return "fb", true
			}
			return "", false
		}

		lookup, err := MergeVars([]string{path}, fallback)
		require.NoError(t, err)

		v1, ok1 := lookup("INFILE")
		require.True(t, ok1)
		assert.Equal(t, "1", v1)

		v2, ok2 := lookup("FALLBACK")
		require.True(t, ok2)
		assert.Equal(t, "fb", v2)

		v3, ok3 := lookup("MISSING")
		assert.False(t, ok3)
		assert.Equal(t, "", v3)
	})
}

func writeTempFile(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "vars.env")
	err := os.WriteFile(path, []byte(content), 0o600)
	require.NoError(t, err)
	return path
}
