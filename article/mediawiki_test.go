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

func TestMathFallbackTeX_PlainImageUntouched(t *testing.T) {
	t.Parallel()

	blocks := normalizedBlocks(t, `<p><img src="https://example.com/cat.png" alt="A cat"/></p>`)

	require.Len(t, blocks, 1)
	assert.Equal(t, blockImage, blocks[0].kind)
	assert.Equal(t, "A cat", blocks[0].plainText())
}
