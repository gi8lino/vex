package fsm

import (
	"bufio"
	"errors"
	"io"
)

// TokType enumerates the different token kinds produced by the tokenizer.
type TokType int

const (
	TOK_EOF        TokType = iota // end of input
	TOK_TEXT                      // literal text (non-special)
	TOK_DOLLAR                    // '$'
	TOK_LBRACE                    // '{'
	TOK_RBRACE                    // '}'
	TOK_COLON                     // ':'
	TOK_OP                        // one of - = + ?
	TOK_NAME                      // variable name: [A-Za-z_][A-Za-z0-9_]* or digits
	TOK_ESC_DOLLAR                // "\$" escape sequence (if escapes enabled)
	TOK_CARET                     // '^' (case ops)
	TOK_COMMA                     // ',' (case ops)
	TOK_SLASH                     // '/' (replace ops)
	TOK_HASH                      // '#' (length/trim ops)
	TOK_PERCENT                   // '%' (trim ops)
	TOK_AT                        // '@' (quoting ops)
)

// Token represents a lexical token with its type and literal bytes.
type Token struct {
	Type TokType // token type
	Lit  []byte  // literal bytes (may be empty for structural tokens)
}

// Tokenizer reads from a buffered stream and emits Tokens.
type Tokenizer struct {
	br       *bufio.Reader // input reader
	noEscape bool          // whether to disable \$ escape
}

// NewTokenizerWithSize constructs a tokenizer with a specific buffer size.
// size is clamped to at least 64 bytes.
func NewTokenizerWithSize(r io.Reader, noEscape bool, size int) *Tokenizer {
	if size < 64 {
		size = 64
	}
	return &Tokenizer{
		br:       bufio.NewReaderSize(r, size),
		noEscape: noEscape,
	}
}

// specialTable builds the sentinel table of bytes considered "special".
func specialTable() (t [256]bool) {
	for _, c := range [...]byte{
		'$', '{', '}', ':', '-', '+', '=', '?', '\\',
		'^', ',', '/', '#', '%', '@',
	} {
		t[c] = true
	}
	return
}

// isSpecial marks characters that delimit tokens in TEXT runs.
var isSpecial = specialTable()

// Next returns the next token in the stream or TOK_EOF at end of input.
func (t *Tokenizer) Next() (Token, error) {
	b, err := t.br.ReadByte()
	switch {
	case err == nil:
		// Note: this is a hot path; avoid allocating a string for single-char tokens.
	case errors.Is(err, io.EOF):
		return Token{Type: TOK_EOF}, nil
	default:
		return Token{}, err
	}

	// "\$" escape (when enabled) → produce TOK_ESC_DOLLAR
	if b == '\\' && !t.noEscape {
		n, err := t.br.ReadByte()
		switch {
		case err == nil:
			if n == '$' {
				return Token{Type: TOK_ESC_DOLLAR, Lit: []byte("$")}, nil
			}
			_ = t.br.UnreadByte()
		case errors.Is(err, io.EOF):
			return Token{Type: TOK_TEXT, Lit: []byte{'\\'}}, nil
		default:
			return Token{}, err
		}
		// fallthrough with b='\' → treated as TEXT start below
	}

	// Single-character structural tokens.
	switch b {
	case '$':
		return Token{Type: TOK_DOLLAR, Lit: []byte{'$'}}, nil
	case '{':
		return Token{Type: TOK_LBRACE, Lit: []byte{'{'}}, nil
	case '}':
		return Token{Type: TOK_RBRACE, Lit: []byte{'}'}}, nil
	case ':':
		return Token{Type: TOK_COLON, Lit: []byte{':'}}, nil
	case '-', '+', '=', '?':
		return Token{Type: TOK_OP, Lit: []byte{b}}, nil
	case '^':
		return Token{Type: TOK_CARET, Lit: []byte{'^'}}, nil
	case ',':
		return Token{Type: TOK_COMMA, Lit: []byte{','}}, nil
	case '/':
		return Token{Type: TOK_SLASH, Lit: []byte{'/'}}, nil
	case '#':
		return Token{Type: TOK_HASH, Lit: []byte{'#'}}, nil
	case '%':
		return Token{Type: TOK_PERCENT, Lit: []byte{'%'}}, nil
	case '@':
		return Token{Type: TOK_AT, Lit: []byte{'@'}}, nil
	}

	// NAME token (variable identifiers or digits).
	if isNameStart(b) || isDigit(b) {
		name := []byte{b}
		for {
			nx, err := t.br.ReadByte()
			if err != nil {
				if errors.Is(err, io.EOF) {
					return Token{Type: TOK_NAME, Lit: name}, nil
				}
				return Token{}, err
			}
			if isNameCont(nx) {
				name = append(name, nx)
				continue
			}
			_ = t.br.UnreadByte()
			return Token{Type: TOK_NAME, Lit: name}, nil
		}
	}

	// TEXT run: accumulate until encountering a special or EOF.
	text := []byte{b}
	for {
		nx, err := t.br.ReadByte()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return Token{Type: TOK_TEXT, Lit: text}, nil
			}
			return Token{}, err
		}
		if isSpecial[nx] || isNameStart(nx) || isDigit(nx) {
			_ = t.br.UnreadByte()
			return Token{Type: TOK_TEXT, Lit: text}, nil
		}
		text = append(text, nx)
	}
}

// EmitUntilDollar streams bytes to w until an *unescaped* '$' or EOF.
func (t *Tokenizer) EmitUntilDollar(w *bufio.Writer) (Token, error) {
	for {
		chunk, err := t.br.ReadSlice('$') // includes '$' if found
		switch {
		case err == nil:
			// Found '$' at end; decide if it's escaped when escapes are enabled.
			if !t.noEscape && len(chunk) >= 2 {
				// Count trailing backslashes before the '$'.
				i := len(chunk) - 2
				bs := 0
				for i >= 0 && chunk[i] == '\\' {
					bs++
					i--
				}
				if bs%2 == 1 {
					// Escaped: write prefix up to the backslash, then literal '$', continue.
					// Example: "abc\$" → write "abc", then '$'.
					if _, werr := w.Write(chunk[:len(chunk)-2]); werr != nil {
						return Token{}, werr
					}
					if err := w.WriteByte('$'); err != nil {
						return Token{}, err
					}
					continue
				}
			}
			// Unescaped '$': write preceding bytes (not the '$') and return.
			if len(chunk) > 1 {
				if _, werr := w.Write(chunk[:len(chunk)-1]); werr != nil {
					return Token{}, werr
				}
			}
			return Token{Type: TOK_DOLLAR, Lit: []byte{'$'}}, nil

		case errors.Is(err, bufio.ErrBufferFull):
			// No '$' yet; stream the buffer and keep going.
			if len(chunk) > 0 {
				if _, werr := w.Write(chunk); werr != nil {
					return Token{}, werr
				}
			}
			continue

		case errors.Is(err, io.EOF):
			// EOF: flush remainder and signal EOF.
			if len(chunk) > 0 {
				if _, werr := w.Write(chunk); werr != nil {
					return Token{}, werr
				}
			}
			return Token{Type: TOK_EOF}, nil

		default:
			return Token{}, err
		}
	}
}

// isNameStart reports whether a byte can start a variable name.
func isNameStart(b byte) bool { return (b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z') || b == '_' }

// isNameCont reports whether a byte can continue a variable name.
func isNameCont(b byte) bool { return isNameStart(b) || isDigit(b) }

// isDigit reports whether a byte is a decimal digit.
func isDigit(b byte) bool { return b >= '0' && b <= '9' }
