package fsm

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// lit is a small helper to compare token literals as strings
func lit(t *testing.T, tok Token) string { t.Helper(); return string(tok.Lit) }

func TestNewTokenizerWithSize(t *testing.T) {
	t.Parallel()

	t.Run("empty input returns EOF", func(t *testing.T) {
		t.Parallel()
		tok := NewTokenizerWithSize(strings.NewReader(""), false, 1<<20)
		t1, err := tok.Next()
		require.NoError(t, err)
		assert.Equal(t, TOK_EOF, t1.Type)
	})

	t.Run("escape dollar when allowed returns ESC DOLLAR", func(t *testing.T) {
		t.Parallel()
		tok := NewTokenizerWithSize(strings.NewReader(`\$`), false, 1<<20) // escapes enabled
		t1, err := tok.Next()
		require.NoError(t, err)
		assert.Equal(t, TOK_ESC_DOLLAR, t1.Type)
		assert.Equal(t, "$", lit(t, t1))

		t2, err := tok.Next()
		require.NoError(t, err)
		assert.Equal(t, TOK_EOF, t2.Type)
	})

	t.Run("escape sequence ignored when noEscape true", func(t *testing.T) {
		t.Parallel()
		tok := NewTokenizerWithSize(strings.NewReader(`\$`), true, 1<<20) // escapes disabled
		// First token is TEXT "\" (backslash literal)
		t1, err := tok.Next()
		require.NoError(t, err)
		assert.Equal(t, TOK_TEXT, t1.Type)
		assert.Equal(t, `\`, lit(t, t1))

		// Then a DOLLAR token
		t2, err := tok.Next()
		require.NoError(t, err)
		assert.Equal(t, TOK_DOLLAR, t2.Type)
		assert.Equal(t, "$", lit(t, t2))

		t3, err := tok.Next()
		require.NoError(t, err)
		assert.Equal(t, TOK_EOF, t3.Type)
	})

	t.Run("backslash not followed by dollar with escapes enabled", func(t *testing.T) {
		t.Parallel()
		tok := NewTokenizerWithSize(strings.NewReader(`\a`), false, 1<<20)
		// TEXT with "\" only
		t1, err := tok.Next()
		require.NoError(t, err)
		assert.Equal(t, TOK_TEXT, t1.Type)
		assert.Equal(t, `\`, lit(t, t1))
		// NAME "a"
		t2, err := tok.Next()
		require.NoError(t, err)
		assert.Equal(t, TOK_NAME, t2.Type)
		assert.Equal(t, "a", lit(t, t2))
		// EOF
		t3, err := tok.Next()
		require.NoError(t, err)
		assert.Equal(t, TOK_EOF, t3.Type)
	})

	t.Run("lone backslash at eof becomes TEXT backslash", func(t *testing.T) {
		t.Parallel()
		tok := NewTokenizerWithSize(strings.NewReader(`\`), false, 1<<20)
		t1, err := tok.Next()
		require.NoError(t, err)
		assert.Equal(t, TOK_TEXT, t1.Type)
		assert.Equal(t, `\`, lit(t, t1))
		t2, err := tok.Next()
		require.NoError(t, err)
		assert.Equal(t, TOK_EOF, t2.Type)
	})

	t.Run("single char tokens all", func(t *testing.T) {
		t.Parallel()
		// Sequence covers: $, {, }, :, -, +, =, ?, ^, ,, /, #, %, @
		input := "${}:-+=?^,/#%@"
		tok := NewTokenizerWithSize(strings.NewReader(input), false, 1<<20)

		expect := []struct {
			typ TokType
			lit string
		}{
			{TOK_DOLLAR, "$"},
			{TOK_LBRACE, "{"},
			{TOK_RBRACE, "}"},
			{TOK_COLON, ":"},
			{TOK_OP, "-"},
			{TOK_OP, "+"},
			{TOK_OP, "="},
			{TOK_OP, "?"},
			{TOK_CARET, "^"},
			{TOK_COMMA, ","},
			{TOK_SLASH, "/"},
			{TOK_HASH, "#"},
			{TOK_PERCENT, "%"},
			{TOK_AT, "@"},
		}

		for i := range expect {
			tokn, err := tok.Next()
			require.NoError(t, err)
			assert.Equalf(t, expect[i].typ, tokn.Type, "index %d", i)
			assert.Equalf(t, expect[i].lit, lit(t, tokn), "index %d", i)
		}

		tEOF, err := tok.Next()
		require.NoError(t, err)
		assert.Equal(t, TOK_EOF, tEOF.Type)
	})

	t.Run("NAME token letters digits underscore", func(t *testing.T) {
		t.Parallel()
		tok := NewTokenizerWithSize(strings.NewReader("FOO1_bar9"), false, 1<<20)
		t1, err := tok.Next()
		require.NoError(t, err)
		assert.Equal(t, TOK_NAME, t1.Type)
		assert.Equal(t, "FOO1_bar9", lit(t, t1))
		t2, err := tok.Next()
		require.NoError(t, err)
		assert.Equal(t, TOK_EOF, t2.Type)
	})

	t.Run("NAME token starting with digit collects letters", func(t *testing.T) {
		t.Parallel()
		tok := NewTokenizerWithSize(strings.NewReader("123abc"), false, 1<<20)
		t1, err := tok.Next()
		require.NoError(t, err)
		assert.Equal(t, TOK_NAME, t1.Type)
		assert.Equal(t, "123abc", lit(t, t1))
		t2, err := tok.Next()
		require.NoError(t, err)
		assert.Equal(t, TOK_EOF, t2.Type)
	})

	t.Run("TEXT run until special or name start", func(t *testing.T) {
		t.Parallel()
		// '.' then '-' (special, OP) then 'A' (NAME start)
		tok := NewTokenizerWithSize(strings.NewReader(".-A"), false, 1<<20)

		t1, err := tok.Next()
		require.NoError(t, err)
		assert.Equal(t, TOK_TEXT, t1.Type)
		assert.Equal(t, ".", lit(t, t1))

		t2, err := tok.Next()
		require.NoError(t, err)
		assert.Equal(t, TOK_OP, t2.Type)
		assert.Equal(t, "-", lit(t, t2))

		t3, err := tok.Next()
		require.NoError(t, err)
		assert.Equal(t, TOK_NAME, t3.Type)
		assert.Equal(t, "A", lit(t, t3))

		t4, err := tok.Next()
		require.NoError(t, err)
		assert.Equal(t, TOK_EOF, t4.Type)
	})
}

func TestNameAndDigitHelpers(t *testing.T) {
	t.Parallel()

	t.Run("isNameStart", func(t *testing.T) {
		t.Parallel()
		assert.True(t, isNameStart('A'))
		assert.True(t, isNameStart('z'))
		assert.True(t, isNameStart('_'))
		assert.False(t, isNameStart('1'))
		assert.False(t, isNameStart('-'))
	})

	t.Run("isNameCont", func(t *testing.T) {
		t.Parallel()
		assert.True(t, isNameCont('A'))
		assert.True(t, isNameCont('9'))
		assert.True(t, isNameCont('_'))
		assert.False(t, isNameCont('-'))
	})

	t.Run("isDigit", func(t *testing.T) {
		t.Parallel()
		assert.True(t, isDigit('0'))
		assert.True(t, isDigit('9'))
		assert.False(t, isDigit('a'))
		assert.False(t, isDigit('_'))
	})
}
