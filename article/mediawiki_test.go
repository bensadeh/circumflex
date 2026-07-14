package article

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"golang.org/x/net/html"
)

func normalizedBlocks(t *testing.T, src string) []block {
	t.Helper()

	node, err := html.Parse(strings.NewReader(src))
	require.NoError(t, err)

	normalizeMediaWiki(node)

	return parseBlocks(node)
}

func TestNormalizeMediaWiki_HeadingWrapper(t *testing.T) {
	t.Parallel()

	blocks := normalizedBlocks(t, `
		<div class="mw-heading mw-heading2"><h2 id="See_also">See also</h2><span
			class="mw-editsection"><span class="mw-editsection-bracket">[</span><a
			href="https://en.wikipedia.org/w/index.php?action=edit"><span>edit</span></a><span
			class="mw-editsection-bracket">]</span></span></div>
		<p>content</p>`)

	require.Len(t, blocks, 2)
	assert.Equal(t, blockHeading, blocks[0].kind)
	assert.Equal(t, "See also", blocks[0].text)
	assert.Equal(t, "content", blocks[1].plainText())
}

func TestNormalizeMediaWiki_MathWrapper(t *testing.T) {
	t.Parallel()

	blocks := normalizedBlocks(t, `
		<p>the union <span class="mwe-math-element"><span class="mwe-math-mathml-inline
			mwe-math-mathml-a11y" style="display: none;"><math alttext="{\displaystyle A+B}">
			<semantics><mi>A</mi><mo>+</mo><mi>B</mi><annotation
			encoding="application/x-tex">{\displaystyle A+B}</annotation></semantics>
			</math></span><img src="https://wikimedia.org/api/rest_v1/media/math/render/svg/42"
			class="mwe-math-fallback-image-inline" alt="{\displaystyle A+B}"/></span> is tagged</p>`)

	require.Len(t, blocks, 1)
	assert.Equal(t, "the union 𝐴+𝐵 is tagged", blocks[0].plainText())
}

// A math wrapper without MathML keeps its fallback image, which renders as
// TeX through the image alt path instead.
func TestNormalizeMediaWiki_ImageOnlyMath(t *testing.T) {
	t.Parallel()

	blocks := normalizedBlocks(t, `
		<p>from <img src="https://wikimedia.org/api/rest_v1/media/math/render/svg/7d"
			class="mwe-math-fallback-image-inline" alt="{\displaystyle A}"/> instead</p>`)

	require.Len(t, blocks, 1)
	assert.Equal(t, "from 𝐴 instead", blocks[0].plainText())
}

func TestMathFallbackTeX_WordPressLatexInline(t *testing.T) {
	t.Parallel()

	blocks := normalizedBlocks(t, `
		<p>is a functor <img decoding="async"
			src="https://s0.wp.com/latex.php?latex=%5Cotimes+%5Ccolon+%5Cmathbf+M&bg=ffffff&fg=29303b&s=0"
			alt="\otimes \colon \mathbf M \times \mathbf M \to \mathbf M" class="latex"/>. We assume</p>`)

	require.Len(t, blocks, 1)
	assert.Equal(t, "is a functor ⊗ : M × M → M. We assume", blocks[0].plainText())
}

func TestMathFallbackTeX_WordPressLatexDisplay(t *testing.T) {
	t.Parallel()

	blocks := normalizedBlocks(t, `
		<p class="has-text-align-center"><img decoding="async"
			src="https://s0.wp.com/latex.php?latex=%5Clambda_a+%5Ccolon+1+%5Cotimes+a+%5Cto+a&bg=ffffff"
			alt="\lambda_a \colon 1 \otimes a \to a" class="latex"/></p>`)

	require.Len(t, blocks, 1)
	assert.Equal(t, blockParagraph, blocks[0].kind)
	assert.Equal(t, "λₐ : 1 ⊗ 𝑎 → 𝑎", blocks[0].plainText())
}

func TestMathFallbackTeX_CodecogsAlt(t *testing.T) {
	t.Parallel()

	blocks := normalizedBlocks(t, `
		<p>the value <img src="https://latex.codecogs.com/svg.latex?x%5E2" alt="x^2"/> grows</p>`)

	require.Len(t, blocks, 1)
	assert.Equal(t, "the value 𝑥² grows", blocks[0].plainText())
}

// Hand-pasted codecogs embeds often carry no alt attribute; the TeX is
// recovered from the URL instead, directives and &space; encoding included.
func TestMathFallbackTeX_CodecogsURLOnly(t *testing.T) {
	t.Parallel()

	blocks := normalizedBlocks(t, `
		<p>so <img src="https://latex.codecogs.com/svg.image?\dpi{110}&space;x&space;=&space;2"/> holds</p>`)

	require.Len(t, blocks, 1)
	assert.Equal(t, "so 𝑥 = 2 holds", blocks[0].plainText())
}

func TestMathFallbackTeX_QuickLaTeX(t *testing.T) {
	t.Parallel()

	blocks := normalizedBlocks(t, `
		<p>group <img class="ql-img-inline-formula" src="https://quicklatex.com/cache3/a1/ql_a1.png"
			alt="\Omega"/> here</p>
		<p><img class="ql-img-displayed-equation" src="https://quicklatex.com/cache3/b2/ql_b2.png"
			alt="Rendered by QuickLaTeX.com"/></p>`)

	require.Len(t, blocks, 2)
	assert.Equal(t, "group Ω here", blocks[0].plainText())
	assert.Equal(t, blockImage, blocks[1].kind, "display equations without source stay images")
}

func TestMathFallbackTeX_PlainImageUntouched(t *testing.T) {
	t.Parallel()

	blocks := normalizedBlocks(t, `<p><img src="https://example.com/cat.png" alt="A cat"/></p>`)

	require.Len(t, blocks, 1)
	assert.Equal(t, blockImage, blocks[0].kind)
	assert.Equal(t, "A cat", blocks[0].plainText())
}
