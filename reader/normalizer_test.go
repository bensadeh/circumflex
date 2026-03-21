package reader

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizeHeaders(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		blocks []*block
		want   []int
	}{
		{
			name:   "no blocks",
			blocks: []*block{},
			want:   []int{},
		},
		{
			name: "no headers",
			blocks: []*block{
				{Kind: blockText, Text: "hello"},
				{Kind: blockCode, Text: "code"},
			},
			want: []int{blockText, blockCode},
		},
		{
			name: "already normalized h1 h2 h3",
			blocks: []*block{
				{Kind: blockH1, Text: "# A"},
				{Kind: blockH2, Text: "## B"},
				{Kind: blockH3, Text: "### C"},
			},
			want: []int{blockH1, blockH2, blockH3},
		},
		{
			name: "shift up h3 h4",
			blocks: []*block{
				{Kind: blockH3, Text: "### A"},
				{Kind: blockH4, Text: "#### B"},
			},
			want: []int{blockH1, blockH2},
		},
		{
			name: "gap collapse h1 h2 h4",
			blocks: []*block{
				{Kind: blockH1, Text: "# A"},
				{Kind: blockH2, Text: "## B"},
				{Kind: blockH4, Text: "#### C"},
			},
			want: []int{blockH1, blockH2, blockH3},
		},
		{
			name: "shift and collapse h3 h5 h6",
			blocks: []*block{
				{Kind: blockH3, Text: "### A"},
				{Kind: blockH5, Text: "##### B"},
				{Kind: blockH6, Text: "###### C"},
			},
			want: []int{blockH1, blockH2, blockH3},
		},
		{
			name: "single header h4 becomes h1",
			blocks: []*block{
				{Kind: blockH4, Text: "#### Only"},
			},
			want: []int{blockH1},
		},
		{
			name: "mixed blocks only headers remapped",
			blocks: []*block{
				{Kind: blockText, Text: "intro"},
				{Kind: blockH3, Text: "### Title"},
				{Kind: blockCode, Text: "code()"},
				{Kind: blockH4, Text: "#### Sub"},
				{Kind: blockList, Text: "- item"},
			},
			want: []int{blockText, blockH1, blockCode, blockH2, blockList},
		},
		{
			name: "all six levels stay unchanged",
			blocks: []*block{
				{Kind: blockH1, Text: "# A"},
				{Kind: blockH2, Text: "## B"},
				{Kind: blockH3, Text: "### C"},
				{Kind: blockH4, Text: "#### D"},
				{Kind: blockH5, Text: "##### E"},
				{Kind: blockH6, Text: "###### F"},
			},
			want: []int{blockH1, blockH2, blockH3, blockH4, blockH5, blockH6},
		},
		{
			name: "duplicate levels",
			blocks: []*block{
				{Kind: blockH3, Text: "### A"},
				{Kind: blockH3, Text: "### B"},
				{Kind: blockH5, Text: "##### C"},
			},
			want: []int{blockH1, blockH1, blockH2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			normalizeHeaders(tt.blocks)

			require.Len(t, tt.blocks, len(tt.want))

			for i, b := range tt.blocks {
				assert.Equal(t, tt.want[i], b.Kind, "block %d", i)
			}
		})
	}
}

func TestIsHeader(t *testing.T) {
	t.Parallel()

	tests := []struct {
		kind int
		want bool
	}{
		{blockText, false},
		{blockImage, false},
		{blockH1, true},
		{blockH2, true},
		{blockH3, true},
		{blockH4, true},
		{blockH5, true},
		{blockH6, true},
		{blockQuote, false},
		{blockCode, false},
		{blockList, false},
		{blockTable, false},
		{blockDivider, false},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.want, isHeader(tt.kind), "kind: %d", tt.kind)
	}
}
