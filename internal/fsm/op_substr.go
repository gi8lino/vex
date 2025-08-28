package fsm

import "github.com/gi8lino/vex/internal/xerr"

// opSubstr handles ${VAR:off[:len]}.
func (e *Engine) opSubstr(name string, isSet bool, val, spec string) (string, error) {
	if !isSet {
		if e.Opts.ErrorUnset {
			return "", xerr.Unset(e.Format.UnsetStr(name))
		}
		if e.Opts.KeepUnset {
			return e.Format.UnsetStr("${" + name + ":" + spec + "}"), nil
		}
		val = ""
	}
	out := substr(spec, val)
	if e.Opts.ErrorEmpty && out == "" {
		return "", xerr.Empty(e.Format.EmptyStr(name))
	}
	return e.Format.OkStr(out), nil
}
