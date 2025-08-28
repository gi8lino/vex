package flag

import (
	"strings"

	tinyflags "github.com/containeroo/tinyflags"
)

// Options holds all parsed CLI flags.
type Options struct {
	// I/O mode
	InPlace   bool   // -i, --in-place
	BackupExt string // --backup

	// Parsing/behavior
	NoOps    bool // --no-ops
	NoEscape bool // --literal-dollar

	// Failure policy
	ErrorEmpty bool // --error-empty (or via --strict)
	ErrorUnset bool // --error-unset (or via --strict)
	Strict     bool // (reserved for future)

	// Replacement policy
	KeepUnset bool // --keep-unset
	KeepEmpty bool // --keep-empty
	KeepVars  bool // --keep-vars  (implies both)

	// Filter lists
	Prefix    []string // -p, --prefix
	Suffix    []string // -s, --suffix
	Variables []string // -v, --variable

	// Coloring (content + diagnostics). Incompatible with --in-place.
	Colored bool // --colored

	// Vars injection (files only, multiple allowed)
	VarsFiles []string // --vars FILE [--vars FILE...]

	// Positional file args
	Positional []string
}

// ParseFlags parses CLI flags and returns Options or a (help/version) error.
func ParseFlags(args []string, version, commit string) (Options, error) {
	var out Options

	fs := tinyflags.NewFlagSet("vex", tinyflags.ContinueOnError)
	fs.Version(version)
	fs.HelpText("show help")
	fs.VersionText("show version")

	// I/O mode
	fs.BoolVar(&out.InPlace, "in-place", false, "edit files in place; with no files, stdin->stdout").
		Short("i").
		OneOfGroup("mode").
		Value()
	fs.StringVar(&out.BackupExt, "backup", "", "when -i, create a backup with this extension (e.g. .bak)").
		Finalize(func(s string) string {
			cleaned, _ := strings.CutPrefix(s, ".")
			return "." + cleaned
		}).
		Short("b").
		Requires("in-place").
		Value()

	// Behavior
	fs.BoolVar(&out.NoOps, "no-ops", false, "treat operator forms as literals (envsubst-compatible mode)").
		Value()
	fs.BoolVar(&out.NoEscape, "literal-dollar", false, "treat \\$ as two bytes (disable dollar-escape)").
		Short("l").
		Value()

	// Failure policy
	var strict bool
	fs.BoolVar(&strict, "strict", false, "exit on unset or empty (equivalent to --error-unset --error-empty)").
		Short("x").
		Value()
	fs.BoolVar(&out.ErrorUnset, "error-unset", false, "error if a variable is unset").
		Short("u").
		Value()
	fs.BoolVar(&out.ErrorEmpty, "error-empty", false, "error if a substitution resolves to empty").
		Short("e").
		Value()

	// Replacement policy
	var keepVars bool
	fs.BoolVar(&keepVars, "keep-vars", false, "leave all ${VAR} literals (implies --keep-unset --keep-empty)").
		Short("K").
		Value()
	fs.BoolVar(&out.KeepUnset, "keep-unset", false, "leave ${VAR} literal if unset").
		Short("U").
		Value()
	fs.BoolVar(&out.KeepEmpty, "keep-empty", false, "leave ${VAR} literal if empty").
		Short("E").
		Value()

	// Allow lists
	fs.StringSliceVar(&out.Prefix, "prefix", nil, "only replace variables that match any of these prefixes").
		Short("p").
		Value()
	fs.StringSliceVar(&out.Suffix, "suffix", nil, "only replace variables that match any of these suffixes").
		Short("s").
		Value()
	fs.StringSliceVar(&out.Variables, "variable", nil, "only replace variables with these exact names").
		Short("v").
		Value()

	// Coloring
	fs.BoolVar(&out.Colored, "colored", false, "colorize formatter (content and diagnostics)").
		Short("c").
		OneOfGroup("mode").
		Value()

	// Vars files
	fs.StringSliceVar(&out.VarsFiles, "extra-vars", nil, "read variables from file (can be repeated)").
		Short("e").
		Placeholder("PATH...").
		Value()

	// Parse
	if err := fs.Parse(args); err != nil {
		return Options{}, err
	}
	out.Positional = fs.Args()

	// --strict implies both error flags
	if strict {
		out.ErrorUnset, out.ErrorEmpty = true, true
	}

	// --keep-vars implies both keep-*
	if keepVars {
		out.KeepUnset, out.KeepEmpty = true, true
	}

	// if --colored is set, it alwasys should show which variables are missing or empty
	if out.Colored {
		out.KeepUnset, out.KeepEmpty = true, true
	}

	return out, nil
}
