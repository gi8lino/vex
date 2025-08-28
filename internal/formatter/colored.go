package formatter

// coloredFormatter implements Formatter with ANSI escape sequences to colorize substitutions.
type coloredFormatter struct{}

const (
	green   = "\x1b[32m"       // ok
	yell    = "\x1b[33m"       // default
	orange  = "\x1b[38;5;208m" // empty
	magenta = "\x1b[35m"       // unset
	red     = "\x1b[91m"       // engine/internal error
	purple  = "\x1b[95m"       // user error message
	gray    = "\x1b[90m"       // filtered
	reset   = "\x1b[0m"
)

// OkStr wraps s in green ANSI codes for successful substitutions.
func (coloredFormatter) OkStr(s string) string { return green + s + reset }

// DefaultStr wraps s in yellow ANSI codes for defaulted substitutions.
func (coloredFormatter) DefaultStr(s string) string { return yell + s + reset }

// UserErrorStr wraps s in purple ANSI codes for user error messages.
func (coloredFormatter) UserErrorStr(s string) string { return purple + s + reset }

// EmptyStr wraps s in orange ANSI codes for empty values.
func (coloredFormatter) EmptyStr(s string) string { return orange + s + reset }

// UnsetStr wraps s in pink ANSI codes for unset values.
func (coloredFormatter) UnsetStr(s string) string { return magenta + s + reset }

// FilterStr wraps s in gray ANSI codes for filtered variables.
func (coloredFormatter) FilterStr(s string) string { return gray + s + reset }

// ErrorStr wraps s in red ANSI codes for substitution errors.
func (coloredFormatter) ErrorStr(s string) string { return red + s + reset }
