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
		p := &Processor{
			Formatter: formatter.NewFormatter(false),
			Opts: flag.Options{
				BackupExt: ".bak",
				Colored:   false,
			},
			Lookup: func(name string) (string, bool) {
				if name == "NAME" {
					return "Ada", true
				}
				return "", false
			},
		}

		require.NoError(t, p.ProcessInPlace(path))

		// 1) Original file now contains expanded content.
		got, err := os.ReadFile(path)
		require.NoError(t, err)
		assert.Equal(t, "hello Ada", string(got))

		// 2) Mode preserved.
		info, err := os.Stat(path)
		require.NoError(t, err)
		assert.Equal(t, origMode, info.Mode().Perm())

		// 3) ModTime preserved (equal to the original).
		assert.True(t, info.ModTime().Equal(oldTime), "modtime should be preserved")

		// 4) Backup exists and contains the *original* content.
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

		p := &Processor{
			Formatter: formatter.NewFormatter(false),
			Opts: flag.Options{
				BackupExt: ".bak",
				Colored:   false,
			},
			Lookup: func(name string) (string, bool) {
				return "Ada", true
			},
		}

		require.NoError(t, p.ProcessInPlace(path))

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

		p := &Processor{
			Formatter: formatter.NewFormatter(false),
			Opts:      flag.Options{Colored: false},
			Lookup: func(name string) (string, bool) {
				// unset -> triggers error
				return "", false
			},
		}

		err := p.ProcessInPlace(path)
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

	t.Run("Open errors are wrapped as IO", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		nonexistent := filepath.Join(dir, "does not exist.txt")

		p := &Processor{
			Formatter: formatter.NewFormatter(false),
			Opts:      flag.Options{},
		}

		err := p.ProcessInPlace(nonexistent)
		require.Error(t, err)
		assert.EqualError(t, err, "open "+nonexistent+": no such file or directory")
	})

	t.Run("Is used by ProcessInPlace", func(t *testing.T) {
		t.Parallel()

		// Indirect verification: process a file that uses assignment.
		// Ensures ProcessStream’s Setenv hook is wired through.
		dir := t.TempDir()
		path := filepath.Join(dir, "file.txt")
		require.NoError(t, os.WriteFile(path, []byte("${NEW:=value}"), 0o600))

		var calls int
		var gotName, gotVal string

		p := &Processor{
			Formatter: formatter.NewFormatter(false),
			Opts:      flag.Options{},
			Lookup: func(name string) (string, bool) {
				return "", false // unset to trigger :=
			},
			Setenv: func(name, val string) error {
				calls++
				gotName, gotVal = name, val
				return nil
			},
		}

		require.NoError(t, p.ProcessInPlace(path))

		out, err := os.ReadFile(path)
		require.NoError(t, err)
		assert.Equal(t, "value", string(out))
		assert.Equal(t, 1, calls)
		assert.Equal(t, "NEW", gotName)
		assert.Equal(t, "value", gotVal)
	})
	t.Run("Backup copy fallback when link fails", func(t *testing.T) {
		t.Parallel()

		// This test exercises the fallback path indirectly:
		// We can't reliably force os.Link to fail across platforms, so we simulate
		// by creating the backup as a directory (so Link will fail) and ensuring
		// ProcessInPlace still leaves a readable backup (copy fallback).
		// If creating a directory with the backup name fails, we skip.
		dir := t.TempDir()
		path := filepath.Join(dir, "file.txt")
		require.NoError(t, os.WriteFile(path, []byte("X=${X:-x}"), 0o600))

		bak := path + ".bak"
		// Create a directory where the backup file should go; os.Link will fail with EEXIST or EPERM.
		require.NoError(t, os.Mkdir(bak, 0o700))

		p := &Processor{
			Formatter: formatter.NewFormatter(false),
			Opts: flag.Options{
				BackupExt: ".bak",
			},
			Lookup: func(name string) (string, bool) { return "y", true },
		}

		// Even if linking fails, fallback copy should succeed by removing and re-creating.
		require.NoError(t, p.ProcessInPlace(path))

		// Backup should now be a file, not a directory (the code removes old backup before creating).
		info, err := os.Stat(bak)
		require.NoError(t, err)
		assert.False(t, info.IsDir())

		// And it should contain the *original* content prior to processing.
		bs, err := os.ReadFile(bak)
		require.NoError(t, err)
		assert.Equal(t, "X=${X:-x}", string(bs))
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
