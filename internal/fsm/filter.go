package fsm

import (
	"slices"
	"strings"
)

// filter checks name against --variable/--prefix/--suffix allowlists.
// When all lists are empty, everything is allowed.
func (e *Engine) filter(name string) bool {
	if len(e.Opts.Variables) == 0 && len(e.Opts.Prefix) == 0 && len(e.Opts.Suffix) == 0 {
		return true
	}
	if slices.Contains(e.Opts.Variables, name) {
		return true
	}
	for _, p := range e.Opts.Prefix {
		if strings.HasPrefix(name, p) {
			return true
		}
	}
	for _, s := range e.Opts.Suffix {
		if strings.HasSuffix(name, s) {
			return true
		}
	}
	return false
}
