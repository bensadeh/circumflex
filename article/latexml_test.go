package article

import (
	nurl "net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFullTextURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		url  string
		want string
	}{
		{url: "https://arxiv.org/abs/2607.06377", want: "https://arxiv.org/html/2607.06377"},
		{url: "https://arxiv.org/abs/2607.06377v2", want: "https://arxiv.org/html/2607.06377v2"},
		{url: "https://www.arxiv.org/abs/2607.06377", want: "https://arxiv.org/html/2607.06377"},
		{url: "https://arxiv.org/pdf/2607.06377", want: "https://arxiv.org/html/2607.06377"},
		{url: "https://arxiv.org/pdf/2607.06377v1.pdf", want: "https://arxiv.org/html/2607.06377v1"},
		{url: "https://arxiv.org/abs/quant-ph/0410100", want: "https://arxiv.org/html/quant-ph/0410100"},
		{url: "https://arxiv.org/abs/2607.06377?context=math", want: "https://arxiv.org/html/2607.06377"},
		{url: "https://export.arxiv.org/abs/2607.06377", want: "https://arxiv.org/html/2607.06377"},
		{url: "https://arxiv.org/html/2607.06377", want: ""},
		{url: "https://arxiv.org/list/math.HO/recent", want: ""},
		{url: "https://arxiv.org", want: ""},
		{url: "https://example.com/abs/2607.06377", want: ""},
		{url: "https://notarxiv.org/abs/2607.06377", want: ""},
	}

	for _, tt := range tests {
		parsed, err := nurl.Parse(tt.url)
		require.NoError(t, err)

		assert.Equal(t, tt.want, fullTextURL(parsed), tt.url)
	}
}

func TestExtractReadable_RawPictureCollapsesToFigure(t *testing.T) {
	t.Parallel()

	// When LaTeXML cannot convert a picture environment it leaves the raw
	// LaTeX source as the ltx_picture span's text and never copies the
	// referenced graphic into the asset tree. The source must not render as
	// prose; the figure collapses to its caption under the Figure designation.
	page := `<html><head><title>T</title></head><body><article>` +
		`<p>Enough readable content for the extractor to accept this page as an article, ` +
		`repeated to pass its length heuristics. Enough readable content for the extractor ` +
		`to accept this page as an article, repeated to pass its length heuristics.</p>` +
		`<figure id="S2.F1" class="ltx_figure">` +
		`<span id="S2.F1.pic1" class="ltx_picture ltx_centering" style="width:346.2pt;height:87.4pt;">` +
		`\begin{overpic}[width=346.89731pt]{covering-space-landscape.png}` + "\n" +
		`\put(-7.0,14.0){$\dots$}` + "\n" +
		`\par\put(1.5,15.5){\tiny$+$}` + "\n" +
		`\par\end{overpic}</span>` +
		`<figcaption class="ltx_caption ltx_centering">Figure 2.1: The universal cyclic cover of the disk.</figcaption>` +
		`</figure>` +
		`</article></body></html>`

	u, err := nurl.Parse("https://arxiv.org/html/2607.05283")
	require.NoError(t, err)

	node, _, err := extractReadable([]byte(page), u)
	require.NoError(t, err)

	var figure *block

	for _, b := range parseBlocks(node) {
		if b.kind == blockImage {
			figure = &b

			break
		}

		assert.NotContains(t, spanText(b.spans), `\begin{overpic}`, "raw LaTeX must not render as prose")
		assert.NotContains(t, b.text, `\begin{overpic}`, "raw LaTeX must not render as prose")
	}

	require.NotNil(t, figure, "the caption must survive as a designated figure")
	assert.True(t, figure.figure)
	assert.Empty(t, figure.imageURL)
	assert.Equal(t, "Figure 2.1: The universal cyclic cover of the disk.", spanText(figure.spans))
}

func TestExtractReadable_GuessedHeadersTableSurvives(t *testing.T) {
	t.Parallel()

	// LaTeXML marks tables with inferred header rows ltx_guessed_headers;
	// readability's unlikely-candidate regex reads the "header" substring as
	// page chrome and deletes the whole table unless the marker is scrubbed.
	page := `<html><head><title>T</title></head><body><article>` +
		`<p>Enough readable content for the extractor to accept this page as an article, ` +
		`repeated to pass its length heuristics. Enough readable content for the extractor ` +
		`to accept this page as an article, repeated to pass its length heuristics.</p>` +
		`<figure class="ltx_table">` +
		`<table class="ltx_tabular ltx_centering ltx_guessed_headers ltx_align_middle">` +
		`<thead class="ltx_thead"><tr><th class="ltx_td ltx_th">Label</th><th class="ltx_td ltx_th">Criterion</th></tr></thead>` +
		`<tbody class="ltx_tbody"><tr><td class="ltx_td">Handled</td><td class="ltx_td">Agent reaches the challenge.</td></tr></tbody>` +
		`</table>` +
		`<figcaption class="ltx_caption">Table 4: Exposure attribution labels.</figcaption>` +
		`</figure>` +
		`</article></body></html>`

	u, err := nurl.Parse("https://arxiv.org/html/2606.29537")
	require.NoError(t, err)

	node, _, err := extractReadable([]byte(page), u)
	require.NoError(t, err)

	var table *block

	for _, b := range parseBlocks(node) {
		if b.kind == blockTable {
			table = &b

			break
		}
	}

	require.NotNil(t, table, "the table must survive readability's chrome heuristics")
	assert.True(t, table.hasHeader)
	require.Len(t, table.rows, 2)
	assert.Equal(t, []string{"Label", "Criterion"}, table.rows[0])
}
