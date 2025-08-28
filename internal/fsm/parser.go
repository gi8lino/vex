package fsm

import (
	"bufio"
	"bytes"
	"io"
	"sync"

	"github.com/gi8lino/vex/internal/flag"
	"github.com/gi8lino/vex/internal/formatter"
)

// Engine is the top-level expander state machine.
type Engine struct {
	Label  string                      // label used in error reporting (e.g., file name)
	Opts   flag.Options                // parsed CLI options controlling expansion
	Lookup func(string) (string, bool) // environment lookup (name → value, ok)
	Setenv func(string, string) error  // environment setter (for := and = operators)
	Format formatter.Formatter         // formatter (plain/colored)
}

// pool for op-word buffers to avoid per-expression allocations
var wordPool = sync.Pool{New: func() any { return new(bytes.Buffer) }}

// smallOp is a tiny fixed-size operator accumulator ("-",":-","##","//",",,", etc.).
type smallOp struct {
	b [2]byte
	n int
}

func (o *smallOp) reset() { o.n = 0 }
func (o *smallOp) addByte(c byte) {
	if o.n < 2 {
		o.b[o.n] = c
		o.n++
	}
}
func (o *smallOp) String() string { return string(o.b[:o.n]) }

// contextBuffers holds transient data while parsing a ${...} expression.
type contextBuffers struct {
	name  bytes.Buffer  // variable name being parsed
	op    smallOp       // operator accumulator (no heap)
	word  *bytes.Buffer // pooled buffer for operator word (may be nil until used)
	depth int           // nesting depth inside {...} while reading word
}

func (b *contextBuffers) reset() {
	b.name.Reset()
	b.op.reset()
	b.depth = 0
	if b.word != nil {
		b.word.Reset()
		// return to pool at end of expression; we do it explicitly where we finish
	}
}

// stateFn represents a single FSM state function.
type stateFn func(*runCtx) (stateFn, error)

// runCtx is the mutable context during one expansion run.
type runCtx struct {
	e   *Engine       // owning Engine instance
	w   *bufio.Writer // destination writer for expanded output
	tok *Tokenizer    // tokenizer producing tokens from input
	b   contextBuffers
}

// consumeWithTokenizer runs the FSM using a provided tokenizer.
func (e *Engine) consumeWithTokenizer(tok *Tokenizer, w *bufio.Writer) error {
	ctx := &runCtx{e: e, w: w, tok: tok}
	state := stateText
	for {
		next, err := state(ctx)
		if err != nil {
			return err
		}
		if next == nil {
			return nil
		}
		state = next
	}
}

// Consume runs the FSM on an input stream and writes expanded output.
func (e *Engine) Consume(r io.Reader, w *bufio.Writer) error {
	tok := NewTokenizerWithSize(r, e.Opts.NoEscape, 1<<20)
	return e.consumeWithTokenizer(tok, w)
}

// stateText streams text to '$' or EOF using EmitUntilDollar (zero-alloc).
func stateText(ctx *runCtx) (stateFn, error) {
	tok, err := ctx.tok.EmitUntilDollar(ctx.w)
	if err != nil {
		return nil, err
	}
	switch tok.Type {
	case TOK_DOLLAR:
		return stateAfterDollar, nil
	case TOK_EOF:
		return nil, ctx.w.Flush()
	default:
		return stateText, nil
	}
}

// stateAfterDollar decides between bare name, braced form, or literal '$'.
func stateAfterDollar(ctx *runCtx) (stateFn, error) {
	t, err := ctx.tok.Next()
	if err != nil {
		return nil, err
	}
	switch t.Type {
	case TOK_LBRACE:
		ctx.b.reset()
		return stateBracedName, nil
	case TOK_NAME:
		// Avoid building a long-lived string; but expandSimple needs a string today.
		// This path is single-token name => convert once.
		if err := ctx.e.expandSimple(ctx.w, BareRef(string(t.Lit))); err != nil {
			return nil, err
		}
		return stateText, nil
	case TOK_EOF:
		if err := ctx.w.WriteByte('$'); err != nil {
			return nil, err
		}
		return nil, ctx.w.Flush()
	default:
		// "$<non-name>" => emit '$' then that literal token (if not EOF)
		if err := ctx.w.WriteByte('$'); err != nil {
			return nil, err
		}
		if t.Type != TOK_EOF {
			if _, err := ctx.w.Write(t.Lit); err != nil {
				return nil, err
			}
		}
		return stateText, nil
	}
}

