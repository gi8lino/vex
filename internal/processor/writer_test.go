package processor

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gi8lino/vex/internal/flag"
	"github.com/gi8lino/vex/internal/formatter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProcessInPlace(t *testing.T) {
	t.Parallel()

	t.Run("Success with backup and meta", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		path := filepath.Join(dir, "file.txt")
		orig := "hello ${NAME}"

		// Create original file with content and specific mode + modtime.
		require.NoError(t, os.WriteFile(path, []byte(orig), 0o640))

		origInfo, err := os.Stat(path)
		require.NoError(t, err)
		origMode := origInfo.Mode().Perm()
		// Set a distinct, older mod time to verify it is preserved
		oldTime := time.Now().Add(-3 * time.Hour).Truncate(time.Second)
		require.NoError(t, os.Chtimes(path, oldTime, oldTime))

		// Prepare processor that will expand ${NAME} -> Ada and make a backup.
		p := NewProcessor(
			flag.Options{
				BackupExt: ".bak",
				Colored:   false,
			},
			func(name string) (string, bool) {
				if name == "NAME" {
					return "Ada", true
				}
				return "", false
			},
			nil,
			formatter.NewFormatter(false),
			testBufSize,
		)

		require.NoError(t, p.ProcessInPlace(path, testBufSize))

		// Original file now contains expanded content.
		got, err := os.ReadFile(path)
		require.NoError(t, err)
		assert.Equal(t, "hello Ada", string(got))

		// Mode preserved.
		info, err := os.Stat(path)
		require.NoError(t, err)
		assert.Equal(t, origMode, info.Mode().Perm())

		// ModTime preserved (equal to the original).
		assert.True(t, info.ModTime().Equal(oldTime), "modtime should be preserved")

		// Backup exists and contains the *original* content.
		bak := path + ".bak"
		bs, err := os.ReadFile(bak)
		require.NoError(t, err)
		assert.Equal(t, orig, string(bs))
	})

	t.Run("Success with replacement backup", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		path := filepath.Join(dir, "file.txt")
		require.NoError(t, os.WriteFile(path, []byte("${NAME}"), 0o600))

		// Pre-create an old backup that should be overwritten.
		bak := path + ".bak"
		require.NoError(t, os.WriteFile(bak, []byte("OLD BACKUP"), 0o600))

		p := NewProcessor(
			flag.Options{
				BackupExt: ".bak",
				Colored:   false,
			},
			func(name string) (string, bool) {
				return "Ada", true
			},
			nil,
			formatter.NewFormatter(false),
			testBufSize,
		)

		require.NoError(t, p.ProcessInPlace(path, testBufSize))

		// Backup should now reflect the original pre-processed content ("${NAME}")
		bs, err := os.ReadFile(bak)
		require.NoError(t, err)
		assert.Equal(t, "${NAME}", string(bs))
	})

	t.Run("Propagates subst error and cleans temp", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		path := filepath.Join(dir, "file.txt")
		require.NoError(t, os.WriteFile(path, []byte("${VAR?boom}"), 0o600))

		p := NewProcessor(
			flag.Options{Colored: false},
			func(name string) (string, bool) {
				// unset -> triggers error
				return "", false
			},
			nil,
			formatter.NewFormatter(false),
			testBufSize,
		)

		err := p.ProcessInPlace(path, testBufSize)
		require.Error(t, err)
		assert.EqualError(t, err, "VAR: boom")

		// Ensure no lingering temp files matching .<base>.vex-*
		base := filepath.Base(path)
		entries, listErr := os.ReadDir(dir)
		require.NoError(t, listErr)
		for _, e := range entries {
			require.Falsef(t, strings.HasPrefix(e.Name(), "."+base+".vex-"),
				"temporary file should be cleaned up, found: %s", e.Name())
		}
	})

	t.Run("Open errors are surfaced as I/O errors", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		nonexistent := filepath.Join(dir, "does not exist.txt")

		p := NewProcessor(
			flag.Options{},
			nil,
			nil,
			formatter.NewFormatter(false),
			testBufSize,
		)

		err := p.ProcessInPlace(nonexistent, testBufSize)
		require.Error(t, err)
		assert.EqualError(t, err, "open "+nonexistent+": no such file or directory")
	})

	t.Run("Setenv hook is used during in-place processing", func(t *testing.T) {
		t.Parallel()

		// Indirect verification: process a file that uses assignment.
		// Ensures ProcessStream’s Setenv hook is wired through.
		dir := t.TempDir()
		path := filepath.Join(dir, "file.txt")
		require.NoError(t, os.WriteFile(path, []byte("${NEW:=value}"), 0o600))

		var calls int
		var gotName, gotVal string

		p := NewProcessor(
			flag.Options{},
			func(name string) (string, bool) {
				return "", false // unset to trigger :=
			},
			func(name, val string) error {
				calls++
				gotName, gotVal = name, val
				return nil
			},
			formatter.NewFormatter(false),
			testBufSize,
		)

		require.NoError(t, p.ProcessInPlace(path, testBufSize))

		out, err := os.ReadFile(path)
		require.NoError(t, err)
		assert.Equal(t, "value", string(out))
		assert.Equal(t, 1, calls)
		assert.Equal(t, "NEW", gotName)
		assert.Equal(t, "value", gotVal)
	})

	t.Run("Backup is created (link or copy) and contains original", func(t *testing.T) {
		t.Parallel()

		// Cross-platform validation that regardless of whether hard-linking
		// is supported, a backup file ends up with the *original* contents.
		dir := t.TempDir()
		path := filepath.Join(dir, "file.txt")
		orig := "X=${X:-x}"
		require.NoError(t, os.WriteFile(path, []byte(orig), 0o600))

		p := NewProcessor(
			flag.Options{BackupExt: ".bak"},
			func(name string) (string, bool) { return "y", true },
			nil,
			formatter.NewFormatter(false),
			testBufSize,
		)

		require.NoError(t, p.ProcessInPlace(path, testBufSize))

		bak := path + ".bak"
		info, err := os.Stat(bak)
		require.NoError(t, err)
		assert.False(t, info.IsDir())

		// And it should contain the *original* content prior to processing.
		bs, err := os.ReadFile(bak)
		require.NoError(t, err)
		assert.Equal(t, orig, string(bs))
	})
}

