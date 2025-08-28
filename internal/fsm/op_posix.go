package fsm

import (
	"errors"
)

// opDefault implements ${VAR-word}.
func (e *Engine) opDefault(isSet bool, val, word string) (string, error) {
	if !isSet {
		return e.Format.DefaultStr(word), nil
	}
	return e.Format.OkStr(val), nil
}

// opDefaultNull implements ${VAR:-word}.
func (e *Engine) opDefaultNull(notNull bool, val, word string) (string, error) {
	if !notNull {
		return e.Format.DefaultStr(word), nil
	}
	return e.Format.OkStr(val), nil
}

// opAssign implements ${VAR=word}.
func (e *Engine) opAssign(name string, isSet bool, val, word string) (string, error) {
	if !isSet {
		_ = e.Setenv(name, word)
		return e.Format.DefaultStr(word), nil
	}
	return e.Format.OkStr(val), nil
}

// opAssignNull implements ${VAR:=word}.
func (e *Engine) opAssignNull(name string, notNull bool, val, word string) (string, error) {
	if !notNull {
		_ = e.Setenv(name, word)
		return e.Format.DefaultStr(word), nil
	}
	return e.Format.OkStr(val), nil
}

// opAlt implements ${VAR+word}.
func (e *Engine) opAlt(name string, isSet bool, word string) (string, error) {
	if isSet {
		return e.Format.OkStr(name + ": " + word), nil
	}
	return "", nil
}

// opAltNull implements ${VAR:+word}.
func (e *Engine) opAltNull(name string, notNull bool, word string) (string, error) {
	if notNull {
		return e.Format.OkStr(name + ": " + word), nil
	}
	return "", nil
}

// opErrorUnset implements ${VAR?word}.
func (e *Engine) opErrorUnset(name string, isSet bool, word string) (string, error) {
	if !isSet {
		return "", errors.New(e.Format.UserErrorStr(name + ": " + word))
	}
	return e.Format.OkStr(""), nil
}

// opErrorNull implements ${VAR:?word}.
func (e *Engine) opErrorNull(name string, notNull bool, word string) (string, error) {
	if !notNull {
		return "", errors.New(e.Format.UserErrorStr(name + ": " + word))
	}
	return e.Format.OkStr(""), nil
}
