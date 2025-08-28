package xerr

import (
	"errors"
	"fmt"
)

var (
	ErrSubst = errors.New("variable not set")   // ErrSubst marks an unset/invalid substitution.
	ErrEmpty = errors.New("substitution empty") // ErrEmpty marks a substitution that resolved to empty.
)

// Unset returns an ErrSubst-wrapped error with the given message.
func Unset(msg string) error {
	return fmt.Errorf("%w: %s", ErrSubst, msg)
}

// Empty returns an ErrEmpty-wrapped error with the given message.
func Empty(msg string) error {
	return fmt.Errorf("%w: %s", ErrEmpty, msg)
}
