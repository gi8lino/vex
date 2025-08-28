package fsm

import (
	"bufio"
	"bytes"
	"testing"

	"github.com/gi8lino/vex/internal/flag"
	"github.com/gi8lino/vex/internal/formatter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExpandWithOp(t *testing.T) {
	t.Parallel()
	const label = "EngineLabel"

	t.Run("NoOps falls back to simple braced expansion", func(t *testing.T) {
		t.Parallel()
		e := &Engine{
			Format: formatter.NewFormatter(false),
			Label:  label,
			Opts:   flag.Options{NoOps: true},
			Lookup: func(name string) (string, bool) { return "val", true },
		}
		out, err := e.expandWithOp("VAR", "##", []byte("a"))
		require.NoError(t, err)
		assert.Equal(t, "val", out)
	})

	t.Run("empty op falls back to simple braced expansion", func(t *testing.T) {
		t.Parallel()
		e := &Engine{
			Format: formatter.NewFormatter(false),
			Label:  label,
			Opts:   flag.Options{},
			Lookup: func(name string) (string, bool) { return "v", true },
		}
		out, err := e.expandWithOp("VAR", "", []byte("ignored"))
		require.NoError(t, err)
		assert.Equal(t, "v", out)
	})

	t.Run("#len returns rune length", func(t *testing.T) {
		t.Parallel()
		e := &Engine{
			Format: formatter.NewFormatter(false),
			Label:  label,
			Lookup: func(name string) (string, bool) { return "你好😊", true }, // 3 runes
		}
		out, err := e.expandWithOp("VAR", "#len", nil)
		require.NoError(t, err)
		assert.Equal(t, "3", out)
	})

	t.Run("trim prefix single #", func(t *testing.T) {
		t.Parallel()
		e := &Engine{
			Format: formatter.NewFormatter(false), Label: label,
			Lookup: func(name string) (string, bool) { return "aaab", true },
		}
		out, err := e.expandWithOp("VAR", "#", []byte("a"))
		require.NoError(t, err)
		assert.Equal(t, "aab", out)
	})

	t.Run("trim prefix double ##", func(t *testing.T) {
		t.Parallel()
		e := &Engine{
			Format: formatter.NewFormatter(false), Label: label,
			Lookup: func(name string) (string, bool) { return "aaab", true },
		}
		out, err := e.expandWithOp("VAR", "##", []byte("a"))
		require.NoError(t, err)
		assert.Equal(t, "b", out)
	})

	t.Run("trim suffix single %", func(t *testing.T) {
		t.Parallel()
		e := &Engine{
			Format: formatter.NewFormatter(false), Label: label,
			Lookup: func(name string) (string, bool) { return "baaa", true },
		}
		out, err := e.expandWithOp("VAR", "%", []byte("a"))
		require.NoError(t, err)
		assert.Equal(t, "baa", out)
	})

	t.Run("trim suffix double %%", func(t *testing.T) {
		t.Parallel()
		e := &Engine{
			Format: formatter.NewFormatter(false), Label: label,
			Lookup: func(name string) (string, bool) { return "baaa", true },
		}
		out, err := e.expandWithOp("VAR", "%%", []byte("a"))
		require.NoError(t, err)
		assert.Equal(t, "b", out)
	})

	t.Run("case upper ^", func(t *testing.T) {
		t.Parallel()
		e := &Engine{
			Format: formatter.NewFormatter(false), Label: label,
			Lookup: func(name string) (string, bool) { return "hello", true },
		}
		out, err := e.expandWithOp("VAR", "^", nil)
		require.NoError(t, err)
		assert.Equal(t, "Hello", out)
	})

	t.Run("case upper ^^", func(t *testing.T) {
		t.Parallel()
		e := &Engine{
			Format: formatter.NewFormatter(false), Label: label,
			Lookup: func(name string) (string, bool) { return "AbcXyZ", true },
		}
		out, err := e.expandWithOp("VAR", "^^", nil)
		require.NoError(t, err)
		assert.Equal(t, "ABCXYZ", out)
	})

	t.Run("case lower ,", func(t *testing.T) {
		t.Parallel()
		e := &Engine{
			Format: formatter.NewFormatter(false), Label: label,
			Lookup: func(name string) (string, bool) { return "Hello", true },
		}
		out, err := e.expandWithOp("VAR", ",", nil)
		require.NoError(t, err)
		assert.Equal(t, "hello", out)
	})

	t.Run("case lower ,,", func(t *testing.T) {
		t.Parallel()
		e := &Engine{
			Format: formatter.NewFormatter(false), Label: label,
			Lookup: func(name string) (string, bool) { return "AbcXyZ", true },
		}
		out, err := e.expandWithOp("VAR", ",,", nil)
		require.NoError(t, err)
		assert.Equal(t, "abcxyz", out)
	})

	t.Run("substr :", func(t *testing.T) {
		t.Parallel()
		e := &Engine{
			Format: formatter.NewFormatter(false), Label: label,
			Lookup: func(name string) (string, bool) { return "abcdef", true },
		}
		out, err := e.expandWithOp("VAR", ":", []byte("1:3"))
		require.NoError(t, err)
		assert.Equal(t, "bcd", out)
	})

	t.Run("replace first /", func(t *testing.T) {
		t.Parallel()
		e := &Engine{
			Format: formatter.NewFormatter(false), Label: label,
			Lookup: func(name string) (string, bool) { return "aa-aa", true },
		}
		out, err := e.expandWithOp("VAR", "/", []byte("aa/X"))
		require.NoError(t, err)
		assert.Equal(t, "X-aa", out)
	})

	t.Run("replace all //", func(t *testing.T) {
		t.Parallel()
		e := &Engine{
			Format: formatter.NewFormatter(false), Label: label,
			Lookup: func(name string) (string, bool) { return "aa-aa", true },
		}
		out, err := e.expandWithOp("VAR", "//", []byte("aa/X"))
		require.NoError(t, err)
		assert.Equal(t, "X-X", out)
	})

	t.Run("quote @Q", func(t *testing.T) {
		t.Parallel()
		e := &Engine{
			Format: formatter.NewFormatter(false), Label: label,
			Lookup: func(name string) (string, bool) { return "a'b", true },
		}
		out, err := e.expandWithOp("VAR", "@", []byte("Q"))
		require.NoError(t, err)
		assert.Equal(t, `'a'"'"'b'`, out)
	})

	t.Run("default - when unset uses word with nested expansion", func(t *testing.T) {
		t.Parallel()
		e := &Engine{
			Format: formatter.NewFormatter(false),
			Label:  label,
			Lookup: func(name string) (string, bool) {
				if name == "X" {
					return "Z", true
				}
				return "", false // VAR is unset
			},
		}
		out, err := e.expandWithOp("VAR", "-", []byte("hi ${X}"))
		require.NoError(t, err)
		assert.Equal(t, "hi Z", out)
	})

	t.Run("default :- when empty/unset uses fallback", func(t *testing.T) {
		t.Parallel()
		e := &Engine{
			Format: formatter.NewFormatter(false), Label: label,
			Lookup: func(name string) (string, bool) { return "", true },
		}
		out, err := e.expandWithOp("VAR", ":-", []byte("fallback"))
		require.NoError(t, err)
		assert.Equal(t, "fallback", out)
	})

	t.Run("assign = when unset calls Setenv", func(t *testing.T) {
		t.Parallel()
		var calls int
		var gotName, gotVal string
		e := &Engine{
			Format: formatter.NewFormatter(false),
			Label:  label,
			Lookup: func(name string) (string, bool) { return "", false },
			Setenv: func(name, val string) error { calls++; gotName, gotVal = name, val; return nil },
		}
		out, err := e.expandWithOp("VAR", "=", []byte("v"))
		require.NoError(t, err)
		assert.Equal(t, "v", out)
		assert.Equal(t, 1, calls)
		assert.Equal(t, "VAR", gotName)
		assert.Equal(t, "v", gotVal)
	})

	t.Run("assign := when empty calls Setenv", func(t *testing.T) {
		t.Parallel()
		var calls int
		var gotName, gotVal string
		e := &Engine{
			Format: formatter.NewFormatter(false),
			Label:  label,
			Lookup: func(name string) (string, bool) { return "", true }, // set but empty
			Setenv: func(name, val string) error { calls++; gotName, gotVal = name, val; return nil },
		}
		out, err := e.expandWithOp("VAR", ":=", []byte("def"))
		require.NoError(t, err)
		assert.Equal(t, "def", out)
		assert.Equal(t, 1, calls)
		assert.Equal(t, "VAR", gotName)
		assert.Equal(t, "def", gotVal)
	})

	t.Run("alt + when set returns word (prefix with name per spec)", func(t *testing.T) {
		t.Parallel()
		e := &Engine{
			Format: formatter.NewFormatter(false), Label: label,
			Lookup: func(name string) (string, bool) { return "v", true },
		}
		out, err := e.expandWithOp("VAR", "+", []byte("word"))
		require.NoError(t, err)
		assert.Equal(t, "VAR: word", out)
	})

	t.Run("alt :+ when notNull returns word", func(t *testing.T) {
		t.Parallel()
		e := &Engine{
			Format: formatter.NewFormatter(false), Label: label,
			Lookup: func(name string) (string, bool) { return "v", true },
		}
		out, err := e.expandWithOp("VAR", ":+", []byte("word"))
		require.NoError(t, err)
		assert.Equal(t, "VAR: word", out)
	})

	t.Run("error ? when unset returns labeled error", func(t *testing.T) {
		t.Parallel()
		e := &Engine{
			Format: formatter.NewFormatter(false), Label: label,
			Lookup: func(name string) (string, bool) { return "", false },
		}
		out, err := e.expandWithOp("VAR", "?", []byte("boom"))
		require.Error(t, err)
		assert.EqualError(t, err, "VAR: boom")
		assert.Equal(t, "", out)
	})

	t.Run("error :? when empty returns labeled error", func(t *testing.T) {
		t.Parallel()
		e := &Engine{
			Format: formatter.NewFormatter(false), Label: label,
			Lookup: func(name string) (string, bool) { return "", true },
		}
		out, err := e.expandWithOp("VAR", ":?", []byte("boom"))
		require.Error(t, err)
		assert.EqualError(t, err, "VAR: boom")
		assert.Equal(t, "", out)
	})

	t.Run("unknown operator keeps literal", func(t *testing.T) {
		t.Parallel()
		e := &Engine{
			Format: formatter.NewFormatter(false), Label: label,
			Lookup: func(name string) (string, bool) { return "v", true },
		}
		out, err := e.expandWithOp("VAR", "~", []byte("x"))
		require.NoError(t, err)
		assert.Equal(t, "${VAR~x}", out)
	})

	t.Run("Fails when unset and FailOnUnset", func(t *testing.T) {
		t.Parallel()
		e := &Engine{
			Format: formatter.NewFormatter(false),
			Opts:   flag.Options{ErrorUnset: true, NoOps: true},
			Lookup: func(string) (string, bool) { return "", false },
		}
		var buf bytes.Buffer
		bw := bufio.NewWriter(&buf)

		out, err := e.expandWithOp("VAR", "~", []byte("x"))
		require.Error(t, err)
		assert.EqualError(t, err, "variable not set: ${VAR}")
		assert.Empty(t, out)
		_ = bw.Flush()
		assert.Equal(t, "", buf.String())
	})
}