func TestCopyFile(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		dir := t.TempDir()
		src := filepath.Join(dir, "src.txt")
		dst := filepath.Join(dir, "dst.txt")

		data := []byte("copy me")
		require.NoError(t, os.WriteFile(src, data, 0o640))

		// Explicit mode for dst
		require.NoError(t, copyFile(src, dst, 0o640))

		// Contents equal
		got, err := os.ReadFile(dst)
		require.NoError(t, err)
		assert.Equal(t, data, got)

		// Mode as requested
		info, err := os.Stat(dst)
		require.NoError(t, err)
		assert.Equal(t, fs.FileMode(0o640), info.Mode().Perm())
	})

	t.Run("Source open error is IO", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		src := filepath.Join(dir, "missing.txt")
		dst := filepath.Join(dir, "dst.txt")

		err := copyFile(src, dst, 0o600)
		require.Error(t, err)
		assert.ErrorIs(t, err, os.ErrNotExist)
		assert.Contains(t, err.Error(), src+": ")
	})

	t.Run("Dest open error is IO", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		src := filepath.Join(dir, "src.txt")
		require.NoError(t, os.WriteFile(src, []byte("x"), 0o400))

		// Create a directory where we try to open a file -> will fail
		dstDir := filepath.Join(dir, "d")
		require.NoError(t, os.MkdirAll(dstDir, 0o700))
		dst := filepath.Join(dstDir, "sub/inner.txt") // parent doesn't exist -> OpenFile should error

		err := copyFile(src, dst, 0o600)
		require.Error(t, err)
		assert.ErrorIs(t, err, os.ErrNotExist)
		assert.Contains(t, err.Error(), dst+": ")
	})
}
