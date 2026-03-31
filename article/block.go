package article

const (
	blockText    = 0
	blockImage   = 1
	blockH1      = 2
	blockH2      = 3
	blockH3      = 4
	blockH4      = 5
	blockH5      = 6
	blockH6      = 7
	blockQuote   = 8
	blockCode    = 9
	blockList    = 10
	blockTable   = 11
	blockDivider = 12

	italicStart = "(CLX-ITALIC)"
	italicStop  = "(CLX-ITALIC-STOP)"
)

type block struct {
	Kind int
	Text string
}
