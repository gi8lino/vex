package formatter

// Formatter defines how substitution results and diagnostics are formatted.
type Formatter interface {
	OkStr(string) string        // OkStr formats a successful/valid substitution.
	DefaultStr(string) string   // DefaultStr formats a value that comes from a default/fallback.
	UserErrorStr(string) string // ErrorStr formats an user error message.
	FilterStr(string) string    // FilterStr formats a filtered variable.
	EmptyStr(string) string     // EmptyStr formats an empty variable.
	UnsetStr(string) string     // UnsetStr formats an unset variable.
	ErrorStr(string) string     // ErrorStr formats an engine/internal error.
}

// NewFormatter returns a Formatter selected based on the given flag.
// If colored is true, it returns a coloredFormatter with ANSI output.
// Otherwise, it returns a plainFormatter that performs no styling.
func NewFormatter(colored bool) Formatter {
	if colored {
		return coloredFormatter{}
	}
	return plainFormatter{}
}