func TestExpandSimple(t *testing.T) {
	t.Parallel()

	t.Run("Empty name writes literal", func(t *testing.T) {
		t.Parallel()
		e := &Engine{Format: formatter.NewFormatter(false)}
		var buf bytes.Buffer
		bw := bufio.NewWriter(&buf)

		err := e.expandSimple(bw, BracedRef(""))
		require.NoError(t, err)
		_ = bw.Flush()
		assert.Equal(t, "${}", buf.String())
	})

	t.Run("Filtered name writes missing literal", func(t *testing.T) {
		t.Parallel()
		e := &Engine{
			Format: formatter.NewFormatter(false),
			Opts: flag.Options{
				Variables: []string{"OTHER"}, // filter out NAME
			},
			Lookup: func(string) (string, bool) { return "ignored", true },
		}
		var buf bytes.Buffer
		bw := bufio.NewWriter(&buf)

		err := e.expandSimple(bw, BracedRef("NAME"))
		require.NoError(t, err)
		_ = bw.Flush()
		assert.Equal(t, "${NAME}", buf.String())
	})

	t.Run("Fails when unset and FailOnUnset", func(t *testing.T) {
		t.Parallel()
		e := &Engine{
			Format: formatter.NewFormatter(false),
			Opts:   flag.Options{ErrorUnset: true},
			Lookup: func(string) (string, bool) { return "", false },
		}
		var buf bytes.Buffer
		bw := bufio.NewWriter(&buf)

		err := e.expandSimple(bw, BracedRef("NAME"))
		require.Error(t, err)
		assert.EqualError(t, err, "variable not set: ${NAME}")
		_ = bw.Flush()
		assert.Equal(t, "", buf.String())
	})

	t.Run("Unset and NoReplaceUnset keeps literal", func(t *testing.T) {
		t.Parallel()
		e := &Engine{
			Format: formatter.NewFormatter(false),
			Opts:   flag.Options{KeepUnset: true},
			Lookup: func(string) (string, bool) { return "", false },
		}
		var buf bytes.Buffer
		bw := bufio.NewWriter(&buf)

		err := e.expandSimple(bw, BracedRef("NAME"))
		require.NoError(t, err)
		_ = bw.Flush()
		assert.Equal(t, "${NAME}", buf.String())
	})

	t.Run("Empty and NoReplaceEmpty keeps literal", func(t *testing.T) {
		t.Parallel()
		e := &Engine{
			Format: formatter.NewFormatter(false),
			Lookup: func(string) (string, bool) { return "", false },
		}
		var buf bytes.Buffer
		bw := bufio.NewWriter(&buf)

		err := e.expandSimple(bw, BracedRef("NAME"))
		require.NoError(t, err)
		_ = bw.Flush()
		assert.Equal(t, "", buf.String())
	})

	t.Run("Fails when empty and FailOnEmpty", func(t *testing.T) {
		t.Parallel()
		e := &Engine{
			Format: formatter.NewFormatter(false),
			Opts:   flag.Options{KeepEmpty: true},
			Lookup: func(string) (string, bool) { return "", true }, // set but empty
		}
		var buf bytes.Buffer
		bw := bufio.NewWriter(&buf)

		err := e.expandSimple(bw, BracedRef("NAME"))
		require.NoError(t, err)
		_ = bw.Flush()
		assert.Equal(t, "${NAME}", buf.String())
	})

	t.Run("Fails when empty and FailOnEmpty", func(t *testing.T) {
		t.Parallel()
		e := &Engine{
			Format: formatter.NewFormatter(false),
			Opts:   flag.Options{ErrorEmpty: true},
			Lookup: func(string) (string, bool) { return "", true }, // set but empty
		}
		var buf bytes.Buffer
		bw := bufio.NewWriter(&buf)

		err := e.expandSimple(bw, BracedRef("NAME"))
		require.Error(t, err)
		assert.EqualError(t, err, "substitution empty: ${NAME}")
		_ = bw.Flush()
		assert.Equal(t, "", buf.String())
	})

	t.Run("Sets and writes ok value", func(t *testing.T) {
		t.Parallel()
		e := &Engine{
			Format: formatter.NewFormatter(false),
			Lookup: func(string) (string, bool) { return "ok", true },
		}

		var buf bytes.Buffer
		bw := bufio.NewWriter(&buf)

		err := e.expandSimple(bw, BracedRef("NAME"))
		require.NoError(t, err)
		_ = bw.Flush()
		assert.Equal(t, "ok", buf.String())
	})
}

