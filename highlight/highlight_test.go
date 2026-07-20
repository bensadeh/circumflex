package highlight

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bensadeh/circumflex/ansi"
	"github.com/bensadeh/circumflex/style"
)

func TestCode_UnknownLanguageFallsBack(t *testing.T) {
	t.Parallel()

	assert.Empty(t, Code("func main() {}", ""))
	assert.Empty(t, Code("func main() {}", "not-a-language"))
}

func TestCode_StripsBackToSource(t *testing.T) {
	t.Parallel()

	// Spaces, not tabs, like real block text: preText expands tabs to
	// 8-column stops before any block is stored.
	src := "func main() {\n    fmt.Println(\"hi\") // greet\n}"

	assert.Equal(t, src, ansi.Strip(Code(src, "go")))
}

func TestCode_TokenStyles(t *testing.T) {
	t.Parallel()

	out := Code(`func f() { s := "hi" } // note`, "go")

	assert.Contains(t, out, style.CodeKeyword("func"))
	assert.Contains(t, out, style.CodeString(`"hi"`))
	assert.Contains(t, out, style.CodeComment("// note"))
}

func TestCode_DiffColors(t *testing.T) {
	t.Parallel()

	out := Code("diff --git a/x b/x\n@@ -1 +1 @@\n-old\n+new", "diff")

	assert.Contains(t, out, style.CodeDeleted("-old"))
	assert.Contains(t, out, style.CodeString("+new"))
	assert.Contains(t, out, style.Faint("@@ -1 +1 @@"), "hunk headers are metadata, not content")
}

func TestCode_PreprocessorIsNotAComment(t *testing.T) {
	t.Parallel()

	out := Code("#include <stdio.h>\nint main(void) { return 0; }", "c")

	assert.Contains(t, out, style.CodeKeyword("#include"))
	assert.Contains(t, out, style.CodeString("<stdio.h>"))
	assert.NotContains(t, out, style.CodeComment("#include"))
}

func TestCode_TypesFunctionsEscapes(t *testing.T) {
	t.Parallel()

	out := Code("#include <stdio.h>\nint main(void) {\n    printf(\"hi\\n\");\n}", "c")

	assert.Contains(t, out, style.CodeType("int"))
	assert.Contains(t, out, style.CodeFunction("printf"))
	assert.Contains(t, out, style.CodeEscape(`\n`))

	out = Code("@cache\ndef f(x):\n    return f\"{x!r}\"", "python")

	assert.Contains(t, out, style.CodeFunction("@cache"))
	assert.Contains(t, out, style.CodeEscape("{"), "f-string interpolation stands apart from the string")
}

func TestCode_RegexExceptionsEntities(t *testing.T) {
	t.Parallel()

	out := Code(`const re = /ab+c/gi; s.replace(re, "x");`, "javascript")
	assert.Contains(t, out, style.CodeEscape("/ab+c/gi"))

	out = Code("try:\n    raise ValueError(\"bad\")\nexcept KeyError:\n    pass", "python")
	assert.Contains(t, out, style.CodeEscape("ValueError"))
	assert.Contains(t, out, style.CodeEscape("KeyError"))

	out = Code("<p>a &amp; b</p>", "html")
	assert.Contains(t, out, style.CodeEscape("&amp;"))
}

func TestCode_ShellVariablesAndBuiltins(t *testing.T) {
	t.Parallel()

	out := Code("for f in *.txt\ndo\n  echo $f\ndone", "bash")

	assert.Contains(t, out, style.CodeLiteral("$f"))
	assert.Contains(t, out, style.CodeLiteral("echo"))
}

func TestCode_RustCapitalizedTypes(t *testing.T) {
	t.Parallel()

	out := Code("let osrng = OsRng::new().unwrap();\nlet prng = ChaChaRng::from_seed(&key[..])", "rust")

	assert.Contains(t, out, style.CodeType("OsRng"),
		"Rust's convention makes capitalized names types; the lexer leaves them plain")
	assert.Contains(t, out, style.CodeType("ChaChaRng"))
	assert.NotContains(t, out, style.CodeType("osrng"), "lowercase names stay plain")
}

func TestCode_ConventionRefinements(t *testing.T) {
	t.Parallel()

	out := Code("StringBuilder sb = new StringBuilder();", "java")
	assert.Contains(t, out, style.CodeType("StringBuilder"))

	out = Code("val user = User(name)", "kotlin")
	assert.Contains(t, out, style.CodeType("User"))

	out = Code("MAX_RETRIES = 5\nresult = retry(MAX_RETRIES)", "python")
	assert.Contains(t, out, style.CodeLiteral("MAX_RETRIES"), "PEP 8 makes ALL_CAPS a constant")

	out = Code("const MAX: usize = 8;\nlet t = Tree::new();", "rust")
	assert.Contains(t, out, style.CodeLiteral("MAX"), "ALL_CAPS wins over the capitalized-type rule")
	assert.Contains(t, out, style.CodeType("Tree"))
}

func TestCode_StrictLexerResolution(t *testing.T) {
	t.Parallel()

	// chroma's Get would route these through its file-extension fallback
	// (mod → AMPL, es → Erlang); a declaration that is not a real name or
	// alias must fall back to plain instead.
	assert.Empty(t, Code("module example.com/x\n\ngo 1.26", "mod"))
	assert.Empty(t, Label("mod"))
	assert.Empty(t, Code("texto", "es"))
	assert.Empty(t, Label("es"))
	assert.False(t, Resolves("mod"))
	assert.True(t, Resolves("golang"), "real aliases still resolve")
}

