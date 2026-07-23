package highlight

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bensadeh/circumflex/ansi"
)

func TestLabel(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "Go", ansi.Strip(Label("go")))
	assert.Equal(t, "C++", ansi.Strip(Label("cpp")), "aliases resolve to the canonical name")
	assert.Equal(t, "C#", ansi.Strip(Label("csharp")))
	assert.NotEqual(t, Label("csharp"), ansi.Strip(Label("csharp")), "C# carries a brand color")
	assert.Equal(t, "Shell", ansi.Strip(Label("console")), "session blocks read as Shell, not Bash Session")
	assert.Equal(t, "JSX", ansi.Strip(Label("jsx")), "the react lexer's lowercase name reads as JSX")
	assert.Equal(t, "TSX", ansi.Strip(Label("tsx")), "the tsx alias keeps its own name, not TypeScript")
	assert.Equal(t, "TypeScript", ansi.Strip(Label("ts")))
	assert.Equal(t, "Markdown", ansi.Strip(Label("markdown")), "chroma's lowercase lexer name is capitalized")
	assert.Empty(t, Label("not-a-language"))
	assert.Empty(t, Label(""))

	assert.NotEqual(t, Label("go"), ansi.Strip(Label("go")), "known languages carry their brand color")
}
