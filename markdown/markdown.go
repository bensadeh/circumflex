package markdown

const (
	Text  = 0
	Image = 1
	H1    = 2
	H2    = 3
	H3    = 4
	H4    = 5
	H5    = 6
	H6    = 7
	Quote = 8
	Code  = 8
)

type Block struct {
	Kind int
	Text string
}
