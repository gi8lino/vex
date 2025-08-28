package fsm

import (
	"strings"

	"github.com/gi8lino/vex/internal/xerr"
)

// opTrimPrefix implements ${VAR#pat} and ${VAR##pat}.
func (e *Engine) opTrimPrefix(name, op string, isSet bool, val, pat string) (string, error) {
	if !isSet {
		if e.Opts.ErrorUnset {
			return "", xerr.Unset(e.Format.ErrorStr(name))
		}
		if e.Opts.KeepUnset {
			return e.Format.UnsetStr("${" + name + op + pat + "}"), nil
		}
		val = ""
	}
	if pat == "" {
		return e.Format.ErrorStr("${" + name + op + pat + "}"), nil
	}
	var out string
	if op == "#" {
		out = strings.TrimPrefix(val, pat)
	} else {
		out = trimPrefixAll(val, pat)
	}
	if e.Opts.ErrorEmpty && out == "" {
		return "", xerr.Empty(e.Format.EmptyStr(name))
	}
	return e.Format.OkStr(out), nil
}

// opTrimSuffix implements ${VAR%pat} and ${VAR%%pat}.
func (e *Engine) opTrimSuffix(name, op string, isSet bool, val, pat string) (string, error) {
	if !isSet {
		if e.Opts.ErrorUnset {
			return "", xerr.Unset(e.Format.UnsetStr(name))
		}
		if e.Opts.KeepUnset {
			return e.Format.UnsetStr("${" + name + op + pat + "}"), nil
		}
		val = ""
	}
	if pat == "" {
		return e.Format.ErrorStr("${" + name + op + pat + "}"), nil
	}
	var out string
	if op == "%" {
		out = strings.TrimSuffix(val, pat)
	} else {
		out = trimSuffixAll(val, pat)
	}
	if e.Opts.ErrorEmpty && out == "" {
		return "", xerr.Empty(e.Format.EmptyStr(name))
	}
	return e.Format.OkStr(out), nil
}