// stateBracedName consumes the variable name inside ${...}, or transitions to op/close.
func stateBracedName(ctx *runCtx) (stateFn, error) {
	t, err := ctx.tok.Next()
	if err != nil {
		return nil, err
	}
	switch t.Type {
	case TOK_HASH:
		// Special #{len} handling only valid at start with no name/op yet.
		if ctx.b.name.Len() == 0 && ctx.b.op.n == 0 {
			ctx.b.op.reset()
			// Represent "#len" with a sentinel string via the slow path (rare).
			// We’ll pass "#len" explicitly later.
			ctx.b.op.b[0] = '#'
			ctx.b.op.n = 1 // just a marker that we started with '#'
			// Keep accumulating name/op; actual dispatch happens on RBRACE.
			return stateBracedName, nil
		}
		ctx.b.op.addByte(t.Lit[0])
		return stateBracedOp, nil

	case TOK_PERCENT, TOK_CARET, TOK_COMMA, TOK_SLASH, TOK_COLON, TOK_OP, TOK_AT:
		if ctx.b.name.Len() != 0 && !ctx.e.Opts.NoOps {
			ctx.b.op.addByte(t.Lit[0])
			return stateBracedOp, nil
		}
		// Operators disabled/unexpected → keep literal (format as error).
		if _, err := ctx.w.WriteString(ctx.e.Format.ErrorStr("${" + ctx.b.name.String() + string(t.Lit))); err != nil {
			return nil, err
		}
		return stateText, nil

	case TOK_NAME:
		_, _ = ctx.b.name.Write(t.Lit)
		return stateBracedName, nil

	case TOK_RBRACE:
		// Handle the #len sentinel (name must follow later tokens)
		if ctx.b.op.n == 1 && ctx.b.op.b[0] == '#' && ctx.b.name.Len() == 0 {
			// "${#}" is not valid -> treat as literal error (consistent behavior)
			if _, err := ctx.w.WriteString(ctx.e.Format.ErrorStr("${#}")); err != nil {
				return nil, err
			}
			return stateText, nil
		}
		if ctx.b.op.n == 1 && ctx.b.op.b[0] == '#' && ctx.b.name.Len() > 0 {
			val, err := ctx.e.expandWithOp(ctx.b.name.String(), "#len", nil)
			if err != nil {
				return nil, err
			}
			if _, err := ctx.w.WriteString(val); err != nil {
				return nil, err
			}
			return stateText, nil
		}
		if err := ctx.e.expandSimple(ctx.w, BracedRef(ctx.b.name.String())); err != nil {
			return nil, err
		}
		return stateText, nil

	default:
		// Unexpected token inside braces → keep literal (format as error).
		if _, err := ctx.w.WriteString(ctx.e.Format.ErrorStr("${" + ctx.b.name.String() + string(t.Lit))); err != nil {
			return nil, err
		}
		return stateText, nil
	}
}

// stateBracedOp parses the operator after ${VAR...}, and starts collecting the word.
func stateBracedOp(ctx *runCtx) (stateFn, error) {
	t, err := ctx.tok.Next()
	if err != nil {
		return nil, err
	}
	switch t.Type {
	case TOK_HASH, TOK_PERCENT, TOK_CARET, TOK_COMMA, TOK_SLASH, TOK_COLON, TOK_OP, TOK_AT:
		// up to 2-char operators; if more arrives, it belongs to word
		if ctx.b.op.n < 2 {
			ctx.b.op.addByte(t.Lit[0])
			return stateBracedOp, nil
		}
		if ctx.b.word == nil {
			ctx.b.word = wordPool.Get().(*bytes.Buffer)
			ctx.b.word.Reset()
		}
		ctx.b.depth = 0
		ctx.b.word.Write(t.Lit)
		return stateBracedWord, nil

	case TOK_RBRACE:
		val, err := ctx.e.expandWithOp(ctx.b.name.String(), ctx.b.op.String(), nil)
		if err != nil {
			return nil, err
		}
		if _, err := ctx.w.WriteString(val); err != nil {
			return nil, err
		}
		return stateText, nil

	default:
		if ctx.b.word == nil {
			ctx.b.word = wordPool.Get().(*bytes.Buffer)
			ctx.b.word.Reset()
		}
		ctx.b.depth = 0
		ctx.b.word.Write(t.Lit)
		return stateBracedWord, nil
	}
}

// stateBracedWord collects the operator word until closing brace (supports nesting).
func stateBracedWord(ctx *runCtx) (stateFn, error) {
	t, err := ctx.tok.Next()
	if err != nil {
		return nil, err
	}
	switch t.Type {
	case TOK_LBRACE:
		ctx.b.depth++
		ctx.b.word.Write(t.Lit)
		return stateBracedWord, nil

	case TOK_RBRACE:
		if ctx.b.depth > 0 {
			ctx.b.depth--
			ctx.b.word.Write(t.Lit)
			return stateBracedWord, nil
		}
		val, err := ctx.e.expandWithOp(ctx.b.name.String(), ctx.b.op.String(), ctx.b.word.Bytes())
		// return word buffer to pool now that we’re done with it
		wordPool.Put(ctx.b.word)
		ctx.b.word = nil
		if err != nil {
			return nil, err
		}
		if _, err := ctx.w.WriteString(val); err != nil {
			return nil, err
		}
		return stateText, nil

	case TOK_EOF:
		// Unterminated ${... → emit literally (format as error), WITHOUT adding a '}'.
		lit := "${" + ctx.b.name.String() + ctx.b.op.String() + ctx.b.word.String()
		wordPool.Put(ctx.b.word)
		ctx.b.word = nil
		if _, err := ctx.w.WriteString(ctx.e.Format.ErrorStr(lit)); err != nil {
			return nil, err
		}
		return nil, ctx.w.Flush()

	default:
		ctx.b.word.Write(t.Lit)
		return stateBracedWord, nil
	}
}
