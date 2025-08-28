package app_test

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/gi8lino/vex/internal/app"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRun(t *testing.T) {
	t.Parallel()

	t.Run("prints help", func(t *testing.T) {
		t.Parallel()
		var out bytes.Buffer
		err := app.Run("1.2.3", "abc123", []string{"--help"}, &out, strings.NewReader(""), nil, nil)
		require.NoError(t, err)
		expected := `Usage: vex [flags]
Flags:
    -i, --in-place              edit files in place; with no files, stdin->stdout [Group: mode (One Of)]
    -b, --backup BACKUP         when -i, create a backup with this extension (e.g. .bak) (Requires: in-place)
        --no-ops                treat operator forms as literals (envsubst-compatible mode)
    -l, --literal-dollar        treat \$ as two bytes (disable dollar-escape)
    -x, --strict                exit on unset or empty (equivalent to --error-unset --error-empty)
    -u, --error-unset           error if a variable is unset
    -e, --error-empty           error if a substitution resolves to empty
    -K, --keep-vars             leave all ${VAR} literals (implies --keep-unset --keep-empty)
    -U, --keep-unset            leave ${VAR} literal if unset
    -E, --keep-empty            leave ${VAR} literal if empty
    -p, --prefix PREFIX...      only replace variables that match any of these prefixes
    -s, --suffix SUFFIX...      only replace variables that match any of these suffixes
    -v, --variable VARIABLE...  only replace variables with these exact names
    -c, --colored               colorize formatter (content and diagnostics) [Group: mode (One Of)]
    -e, --extra-vars PATH...    read variables from file (can be repeated)
    -h, --help                  show help
        --version               show version

`
		assert.Equal(t, expected, out.String())
	})

	t.Run("prints version", func(t *testing.T) {
		t.Parallel()
		var out bytes.Buffer
		err := app.Run("1.2.3", "abc123", []string{"--version"}, &out, strings.NewReader(""), nil, nil)
		require.NoError(t, err)
		assert.Equal(t, "1.2.3\n", out.String())
	})

	t.Run("unknown flag returns subst error", func(t *testing.T) {
		t.Parallel()
		var out bytes.Buffer
		err := app.Run("1.2.3", "abc123", []string{"--definitely-not-a-flag"}, &out, strings.NewReader(""), nil, nil)
		require.Error(t, err)
		assert.EqualError(t, err, "unknown flag: --definitely-not-a-flag")
	})

	t.Run("success passthrough", func(t *testing.T) {
		t.Parallel()
		input := "hello world"
		var out bytes.Buffer

		lookupEnv := func(string) (string, bool) { return "", false }

		err := app.Run("v", "c", []string{}, &out, strings.NewReader(input), lookupEnv, nil)
		require.NoError(t, err)
		assert.Equal(t, input, out.String())
	})

	t.Run("subst error is classified", func(t *testing.T) {
		t.Parallel()
		var out bytes.Buffer

		lookupEnv := func(string) (string, bool) { return "", false }
		err := app.Run("v", "c", []string{}, &out, strings.NewReader("${VAR?boom}"), lookupEnv, nil)
		require.Error(t, err)
		assert.EqualError(t, err, "VAR: boom")
	})

	t.Run("success with rewrite", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		p := filepath.Join(dir, "f.txt")
		require.NoError(t, os.WriteFile(p, []byte("hi ${NAME}"), 0o600))

		lookupEnv := func(name string) (string, bool) {
			if name == "NAME" {
				return "Ada", true
			}
			return "", false
		}

		var out bytes.Buffer
		err := app.Run("v", "c", []string{"-i", "--backup", ".bak", p}, &out, strings.NewReader(""), lookupEnv, nil)
		require.NoError(t, err)

		got, rerr := os.ReadFile(p)
		require.NoError(t, rerr)
		assert.Equal(t, "hi Ada", string(got))

		// backup exists with original content
		bs, berr := os.ReadFile(p + ".bak")
		require.NoError(t, berr)
		assert.Equal(t, "hi ${NAME}", string(bs))
	})

	t.Run("subst error is classified and paths in error", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		p := filepath.Join(dir, "f.txt")
		require.NoError(t, os.WriteFile(p, []byte("${X?boom}"), 0o600))

		lookupEnv := func(string) (string, bool) { return "", false }

		var out bytes.Buffer
		err := app.Run("v", "c", []string{"-i", p}, &out, strings.NewReader(""), lookupEnv, nil)
		require.Error(t, err)
		assert.EqualError(t, err, "X: boom")
	})

	t.Run("io error open missing file is classified", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		missing := filepath.Join(dir, "nope.txt")

		lookupEnv := func(string) (string, bool) { return "", false }

		var out bytes.Buffer
		err := app.Run("v", "c", []string{"-i", missing}, &out, strings.NewReader(""), lookupEnv, nil)
		require.Error(t, err)

		msg := err.Error()
		assert.Contains(t, msg, missing+": ", msg) // inner xerr.Path prefix
		if isWindows(t) {
			assert.Contains(t, msg, "open "+missing+": ", msg) // op + path
			assert.Contains(t, msg, "The system cannot find the file specified.", msg)
		} else {
			assert.Contains(t, msg, "open "+missing+": ", msg) // op + path
			assert.Contains(t, msg, "no such file or directory", msg)
		}
	})

	t.Run("success two files concatenated", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		f1 := filepath.Join(dir, "a.txt")
		f2 := filepath.Join(dir, "b.txt")
		require.NoError(t, os.WriteFile(f1, []byte("A=${A:-a}\n"), 0o600))
		require.NoError(t, os.WriteFile(f2, []byte("B=${B:-b}\n"), 0o600))

		lookupEnv := func(name string) (string, bool) {
			switch name {
			case "A":
				return "", false // default to 'a'
			case "B":
				return "bee", true
			default:
				return "", false
			}
		}
		var out bytes.Buffer
		err := app.Run("v", "c", []string{f1, f2}, &out, strings.NewReader(""), lookupEnv, nil)
		require.NoError(t, err)
		// First file: default a; second file: B expands to bee
		assert.Equal(t, "A=a\nB=bee\n", out.String())
	})

	t.Run("open error is classified as io", func(t *testing.T) {
		t.Parallel()
		// Use a non-existent file path
		dir := t.TempDir()
		missing := filepath.Join(dir, "missing.txt")

		lookupEnv := func(string) (string, bool) { return "", false }

		var out bytes.Buffer
		err := app.Run("v", "c", []string{missing}, &out, strings.NewReader(""), lookupEnv, nil)
		require.Error(t, err)
		assert.EqualError(t, err, "open "+missing+": no such file or directory")
	})

	t.Run("process error is classified as subst", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		f := filepath.Join(dir, "bad.txt")
		require.NoError(t, os.WriteFile(f, []byte("${VAR?boom}"), 0o600))

		lookupEnv := func(string) (string, bool) { return "", false }

		var out bytes.Buffer
		err := app.Run("v", "c", []string{f}, &out, strings.NewReader(""), lookupEnv, nil)
		require.Error(t, err)
		assert.EqualError(t, err, "VAR: boom")
	})

	t.Run("variables at line start and line end", func(t *testing.T) {
		t.Parallel()
		in := `${A:-X} rest
prefix ${B:-Y}
end-is-var ${C:-Z}`
		want := "X rest\nprefix Y\nend-is-var Z"

		lookupEnv := func(string) (string, bool) { return "", false }

		var out bytes.Buffer
		err := app.Run("v", "c", []string{}, &out, strings.NewReader(in), lookupEnv, nil)
		require.NoError(t, err)
		assert.Equal(t, want, out.String())
	})

	t.Run("nested substitution in word", func(t *testing.T) {
		t.Parallel()
		in := `${GREETING:-hi ${NAME}}`
		var out bytes.Buffer

		lookupEnv := func(n string) (string, bool) {
			if n == "NAME" {
				return "Ada", true
			}
			return "", false
		}
		err := app.Run("v", "c", []string{}, &out, strings.NewReader(in), lookupEnv, nil)
		require.NoError(t, err)
		assert.Equal(t, "hi Ada", out.String())
	})

	t.Run("unterminated brace kept literally", func(t *testing.T) {
		t.Parallel()
		in := "${A:-x"
		var out bytes.Buffer
		lookupEnv := func(string) (string, bool) { return "", false }
		err := app.Run("v", "c", []string{}, &out, strings.NewReader(in), lookupEnv, nil)
		require.NoError(t, err)
		assert.Equal(t, in, out.String())
	})

	t.Run("escaped dollar enabled default formatters literal dollar", func(t *testing.T) {
		t.Parallel()
		in := `\$NAME end`
		var out bytes.Buffer
		lookupEnv := func(n string) (string, bool) {
			if n == "NAME" {
				return "SHOULD_NOT_APPEAR", true
			}
			return "", false
		}
		err := app.Run("v", "c", []string{}, &out, strings.NewReader(in), lookupEnv, nil)
		require.NoError(t, err)
		assert.Equal(t, "$NAME end", out.String())
	})

	t.Run("escaped dollar with literal-dollar flag expands after backslash", func(t *testing.T) {
		t.Parallel()
		in := `\$NAME`
		var out bytes.Buffer
		lookupEnv := func(n string) (string, bool) {
			if n == "NAME" {
				return "Ada", true
			}
			return "", false
		}
		err := app.Run("v", "c", []string{"--literal-dollar"}, &out, strings.NewReader(in), lookupEnv, nil)
		require.NoError(t, err)
		// Backslash is literal, $NAME expands → \Ada
		assert.Equal(t, `\Ada`, out.String())
	})

	t.Run("large stdin many lines", func(t *testing.T) {
		t.Parallel()
		var b strings.Builder
		var want strings.Builder
		for range 2000 {
			b.WriteString("V=${V:-v}\n")
			want.WriteString("V=v\n")
		}
		lookupEnv := func(string) (string, bool) { return "", false }
		var out bytes.Buffer
		err := app.Run("v", "c", []string{}, &out, strings.NewReader(b.String()), lookupEnv, nil)
		require.NoError(t, err)
		assert.Equal(t, want.String(), out.String())
	})

	t.Run("strict errors on empty value", func(t *testing.T) {
		t.Parallel()
		in := "${EMPTY}"
		var out bytes.Buffer
		lookupEnv := func(n string) (string, bool) { return "", true }
		err := app.Run("v", "c", []string{"--strict"}, &out, strings.NewReader(in), lookupEnv, nil)
		require.Error(t, err)
		assert.EqualError(t, err, "substitution empty: ${EMPTY}")
	})
}

