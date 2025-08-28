package fsm

import (
	"bufio"
	"bytes"
	"sync"

	"github.com/gi8lino/vex/internal/xerr"
)

// Buffer pool to avoid per-call allocations.
var bufPool = sync.Pool{New: func() any { return new(bytes.Buffer) }}

// Pool of small tokenizers for nested operator words.
var smallTokPool = sync.Pool{
	New: func() any {
		// Tiny reader; we'll Reset() per use.
		return NewTokenizerWithSize(bytes.NewReader(nil), false, 512)
	},
}

// expandSimple resolves $VAR and ${VAR} without operator semantics.
func (e *Engine) expandSimple(w *bufio.Writer, v VarRef) error {
	if v.Name == "" {
		_, err := w.WriteString(v.Lit())
		return err
	}
	if !e.filter(v.Name) {
		_, err := w.WriteString(e.Format.FilterStr(v.Lit()))
		return err
	}
	val, ok := e.Lookup(v.Name)
	if !ok {
		if e.Opts.ErrorUnset {
			return xerr.Unset(e.Format.UnsetStr(v.Lit()))
		}
		if e.Opts.KeepUnset {
			_, err := w.WriteString(e.Format.UnsetStr(v.Lit()))
			return err
		}
		return nil // unset allowed → nothing written
	}
	if (e.Opts.KeepVars || e.Opts.KeepEmpty) && val == "" {
		_, err := w.WriteString(e.Format.EmptyStr(v.Lit()))
		return err
	}
	if e.Opts.ErrorEmpty && val == "" {
		return xerr.Empty(e.Format.EmptyStr(v.Lit()))
	}
	_, err := w.WriteString(e.Format.OkStr(val))
	return err
}

// fastWord returns the operator word with nested expansion only if needed.
// If raw contains no '$', nested expansion is unnecessary.
func (e *Engine) fastWord(raw []byte) (string, error) {
	if len(raw) == 0 {
		return "", nil
	}
	if bytes.IndexByte(raw, '$') < 0 {
		return string(raw), nil
	}
	return e.expandBytes(raw)
}

// expandWithOp dispatches ${VAR<op>word} to operator helpers.
func (e *Engine) expandWithOp(name, op string, raw []byte) (string, error) {
	if e.Opts.NoOps || op == "" {
		// treat as simple braced
		var buf bytes.Buffer
		bw := bufio.NewWriter(&buf)
		if err := e.expandSimple(bw, BracedRef(name)); err != nil {
			return "", err
		}
		_ = bw.Flush()
		return buf.String(), nil
	}

	// Fast-path the operator word.
	word := ""
	if raw != nil {
		w, err := e.fastWord(raw)
		if err != nil {
			return "", err
		}
		word = w
	}

	val, isSet := e.Lookup(name)
	notNull := isSet && val != ""

	switch op {
	case "#len":
		return e.opLen(name, isSet, val)
	case "#", "##":
		return e.opTrimPrefix(name, op, isSet, val, word)
	case "%", "%%":
		return e.opTrimSuffix(name, op, isSet, val, word)
	case "^", "^^", ",", ",,":
		return e.opCase(name, op, isSet, val)
	case ":":
		return e.opSubstr(name, isSet, val, word)
	case "/", "//":
		return e.opReplace(name, op, isSet, val, word)
	case "@":
		return e.opQuote(name, isSet, val, word)
	case "-":
		return e.opDefault(isSet, val, word)
	case ":-":
		return e.opDefaultNull(notNull, val, word)
	case "=":
		return e.opAssign(name, isSet, val, word)
	case ":=":
		return e.opAssignNull(name, notNull, val, word)
	case "+":
		return e.opAlt(name, isSet, word)
	case ":+":
		return e.opAltNull(name, notNull, word)
	case "?":
		return e.opErrorUnset(name, isSet, word)
	case ":?":
		return e.opErrorNull(name, notNull, word)
	default:
		return "${" + name + op + string(raw) + "}", nil
	}
}

// expandBytes runs nested expansion with a small, pooled tokenizer.
func (e *Engine) expandBytes(raw []byte) (string, error) {
	b := bufPool.Get().(*bytes.Buffer)
	b.Reset()
	b.Grow(len(raw))
	bw := bufio.NewWriter(b)

	// Reuse the same engine; no need to construct a child.
	tok := smallTokPool.Get().(*Tokenizer)
	tok.noEscape = e.Opts.NoEscape
	tok.br.Reset(bytes.NewReader(raw))

	if err := e.consumeWithTokenizer(tok, bw); err != nil {
		smallTokPool.Put(tok)
		bufPool.Put(b)
		return "", err
	}
	smallTokPool.Put(tok)

	_ = bw.Flush()
	out := b.String()
	b.Reset()
	bufPool.Put(b)
	return out, nil
}
