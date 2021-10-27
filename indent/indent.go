package indent

import "os"

const (
	noIndent      = " "
	itermIndent   = "▎"
	regularIndent = "┃"
)

func GetIndentSymbol(hideIndentSymbol bool) string {
	if hideIndentSymbol {
		return noIndent
	}

	terminal := os.Getenv("LC_TERMINAL")

	if terminal == "iTerm2" {
		return itermIndent
	}

	return regularIndent
}
