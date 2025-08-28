package fsm

import (
	"strings"

	"github.com/gi8lino/vex/internal/xerr"
)

// opReplace handles ${VAR/pat/repl} and ${VAR//pat/repl}.
func (e *Engine) opReplace(name, op string, isSet bool, val, spec string) (string, error) {
	if !isSet {
		if e.Opts.ErrorUnset {
			return "", xerr.Unset(e.Format.UnsetStr(name))
		}
		if e.Opts.KeepUnset {
			return e.Format.UnsetStr("${" + name + op + spec + "}"), nil
		}
		return e.Format.OkStr(""), nil // unset→empty; replace on empty stays empty
	}

	pat, repl, found := strings.Cut(spec, "/")
	if !found || pat == "" {
		return e.Format.ErrorStr("${" + name + op + spec + "}"), nil
	}

	var out string
	if op == "/" {
		out = strings.Replace(val, pat, repl, 1) // first only
	} else {
		out = strings.ReplaceAll(val, pat, repl) // all
	}

	if e.Opts.ErrorEmpty && out == "" {
		return "", xerr.Empty(e.Format.EmptyStr(name))
	}
	return e.Format.OkStr(out), nil
}