func TestCode_JavaCapsClassesAreTypes(t *testing.T) {
	t.Parallel()

	out := Code("URL url = new URL(spec);\nint m = MAX_VALUE;", "java")

	assert.Contains(t, out, style.CodeType("URL"),
		"acronym classes read as types; only underscored ALL_CAPS are constants")
	assert.Contains(t, out, style.CodeLiteral("MAX_VALUE"))
}

func TestCode_SingleGreekCapitalStaysPlain(t *testing.T) {
	t.Parallel()

	out := Code("Σ = compute()\nprint(Σ)", "python")

	assert.NotContains(t, out, style.CodeLiteral("Σ"),
		"a single rune is not ALL_CAPS whatever its byte length")
}

func TestCode_ShellFlags(t *testing.T) {
	t.Parallel()

	out := Code("curl -fsSL --retry 3 \\\n  --output f.tar.gz https://a.io/b-c", "bash")
	assert.Contains(t, out, style.CodeFunction("-fsSL"))
	assert.Contains(t, out, style.CodeFunction("--retry"))
	assert.Contains(t, out, style.CodeFunction("--output"))
	assert.Contains(t, out, style.CodeLiteral(`\`), "the continuation backslash is shell machinery, not an escape")
	assert.NotContains(t, out, style.CodeFunction("-c"), "url dashes are not flags")

	out = Code("head -1 notes-2026-07.txt", "bash")
	assert.Contains(t, out, style.CodeFunction("-1"), "numeric flags count")
	assert.NotContains(t, out, style.CodeFunction("-2026"), "date dashes are not flags")

	out = Code(`git checkout -- main.go
echo \$HOME`, "bash")
	assert.Contains(t, out, style.CodeFunction("--"), "the end-of-options separator is option syntax")
	assert.Contains(t, out, style.CodeEscape(`\$`), "only the line continuation goes cyan, escapes keep red")
}

func TestCode_ConsoleContinuationsStayCommands(t *testing.T) {
	t.Parallel()

	out := Code("$ make build --jobs 4 \\\n    --keep-going \\\n    --silent\nok\n", "console")

	assert.Contains(t, out, style.CodeFunction("--jobs"))
	assert.Contains(t, out, style.CodeFunction("--keep-going"),
		"continuation lines re-lex as command, not output")
	assert.Contains(t, out, style.CodeFunction("--silent"), "chained continuations too")
	assert.NotContains(t, out, style.CodeFunction("ok"), "program output stays plain")
}

func TestCode_ConfigKeys(t *testing.T) {
	t.Parallel()

	out := Code("[[application_scanner]]\ntype = \"steam\"\nresolve_icons = true", "toml")
	assert.Contains(t, out, style.CodeKeyword("application_scanner"), "table headers stand apart from keys")
	assert.Contains(t, out, style.CodeFunction("type"))
	assert.Contains(t, out, style.CodeFunction("resolve_icons"))
	assert.Contains(t, out, style.CodeString(`"steam"`))
	assert.NotContains(t, out, style.CodeKeyword("[["), "brackets stay plain like all punctuation")

	out = Code("[servers.alpha]\nip = \"10.0.0.1\"\ndirs = [\n  \"a\",\n]\nnext = 1", "toml")
	assert.Contains(t, out, style.CodeKeyword("servers"), "every dotted-header segment is a header")
	assert.Contains(t, out, style.CodeKeyword("alpha"))
	assert.Contains(t, out, style.CodeFunction("ip"))
	assert.Contains(t, out, style.CodeFunction("next"), "a line-leading array close is not a header open")

	out = Code("[server]\nport = 8080", "ini")
	assert.Contains(t, out, style.CodeKeyword("[server]"), "ini sections match toml headers")
	assert.Contains(t, out, style.CodeFunction("port"))

	out = Code(`<a href="x">t</a>`, "html")
	assert.Contains(t, out, style.CodeFunction("href"), "attributes are keys too")
}

func TestCode_JSXTags(t *testing.T) {
	t.Parallel()

	out := Code(`const x = <li className="a">{item.text}</li>`, "jsx")

	assert.Contains(t, out, style.CodeKeyword("li"), "JSX tags color like html tags")
}

func TestCode_CommentDebtMarkers(t *testing.T) {
	t.Parallel()

	out := Code("// TODO: fix this properly\nint x;", "c")

	assert.Contains(t, out, style.CodeEscape("TODO"))
	assert.Contains(t, out, style.CodeComment("// "), "the rest of the comment stays quiet")
	assert.Contains(t, out, style.CodeComment(": fix this properly"))
}

func TestCode_MultilineTokenStyledPerLine(t *testing.T) {
	t.Parallel()

	out := Code("s := `a\nb`", "go")

	lines := strings.Split(out, "\n")
	require.Len(t, lines, 2)

	assert.Contains(t, lines[0], style.CodeString("`a"),
		"each line must carry its own escapes; the box splices border cells between lines")
	assert.Contains(t, lines[1], style.CodeString("b`"))
}
