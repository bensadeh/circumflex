package article

import (
	nurl "net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateURL(t *testing.T) {
	t.Parallel()

	require.NoError(t, ValidateURL("https://example.com/some/page"))
	require.NoError(t, ValidateURL("http://blog.example.org/post"))

	require.Error(t, ValidateURL("https://youtube.com/watch?v=1"), "blocklisted domain")
	require.Error(t, ValidateURL("https://www.reddit.com/r/golang"), "www prefix still hits the blocklist")
	require.Error(t, ValidateURL("ftp://example.com/file"), "non-http scheme")
	require.Error(t, ValidateURL("not a url"))
}

func TestValidateURL_NonParseableExtensions(t *testing.T) {
	t.Parallel()

	err := ValidateURL("https://www.bosch-sensortec.com/media/boschsensortec/downloads/datasheets/bst-bme280-ds002.pdf")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "PDFs")

	require.Error(t, ValidateURL("https://example.com/REPORT.PDF"), "extension check is case-insensitive")
	require.Error(t, ValidateURL("https://example.com/release.zip"))
	require.Error(t, ValidateURL("https://example.com/firmware.bin"))
	require.Error(t, ValidateURL("https://example.com/photo.png"))

	require.NoError(t, ValidateURL("https://example.com/download?file=x.pdf"),
		"only the path counts, not the query")
	require.NoError(t, ValidateURL("https://example.com/whitepaper.pdf.html"),
		"only the final extension counts")
}

func TestValidateURL_ArxivPDFExemption(t *testing.T) {
	t.Parallel()

	require.NoError(t, ValidateURL("https://arxiv.org/pdf/2401.12345"),
		"arXiv PDF links fetch through the HTML full-text mirror")
	require.NoError(t, ValidateURL("https://arxiv.org/pdf/2401.12345v2.pdf"))
	require.NoError(t, ValidateURL("https://www.arxiv.org/pdf/2401.12345.pdf"))
}

func TestExtractReadable_ReturnsTitle(t *testing.T) {
	t.Parallel()

	page := `<html><head><title>The Page Title</title></head><body><article>` +
		`<h1>The Page Title</h1>` +
		`<p>Enough readable content for the extractor to accept this page as an article, ` +
		`repeated to pass its length heuristics. Enough readable content for the extractor ` +
		`to accept this page as an article, repeated to pass its length heuristics.</p>` +
		`</article></body></html>`

	u, err := nurl.Parse("https://example.com/post")
	require.NoError(t, err)

	node, title, err := extractReadable([]byte(page), u)
	require.NoError(t, err)
	require.NotNil(t, node)
	assert.Equal(t, "The Page Title", title)
}
