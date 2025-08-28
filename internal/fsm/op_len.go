package fsm

import (
	"strconv"

	"github.com/gi8lino/vex/internal/xerr"
)

// opLen implements ${#VAR}.
func (e *Engine) opLen(name string, isSet bool, val string) (string, error) {
	if !isSet {
		if e.Opts.ErrorUnset {
			return "", xerr.Unset(e.Format.UnsetStr(name))
		}
		return e.Format.OkStr("0"), nil
	}
	n := len([]rune(val))
	return e.Format.OkStr(strconv.Itoa(n)), nil
}
