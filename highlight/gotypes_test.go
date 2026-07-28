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

func TestCode_GoStructFieldNames(t *testing.T) {
	t.Parallel()

	out := Code("type RelatedInformation struct {\n    Pos     token.Pos\n    End     token.Pos // optional\n    Message string\n}", "go")

	assert.Contains(t, out, style.CodeLiteral("Pos"), "field names take the variable hue")
	assert.Contains(t, out, style.CodeLiteral("End"))
	assert.Contains(t, out, style.CodeLiteral("Message"))
	assert.Contains(t, out, style.CodeType("Pos"), "the field's type keeps the type hue")

	out = Code("type Point struct {\n    X, Y int\n}", "go")

	assert.Contains(t, out, style.CodeLiteral("X"))
	assert.Contains(t, out, style.CodeLiteral("Y"))
}

func TestCode_GoLiteralKeysMatchFields(t *testing.T) {
	t.Parallel()

	out := Code("a := &analysis.Analyzer{\n    Name: \"unusedresult\",\n    Run:  run,\n}", "go")

	assert.Contains(t, out, style.CodeLiteral("Name"),
		"a literal's key names the same field its declaration does")
	assert.Contains(t, out, style.CodeLiteral("Run"))
	assert.NotContains(t, out, style.CodeKeyword("Name"))
}

func TestCode_GoEmbeddedFieldsAreTypes(t *testing.T) {
	t.Parallel()

	out := Code("type W struct {\n    token.Pos\n    Reader\n    closer\n}", "go")

	assert.Contains(t, out, style.CodeType("Pos"), "embedded qualified type")
	assert.Contains(t, out, style.CodeType("Reader"), "a lone name on a field line is an embedded type")
	assert.NotContains(t, out, style.CodeLiteral("Reader"))
	assert.NotContains(t, out, style.CodeType("closer"), "lowercase embeds stay plain")
}

func TestCode_GoNestedStructFields(t *testing.T) {
	t.Parallel()

	out := Code("type Outer struct {\n    Meta struct {\n        Count int\n    }\n}", "go")

	assert.Contains(t, out, style.CodeLiteral("Meta"))
	assert.Contains(t, out, style.CodeLiteral("Count"))
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
