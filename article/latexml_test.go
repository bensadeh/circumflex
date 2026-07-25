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
