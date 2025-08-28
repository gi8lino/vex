package fsm

import (
	"strings"

	"github.com/gi8lino/vex/internal/xerr"
)

// opQuote handles ${VAR@Q}, ${VAR@J}, ${VAR@Y}.
func (e *Engine) opQuote(name string, isSet bool, val, modeRaw string) (string, error) {
	if !isSet {
		if e.Opts.ErrorUnset {
			return "", xerr.Unset(e.Format.UnsetStr(name))
		}
		if e.Opts.KeepUnset {
			return e.Format.UnsetStr("${" + name + "@" + modeRaw + "}"), nil
		}
		val = ""
	}
	mode := strings.TrimSpace(strings.ToUpper(modeRaw))
	switch mode {
	case "Q":
		return e.Format.OkStr(shellQuote(val)), nil
	case "J":
		return e.Format.OkStr(jsonQuote(val)), nil
	case "Y":
		return e.Format.OkStr(yamlQuote(val)), nil
	default:
		// unknown mode → keep literal
		return e.Format.ErrorStr("${" + name + "@" + modeRaw + "}"), nil
	}
}

// shellQuote returns a POSIX single-quoted string literal.
func shellQuote(s string) string {
	if s == "" {
		return "''"
	}
	// POSIX single-quote rule: close, insert '\'' sequence, reopen
	return "'" + strings.ReplaceAll(s, "'", `'"'"'`) + "'"
}

// jsonQuote returns a valid JSON string literal without pulling in encoding/json.
func jsonQuote(s string) string {
	var b strings.Builder
	b.Grow(len(s) + 2) // rough pre-alloc
	b.WriteByte('"')
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch c {
		case '"':
			b.WriteString(`\"`)
		case '\\':
			b.WriteString(`\\`)
		case '\b':
			b.WriteString(`\b`)
		case '\f':
			b.WriteString(`\f`)
		case '\n':
			b.WriteString(`\n`)
		case '\r':
			b.WriteString(`\r`)
		case '\t':
			b.WriteString(`\t`)
		default:
			if c < 0x20 {
				// control chars as \u00XX
				b.WriteString(`\u00`)
				b.WriteByte("0123456789abcdef"[c>>4])
				b.WriteByte("0123456789abcdef"[c&0xF])
			} else {
				b.WriteByte(c)
			}
		}
	}
	b.WriteByte('"')
	return b.String()
}

// yamlQuote returns a YAML single-quoted scalar (duplicates single quotes).
func yamlQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}