func TestFastWord(t *testing.T) {
	t.Parallel()
	t.Run("Returns empty when empty", func(t *testing.T) {
		t.Parallel()
		e := &Engine{Format: formatter.NewFormatter(false)}
		out, err := e.fastWord(nil)
		require.NoError(t, err)
		assert.Equal(t, "", out)
	})

	t.Run("Returns raw string when no dollars", func(t *testing.T) {
		t.Parallel()
		e := &Engine{Format: formatter.NewFormatter(false)}
		out, err := e.fastWord([]byte("abc=def"))
		require.NoError(t, err)
		assert.Equal(t, "abc=def", out)
	})

	t.Run("Expands using child engine", func(t *testing.T) {
		t.Parallel()
		e := &Engine{
			Format: formatter.NewFormatter(false),
			Opts:   flag.Options{},
			Lookup: func(name string) (string, bool) {
				if name == "X" {
					return "Y", true
				}
				return "", false
			},
		}
		out, err := e.fastWord([]byte("hi $X and ${X}!"))
		require.NoError(t, err)
		assert.Equal(t, "hi Y and Y!", out)
	})
}

func TestExpandBytes(t *testing.T) {
	t.Parallel()

	t.Run("Text only passthrough", func(t *testing.T) {
		t.Parallel()

		e := &Engine{
			Opts:   flag.Options{},                // defaults
			Format: formatter.NewFormatter(false), // plain
			Lookup: func(string) (string, bool) { return "", false },
		}

		out, err := e.expandBytes([]byte("just some text, no vars"))
		require.NoError(t, err)
		assert.Equal(t, "just some text, no vars", out)
	})

	t.Run("Nested expansion", func(t *testing.T) {
		lookupEnv := func(name string) (string, bool) {
			switch name {
			case "X":
				return "foo", true
			case "Y":
				return "bar", true
			default:
				return "", false
			}
		}

		e := &Engine{
			Opts:   flag.Options{},                // defaults
			Format: formatter.NewFormatter(false), // plain
			Lookup: lookupEnv,
		}

		out, err := e.expandBytes([]byte("hi $X and ${Y}!"))
		require.NoError(t, err)
		assert.Equal(t, "hi foo and bar!", out)
	})

	t.Run("Escaped dollar and no expansion", func(t *testing.T) {
		t.Parallel()

		lookupEnv := func(name string) (string, bool) {
			if name == "Y" {
				return "yes", true
			}
			return "", false
		}

		e := &Engine{
			// Escapes enabled (default: NoEscape=false) → "\$" remains literal '$'
			Opts:   flag.Options{NoEscape: false},
			Format: formatter.NewFormatter(false),
			Lookup: lookupEnv,
		}

		out, err := e.expandBytes([]byte(`prefix \$X and ${Y} suffix`))
		require.NoError(t, err)
		assert.Equal(t, `prefix $X and yes suffix`, out)
	})

	t.Run("Error when unset", func(t *testing.T) {
		t.Parallel()

		lookupEnv := func(name string) (string, bool) { return "", false }
		e := &Engine{
			Opts:   flag.Options{},                // defaults
			Format: formatter.NewFormatter(false), // plain (so error text is uncolored)
			Lookup: lookupEnv,                     // VAR is unset
		}

		out, err := e.expandBytes([]byte("${VAR?boom}"))
		require.Error(t, err)
		assert.EqualError(t, err, "VAR: boom")
		assert.Equal(t, "", out)
	})
}
