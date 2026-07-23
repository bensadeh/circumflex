package article

import (
	"image"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExpandSVGUse_LeavesInstanceFreeSourceAlone(t *testing.T) {
	t.Parallel()

	assert.Nil(t, expandSVGUse([]byte(`<svg><rect width="1" height="1"/></svg>`)),
		"a document canvas already draws whole must reach it untouched")
}

func TestExpandSVGUse_InlinesReferencedElement(t *testing.T) {
	t.Parallel()

	out := expandSVGUse([]byte(`<svg><defs><path id="a" d="M0 0"/></defs>` +
		`<use xlink:href="#a" x="3" y="4" style="fill: #f00"/></svg>`))

	assert.Equal(t, `<svg><defs><path id="a" d="M0 0"/></defs>`+
		`<g transform="translate(3 4)" style="fill: #f00"><path id="a" d="M0 0"/></g></svg>`,
		string(out), "the instance becomes the group it stands for, presentation and all")
}

func TestExpandSVGUse_PlacesInstanceAfterItsOwnTransform(t *testing.T) {
	t.Parallel()

	out := expandSVGUse([]byte(`<svg><defs><path id="a" d="M0 0"/></defs>` +
		`<use href="#a" transform="scale(2)" x="3" y="4"/></svg>`))

	assert.Contains(t, string(out), `<g transform="scale(2) translate(3 4)">`,
		"the offset applies in the instance's own coordinate system, so it composes last")

	out = expandSVGUse([]byte(`<svg><defs><path id="a" d="M0 0"/></defs><use href="#a"/></svg>`))

	assert.Contains(t, string(out), `<g><path id="a" d="M0 0"/></g>`,
		"an instance that neither offsets nor transforms carries no transform")
}

// Every reference this declines is one canvas drops today, so declining costs
// nothing that was not already lost.
func TestExpandSVGUse_DeclinesReferencesItCannotResolve(t *testing.T) {
	t.Parallel()

	assert.Nil(t, expandSVGUse([]byte(`<svg><defs><path id="a" d="M0 0"/></defs>`+
		`<use href="sprites.svg#a"/></svg>`)), "a reference into another file is not ours to resolve")

	assert.Nil(t, expandSVGUse([]byte(`<svg><defs><path id="a" d="M0 0"/></defs>`+
		`<use href="#missing"/></svg>`)), "a reference to nothing stays a reference to nothing")

	assert.Nil(t, expandSVGUse([]byte(`<svg><defs><symbol id="a"><rect width="1" height="1"/></symbol></defs>`+
		`<use href="#a"/></svg>`)), "a symbol establishes its own viewport, which this does not place")

	assert.Nil(t, expandSVGUse([]byte(`<svg><defs><path id="a" d="M0 0"/></defs>`+
		`<use href="#a" x="50%"/></svg>`)), "a percentage offset resolves against the viewport, not a transform")
}

// canvas skips what <defs> encloses but implements no clip path, so it already
// paints one written outside a <defs>. Expanding an instance in there would
// add ink of our own to that.
func TestExpandSVGUse_LeavesInstancesThatDefineRatherThanDraw(t *testing.T) {
	t.Parallel()

	const defs = `<svg><defs><path id="a" d="M0 0"/></defs>`

	assert.Nil(t, expandSVGUse([]byte(defs+`<clipPath id="c"><use href="#a"/></clipPath></svg>`)),
		"a clip path names a region; it does not draw one")

	assert.NotNil(t, expandSVGUse([]byte(defs+`<clipPath id="c"><rect width="1" height="1"/></clipPath>`+
		`<use href="#a"/></svg>`)), "an instance beside it still expands")
}

func TestExpandSVGUse_ResolvesInstancesInsideInstances(t *testing.T) {
	t.Parallel()

	out := expandSVGUse([]byte(`<svg><defs><path id="p" d="M0 0"/>` +
		`<g id="grp"><use href="#p"/></g></defs><use href="#grp" x="5"/></svg>`))

	require.NotNil(t, out)
	assert.NotContains(t, string(out), "<use", "a later pass resolves the instances a clone carries in")
	assert.Equal(t, 3, strings.Count(string(out), `d="M0 0"`),
		"the definition, the instance inside the group, and the group's clone")
}

// A group that instantiates itself resolves forever; the pass and byte caps
// are what make an untrusted document safe to rewrite at all.
func TestExpandSVGUse_TerminatesOnSelfReference(t *testing.T) {
	t.Parallel()

	out := expandSVGUse([]byte(`<svg><defs><g id="a"><use href="#a"/></g></defs><use href="#a"/></svg>`))

	require.NotNil(t, out)
	assert.Less(t, len(out), maxUseBytes)
	assert.Contains(t, string(out), "<use", "the chain is cut off, not completed")
}

// Spans are derived from the nesting, so tags that do not balance leave this
// unable to say which bytes are the element it would clone.
func TestExpandSVGUse_DeclinesUnbalancedSource(t *testing.T) {
	t.Parallel()

	const defs = `<svg><defs><path id="a" d="M0 0"/></defs>`

	assert.Nil(t, expandSVGUse([]byte(defs+`<g><use href="#a"/></svg>`)), "an unclosed group")
	assert.Nil(t, expandSVGUse([]byte(defs+`<use href="#a"/></g></svg>`)), "an unopened group")
	assert.NotNil(t, expandSVGUse([]byte(defs+`<g><use href="#a"/></g></svg>`)), "the same source, balanced")
}

// The end-to-end regression: matplotlib's default svg.fonttype emits every
// glyph as a path under <defs> plus one <use> per occurrence, so a plot whose
// only ink arrives this way used to rasterize blank.
func TestDecodeSVG_DrawsInstancedShapes(t *testing.T) {
	t.Parallel()

	img, viewBox := decodeSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" ` +
		`xmlns:xlink="http://www.w3.org/1999/xlink" viewBox="0 0 100 50">` +
		`<defs><path id="sq" d="M 0 0 L 20 0 L 20 20 L 0 20 z"/></defs>` +
		`<use xlink:href="#sq" x="40" y="15" style="fill: #ff0000"/></svg>`))

	require.NotNil(t, img)
	assert.Equal(t, image.Pt(100, 50), viewBox)

	r, g, b, a := img.At(img.Bounds().Dx()/2, img.Bounds().Dy()/2).RGBA()
	assert.Equal(t, []uint32{0xffff, 0x0, 0x0, 0xffff}, []uint32{r, g, b, a},
		"the instance is drawn where its offset places it, in the fill it carries")
}
