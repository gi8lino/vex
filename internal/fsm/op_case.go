package fsm

import "github.com/gi8lino/vex/internal/xerr"

// opCase handles ${VAR^}, ${VAR^^}, ${VAR,}, ${VAR,,}.
func (e *Engine) opCase(name, op string, isSet bool, val string) (string, error) {
	if !isSet {
		if e.Opts.ErrorUnset {
			return "", xerr.Unset(e.Format.UnsetStr(name))
		}
		if e.Opts.KeepUnset {
			return e.Format.UnsetStr("${" + name + op + "}"), nil
		}
		val = ""
	}
	out := transformCase(op, val)
	if e.Opts.ErrorEmpty && out == "" {
		return "", xerr.Empty(e.Format.EmptyStr(name))
	}
	return e.Format.OkStr(out), nil
}
