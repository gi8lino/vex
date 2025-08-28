package fsm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVarRef_Lit_Bare(t *testing.T) {
	t.Parallel()
	v := VarRef{Name: "FOO", Form: Bare}
	got := v.Lit()
	require.NotEmpty(t, got)
	assert.Equal(t, "$FOO", got)
}

func TestVarRef_Lit_Braced(t *testing.T) {
	t.Parallel()
	v := VarRef{Name: "FOO", Form: Braced}
	got := v.Lit()
	require.NotEmpty(t, got)
	assert.Equal(t, "${FOO}", got)
}

func TestBareRef(t *testing.T) {
	t.Parallel()
	v := BareRef("BAR")
	assert.Equal(t, "BAR", v.Name)
	assert.Equal(t, Bare, v.Form)
	assert.Equal(t, "$BAR", v.Lit())
}

func TestBracedRef(t *testing.T) {
	t.Parallel()
	v := BracedRef("BAR")
	assert.Equal(t, "BAR", v.Name)
	assert.Equal(t, Braced, v.Form)
	assert.Equal(t, "${BAR}", v.Lit())
}
