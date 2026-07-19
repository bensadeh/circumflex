package article

import (
	nurl "net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"golang.org/x/net/html"
)

func TestParseBlocks_CodeLanguage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		src  string
		want string
	}{
		{"code class", `<pre><code class="language-go">x</code></pre>`, "go"},
		{"pre class", `<pre class="lang-rust">x</pre>`, "rust"},
		{"hugo data-lang", `<pre><code data-lang="go">x</code></pre>`, "go"},
		{"github source class", `<pre><code class="highlight-source-c++">x</code></pre>`, "c++"},
		{"pinned attribute", `<pre data-clx-lang="python"><code>x</code></pre>`, "python"},
		{"uppercase normalized", `<pre><code class="language-Go">x</code></pre>`, "go"},
		{"unlabeled", `<pre><code>x</code></pre>`, ""},
		{"unrelated class", `<pre class="chroma"><code>x</code></pre>`, ""},
		{"bare prefix", `<pre><code class="language-">x</code></pre>`, ""},
		{"jekyll plaintext means unlabeled", `<pre><code class="language-plaintext">x</code></pre>`, ""},
		{"text means unlabeled", `<pre><code class="language-text">x</code></pre>`, ""},
		{"i18n wrapper is not a language", `<pre class="lang-en"><code>x</code></pre>`, ""},
		{"extension-fallback junk rejected", `<pre class="lang-es"><code>x</code></pre>`, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			blocks := blocksFromHTML(t, tt.src)

			require.Len(t, blocks, 1)
			require.Equal(t, blockCode, blocks[0].kind)
			assert.Equal(t, tt.want, blocks[0].lang)
		})
	}
}

func TestPreserveCodeLang_WrapperClass(t *testing.T) {
	t.Parallel()

	doc, err := html.Parse(strings.NewReader(`<div class="language-go highlighter-rouge">` +
		`<div class="highlight"><pre class="highlight"><code>x</code></pre></div></div>`))
	require.NoError(t, err)

	preserveCodeLang(doc)

	blocks := parseBlocks(doc)
	require.Len(t, blocks, 1)
	assert.Equal(t, "go", blocks[0].lang, "Rouge declares the language on a wrapper div")
}

func TestExtractReadable_CodeLanguageSurvives(t *testing.T) {
	t.Parallel()

	page := `<html><head><title>T</title></head><body><article>` +
		`<p>Enough readable content for the extractor to accept this page as an article, ` +
		`repeated to pass its length heuristics. Enough readable content for the extractor ` +
		`to accept this page as an article, repeated to pass its length heuristics.</p>` +
		`<pre><code class="language-go">func main() {}</code></pre>` +
		`</article></body></html>`

	u, err := nurl.Parse("https://example.com/post")
	require.NoError(t, err)

	node, _, err := extractReadable([]byte(page), u)
	require.NoError(t, err)

	var lang string

	for _, b := range parseBlocks(node) {
		if b.kind == blockCode {
			lang = b.lang
		}
	}

	assert.Equal(t, "go", lang, "language must survive readability's class stripping")
}
