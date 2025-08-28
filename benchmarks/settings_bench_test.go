package bench_test

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

const (
	BenchSmallFileSize = 64 * 1024       // 64 KiB
	BenchBigFileSize   = 5 * 1024 * 1024 // 5 MiB
	BenchNumSmallFiles = 200
	BenchNumBigFiles   = 10
)

var (
	BenchEnvsubstPattern = "A=$A B=$B LITERAL-0123456789\n"
	BenchExtendedPattern = "A=${A:-a} B=${B} LITERAL-0123456789\n"
)

// For CLI fairness when desired. NOTE: we now use a CLEAN environment by default in runStreaming.
var BenchEnvVars = []string{"A=a", "B=bee"}

// --- binaries (overridable) ---

func envsubstBin() string {
	if s := os.Getenv("ENV_BIN"); s != "" {
		return s
	}
	return "envsubst"
}

func renvsubstBin() string {
	if s := os.Getenv("RENV_BIN"); s != "" {
		return s
	}
	return "./bin/renvsubst"
}

func vexBin() string {
	if s := os.Getenv("VEX_BIN"); s != "" {
		return s
	}
	return "vex"
}

func vexArgs() []string {
	if s := os.Getenv("VEX_ARGS"); s != "" {
		return strings.Fields(s)
	}
	return nil
}

// --- I/O helpers ---

// writeRepeating writes 'pattern' repeatedly until >= targetSize bytes.
func writeRepeating(path, pattern string, targetSize int) (int64, error) {
	f, err := os.Create(path)
	if err != nil {
		return 0, err
	}
	defer f.Close() // nolint:errcheck

	var written int64
	for written < int64(targetSize) {
		n, err := io.WriteString(f, pattern)
		if err != nil {
			return written, err
		}
		written += int64(n)
	}
	return written, f.Sync()
}

// makeFiles creates n files of roughly eachBytes, filled with pattern.
func makeFiles(b *testing.B, dir string, n int, eachBytes int, pattern string) ([]string, int64) {
	paths := make([]string, 0, n)
	var total int64
	for i := range n {
		p := filepath.Join(dir, "f-"+pad3(i)+".txt")
		w, err := writeRepeating(p, pattern, eachBytes)
		if err != nil {
			b.Fatalf("writeRepeating: %v", err)
		}
		paths = append(paths, p)
		total += w
	}
	return paths, total
}

// pad3 zero-pads i to width 3 (1000+ returns full number unchanged).
func pad3(i int) string {
	s := strconv.Itoa(i)
	if len(s) < 3 {
		return "000"[len(s):] + s
	}
	return s
}

// runStreaming executes an external binary, feeds all files to stdin, discards stdout.
// IMPORTANT: uses a CLEAN environment for determinism (only extraEnv).
func runStreaming(bin string, args []string, files []string, extraEnv []string) error {
	path, err := exec.LookPath(bin)
	if err != nil {
		return err
	}
	stderr := &bytes.Buffer{}
	cmd := exec.Command(path, args...)
	cmd.Env = append([]string{}, extraEnv...) // CLEAN ENVIRONMENT
	cmd.Stdout = io.Discard
	cmd.Stderr = stderr

	in, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		_ = in.Close()
		return err
	}

	for _, p := range files {
		f, err := os.Open(p)
		if err != nil {
			_ = in.Close()
			_ = cmd.Wait()
			return err
		}
		if _, cErr := io.Copy(in, f); cErr != nil {
			_ = f.Close()
			_ = in.Close()
			_ = cmd.Wait()
			return cErr
		}
		_ = f.Close()
	}

	_ = in.Close()
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("command %q failed: %w; stderr=%s", bin, err, stderr.String())
	}
	return nil
}