func TestRun_ComplexFileScenarios(t *testing.T) {
	t.Parallel()

	t.Run("error in second file stops and classifies", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		p := filepath.Join(dir, "ok.txt")
		content := `
default=${Z:-default}
  $B
$$C should not be expanded
${broken but okay


`
		wanted := `
default=default
  B
$$C should not be expanded
${broken but okay


`
		require.NoError(t, os.WriteFile(p, []byte(content), 0o600))

		var out bytes.Buffer
		lookupEnv := func(n string) (string, bool) {
			switch n {
			case "A":
				return "A", true
			case "B":
				return "B", true
			default:
				return "", false
			}
		}

		err := app.Run("v", "c", []string{p}, &out, strings.NewReader(""), lookupEnv, nil)
		require.NoError(t, err)
		assert.Equal(t, wanted, out.String())
	})

	t.Run("concat large files", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		f1 := filepath.Join(dir, "a.txt")
		f2 := filepath.Join(dir, "b.txt")

		var aIn, aWant strings.Builder
		for range 1000 {
			aIn.WriteString("A=${A:-a}\n")
			aWant.WriteString("A=a\n")
		}
		require.NoError(t, os.WriteFile(f1, []byte(aIn.String()), 0o600))
		require.NoError(t, os.WriteFile(f2, []byte("X=${X:-x}\nY=${Y}\n"), 0o600))

		var out bytes.Buffer
		lookupEnv := func(n string) (string, bool) {
			switch n {
			case "Y":
				return "yee", true
			default:
				return "", false
			}
		}

		err := app.Run("v", "c", []string{f1, f2}, &out, strings.NewReader(""), lookupEnv, nil)
		require.NoError(t, err)
		want := aWant.String() + "X=x\nY=yee\n"
		assert.Equal(t, want, out.String())
	})

	t.Run("filters only expand prefix matches", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		p := filepath.Join(dir, "f.txt")
		in := "one=${FOO_ONE}\ntwo=${BAR_TWO}\n"
		require.NoError(t, os.WriteFile(p, []byte(in), 0o600))

		var out bytes.Buffer
		lookupEnv := func(n string) (string, bool) {
			switch n {
			case "FOO_ONE":
				return "1", true
			case "BAR_TWO":
				return "2", true
			default:
				return "", false
			}
		}

		err := app.Run("v", "c", []string{"--prefix", "FOO_", p}, &out, strings.NewReader(""), lookupEnv, nil)
		require.NoError(t, err)
		// Only FOO_* expands; BAR_* stays literal
		assert.Equal(t, "one=1\ntwo=${BAR_TWO}\n", out.String())
	})

	t.Run("keep flags preserve literals for unset and empty", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		p := filepath.Join(dir, "f.txt")
		in := "u=${UNSET}\ne=${EMPTY}\n"
		require.NoError(t, os.WriteFile(p, []byte(in), 0o600))

		// Use both keep flags via --keep-vars
		var out bytes.Buffer
		lookupEnv := func(n string) (string, bool) {
			if n == "EMPTY" {
				return "", true // set but empty
			}
			return "", false
		}
		err := app.Run("v", "c", []string{"--keep-vars", p}, &out, strings.NewReader(""), lookupEnv, nil)
		require.NoError(t, err)
		assert.Equal(t, "u=${UNSET}\ne=${EMPTY}\n", out.String())
	})

	t.Run("no ops keeps operator forms literal", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		p := filepath.Join(dir, "f.txt")
		in := "val=${A:-x}\n"
		require.NoError(t, os.WriteFile(p, []byte(in), 0o600))

		var out bytes.Buffer
		lookupEnv := func(string) (string, bool) { return "", false }
		err := app.Run("v", "c", []string{"--no-ops", p}, &out, strings.NewReader(""), lookupEnv, nil)
		require.NoError(t, err)
		assert.Equal(t, in, out.String())
	})

	t.Run("error in second file stops and classifies", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		ok := filepath.Join(dir, "ok.txt")
		bad := filepath.Join(dir, "bad.txt")
		require.NoError(t, os.WriteFile(ok, []byte("ok=${A:-ok}\n"), 0o600))
		require.NoError(t, os.WriteFile(bad, []byte("${B?boom}\n"), 0o600))

		var out bytes.Buffer
		lookupEnv := func(string) (string, bool) { return "", false }
		err := app.Run("v", "c", []string{ok, bad}, &out, strings.NewReader(""), lookupEnv, nil)
		require.Error(t, err)
		assert.Equal(t, "ok=ok\n", out.String())

		assert.EqualError(t, err, "B: boom")
	})

	t.Run("Extra vars file errors are classified", func(t *testing.T) {
		t.Parallel()

		err := app.Run("v", "c", []string{"--extra-vars", "/does/not/exist"}, nil, strings.NewReader(""), nil, nil)
		require.Error(t, err)
		assert.EqualError(t, err, "open /does/not/exist: no such file or directory")
	})
}

// Optional skip for Windows edge cases with file locking/rename semantics in CI.
// (The tests above should still pass on Windows; keep this helper if you extend.)
func isWindows(t *testing.T) bool { t.Helper(); return runtime.GOOS == "windows" }
