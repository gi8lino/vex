package formatter

// plainFormatter implements Formatter without adding any styling.
type plainFormatter struct{}

// OkStr returns s unchanged for successful substitutions.
func (plainFormatter) OkStr(s string) string { return s }

// DefaultStr returns s unchanged for defaulted substitutions.
func (plainFormatter) DefaultStr(s string) string { return s }

// UserErrorStr returns s unchanged for user-supplied error messages.
func (plainFormatter) UserErrorStr(s string) string { return s }

// EmptyStr returns s unchanged for empty values.
func (plainFormatter) EmptyStr(s string) string { return s }

// UnsetStr returns s unchanged for unset values.
func (plainFormatter) UnsetStr(s string) string { return s }

// ErrorStr returns s unchanged for substitution errors.
func (plainFormatter) ErrorStr(s string) string { return s }

// FilterStr returns s unchanged for filtered variables.
func (plainFormatter) FilterStr(s string) string { return s }
