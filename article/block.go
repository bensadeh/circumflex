package article

type blockKind int

const (
	blockText blockKind = iota
	blockImage
	blockH1
	blockH2
	blockH3
	blockH4
	blockH5
	blockH6
	blockQuote
	blockCode
	blockList
	blockTable
	blockDivider

	italicStart = "(CLX-ITALIC)"
	italicStop  = "(CLX-ITALIC-STOP)"
)

type block struct {
	Kind blockKind
	Text string
}
