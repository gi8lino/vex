package processor

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// ProcessInPlace performs safe in-place substitution on a file.
func (p *Processor) ProcessInPlace(path string, ioBufSize int) error {
	src, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() { _ = src.Close() }()

	st, err := src.Stat()
	if err != nil {
		return fmt.Errorf("stat: %w", err)
	}

	// create a temporary file in same dir
	dir := filepath.Dir(path)
	base := filepath.Base(path)

	tmp, err := os.CreateTemp(dir, "."+base+".vex-*")
	if err != nil {
		return err
	}
	defer func() { _ = tmp.Close() }() // safety net; we also close explicitly before rename

	cleanup := func() { _ = os.Remove(tmp.Name()) }

	// match permissions of original
	if err := os.Chmod(tmp.Name(), st.Mode()); err != nil {
		cleanup()
		return err
	}

	// stream process src → tmp
	bw := bufio.NewWriterSize(tmp, ioBufSize)

	if err := p.ProcessStream(path, src, bw); err != nil {
		cleanup()
		return err
	}

	// Ensure data hits disk before rename
	if err := tmp.Sync(); err != nil {
		cleanup()
		return err
	}
	// Close the temp file before rename (important on Windows)
	if err := tmp.Close(); err != nil {
		cleanup()
		return err
	}

	// create backup if requested (best-effort)
	if ext := p.opts.BackupExt; ext != "" {
		bak := path + ext
		_ = os.Remove(bak) // remove old backup
		if err := os.Link(path, bak); err != nil {
			_ = copyFile(path, bak, st.Mode()) // fallback to copy (ignore error)
		}
	}

	// Close source before rename (safer on Windows when replacing)
	_ = src.Close()

	// atomic replace
	if err := os.Rename(tmp.Name(), path); err != nil {
		cleanup()
		return err
	}

	// fsync the directory for durability (best-effort)
	if df, derr := os.Open(dir); derr == nil {
		_ = df.Sync()
		_ = df.Close()
	}

	// preserve modtime (best-effort)
	_ = os.Chtimes(path, time.Now(), st.ModTime())

	return nil
}

// copyFile duplicates file contents with the given mode.
// Used as a fallback when hard-linking backups fails.
func copyFile(src, dst string, mode os.FileMode) error {
	s, err := os.Open(src)
	if err != nil {
		return err
	}
	defer s.Close() // nolint:errcheck
	d, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, mode)
	if err != nil {
		return err
	}
	defer d.Close() // nolint:errcheck
	_, err = io.Copy(d, s)
	return err
}
