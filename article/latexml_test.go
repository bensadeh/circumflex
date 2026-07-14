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
