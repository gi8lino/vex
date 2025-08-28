package fsm

type VarForm uint8 // VarForm encodes how a variable reference should be rendered (bare or braced).

const (
	Bare   VarForm = iota // $VAR Bare renders a variable as $VAR.
	Braced                // ${VAR} Braced renders a variable as ${VAR}.
)

// VarRef holds a variable name and its rendering form.
type VarRef struct {
	Name string  // Name holds the variable name.
	Form VarForm // Form holds the rendering form.
}

// Lit returns the variable in its literal form ($VAR or ${VAR}).
func (v VarRef) Lit() string {
	if v.Form == Braced {
		return "${" + v.Name + "}"
	}
	return "$" + v.Name
}

// BareRef constructs a VarRef rendered as $NAME.
func BareRef(name string) VarRef { return VarRef{Name: name, Form: Bare} }

// BracedRef constructs a VarRef rendered as ${NAME}.
func BracedRef(name string) VarRef { return VarRef{Name: name, Form: Braced} }
