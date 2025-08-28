package fsm

import (
	"testing"

	"github.com/gi8lino/vex/internal/flag"
	"github.com/gi8lino/vex/internal/formatter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpQuote(t *testing.T) {
	t.Parallel()
	const label = "EngineLabel"

	t.Run("error when unset and FailOnUnset", func(t *testing.T) {
		t.Parallel()

		e := &Engine{
			Format: formatter.NewFormatter(false),
			Label:  label,
			Opts: flag.Options{
				ErrorUnset: true,
				KeepUnset:  false,
				ErrorEmpty: false,
			},
		}
		out, err := e.opQuote("VAR", false /*isSet*/, "ignored", "Q")
		require.Error(t, err)
		assert.EqualError(t, err, "variable not set: VAR")
		assert.Empty(t, out)
	})

	t.Run("returns missing marker when unset and NoReplaceUnset Q", func(t *testing.T) {
		t.Parallel()

		e := &Engine{
			Format: formatter.NewFormatter(false),
			Label:  label,
			Opts: flag.Options{
				ErrorUnset: false,
				KeepUnset:  true,
				ErrorEmpty: false,
			},
		}
		out, err := e.opQuote("VAR", false /*isSet*/, "ignored", "Q")
		require.NoError(t, err)
		assert.Equal(t, "${VAR@Q}", out)
	})

	t.Run("returns missing marker when unset and NoReplaceUnset preserves modeRaw", func(t *testing.T) {
		t.Parallel()

		e := &Engine{
			Format: formatter.NewFormatter(false),
			Label:  label,
			Opts: flag.Options{
				ErrorUnset: false,
				KeepUnset:  true,
				ErrorEmpty: false,
			},
		}
		// ensure original casing/spacing is preserved in literal
		out, err := e.opQuote("VAR", false /*isSet*/, "ignored", "  j ")
		require.NoError(t, err)
		assert.Equal(t, "${VAR@  j }", out)
	})

	t.Run("unset without NoReplaceUnset uses empty value and Q quotes empty", func(t *testing.T) {
		t.Parallel()

		e := &Engine{
			Format: formatter.NewFormatter(false),
			Label:  label,
			Opts: flag.Options{
				ErrorUnset: false,
				KeepUnset:  false,
				ErrorEmpty: false,
			},
		}
		out, err := e.opQuote("VAR", false /*isSet*/, "ignored", "Q")
		require.NoError(t, err)
		assert.Equal(t, "''", out) // shell-quoted empty
	})

	t.Run("mode Q shell quotes when set", func(t *testing.T) {
		t.Parallel()

		e := &Engine{
			Label:  label,
			Format: formatter.NewFormatter(false),
		}
		out, err := e.opQuote("VAR", true /*isSet*/, "a'b c", "Q")
		require.NoError(t, err)
		assert.Equal(t, `'a'"'"'b c'`, out)
	})

	t.Run("mode Q is case insensitive and ignores spaces", func(t *testing.T) {
		t.Parallel()

		e := &Engine{
			Label:  label,
			Format: formatter.NewFormatter(false),
		}
		out, err := e.opQuote("VAR", true /*isSet*/, "x", "  q ")
		require.NoError(t, err)
		assert.Equal(t, `'x'`, out)
	})

	t.Run("mode J json quotes when set with escapes", func(t *testing.T) {
		t.Parallel()

		e := &Engine{
			Label:  label,
			Format: formatter.NewFormatter(false),
		}
		out, err := e.opQuote("VAR", true /*isSet*/, "a\"b\\c\n", "J")
		require.NoError(t, err)
		assert.Equal(t, `"a\"b\\c\n"`, out)
	})

	t.Run("mode Y yaml quotes when set", func(t *testing.T) {
		t.Parallel()

		e := &Engine{
			Label:  label,
			Format: formatter.NewFormatter(false),
		}
		out, err := e.opQuote("VAR", true /*isSet*/, "o'hai", "Y")
		require.NoError(t, err)
		assert.Equal(t, `'o''hai'`, out)
	})

	t.Run("unknown mode keeps literal", func(t *testing.T) {
		t.Parallel()

		e := &Engine{
			Label:  label,
			Format: formatter.NewFormatter(false),
		}
		out, err := e.opQuote("VAR", true /*isSet*/, "value", "Zz")
		require.NoError(t, err)
		// literal must preserve original modeRaw (no trim/upper)
		assert.Equal(t, "${VAR@Zz}", out)
	})
}

func TestShellQuote(t *testing.T) {
	t.Parallel()

	t.Run("empty string", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "''", shellQuote(""))
	})

	t.Run("no specials", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "'hello world'", shellQuote("hello world"))
	})

	t.Run("single quote inside", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, `'a'"'"'b'`, shellQuote("a'b"))
	})

	t.Run("multiple single quotes", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, `'a'"'"'b'"'"'c'`, shellQuote("a'b'c"))
	})
}

func TestJsonQuote(t *testing.T) {
	t.Parallel()

	t.Run("empty string", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, `""`, jsonQuote(""))
	})

	t.Run("basic chars", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, `"abc"`, jsonQuote("abc"))
	})

	t.Run("escapes quotes and backslashes", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, `"a\"b\\c"`, jsonQuote(`a"b\c`))
	})

	t.Run("escapes control chars", func(t *testing.T) {
		t.Parallel()
		// includes \b, \f, \n, \r, \t
		assert.Equal(t, `"a\b\f\n\r\t"`, jsonQuote("a\b\f\n\r\t"))
	})

	t.Run("escapes other control as unicode", func(t *testing.T) {
		t.Parallel()
		// 0x01 and 0x1F become \u0001 and \u001f
		in := string([]byte{0x01, 'X', 0x1F})
		assert.Equal(t, `"\u0001X\u001f"`, jsonQuote(in))
	})
}

func TestYamlQuote(t *testing.T) {
	t.Parallel()

	t.Run("empty string", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "''", yamlQuote(""))
	})

	t.Run("no specials", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "'hello'", yamlQuote("hello"))
	})

	t.Run("duplicates single quotes", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, `'o''hai'`, yamlQuote("o'hai"))
	})

	t.Run("multiple single quotes", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, `'a''b''c'`, yamlQuote("a'b'c"))
	})
}
