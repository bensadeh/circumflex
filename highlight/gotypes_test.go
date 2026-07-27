package highlight

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bensadeh/circumflex/style"
)

func TestCode_GoDeclaredTypeNames(t *testing.T) {
	t.Parallel()

	out := Code("type Diagnostic struct {\n    Pos      token.Pos\n    Category string // optional\n}", "go")

	assert.Contains(t, out, style.CodeType("Diagnostic"))
	assert.Contains(t, out, style.CodeType("Pos"), "the field's type, promoted by adjacency")
	assert.NotContains(t, out, style.CodeType("token"), "package qualifiers stay plain")
	assert.NotContains(t, out, style.CodeType("Category"), "field names stay plain")

	out = Code("type isWrapper struct{} // => wrapper fact", "go")

	assert.Contains(t, out, style.CodeType("isWrapper"),
		"the name after the type keyword skips the capitalization gate")
}

func TestCode_GoSigilGluedTypes(t *testing.T) {
	t.Parallel()

	out := Code("type Pass struct {\n    Fset     *token.FileSet\n    Requires []*Analyzer\n    ResultOf map[*Analyzer]bool\n    Ch       chan Item\n}", "go")

	assert.Contains(t, out, style.CodeType("FileSet"))
	assert.Contains(t, out, style.CodeType("Analyzer"))
	assert.Contains(t, out, style.CodeType("Item"))
}

func TestCode_GoSignatureTypes(t *testing.T) {
	t.Parallel()

	out := Code("func (pass *Pass) ReportRangef(rng Range, format string, opts ...Option) *Result", "go")

	assert.Contains(t, out, style.CodeType("Pass"), "receiver type")
	assert.Contains(t, out, style.CodeType("Range"), "parameter type")
	assert.Contains(t, out, style.CodeType("Option"), "variadic element type")
	assert.Contains(t, out, style.CodeType("Result"), "return type after the closing paren")
	assert.NotContains(t, out, style.CodeType("rng"))
}

// The positions this pass deliberately leaves alone: expressions that share a
// shape with types. A spaced * is arithmetic, a * behind an operator is a
// deref, and calls, literals and lone identifiers carry no declaration
// adjacency.
func TestCode_GoExpressionsStayPlain(t *testing.T) {
	t.Parallel()

	out := Code("area := w * Height\nv := *Ptr\nfmt.Println(Verbose)\nu := User{Name: n}", "go")

	assert.NotContains(t, out, style.CodeType("Height"), "multiplication, not a pointer type")
	assert.NotContains(t, out, style.CodeType("Ptr"), "deref follows an operator, never a name")
	assert.NotContains(t, out, style.CodeType("Verbose"))
	assert.NotContains(t, out, style.CodeType("User"), "composite literals need a parser to tell from vars")
}
