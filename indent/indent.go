package indent

import "os"

const (
	noIndent            = " "
	normalIndent        = "▎"
	compatibilityIndent = "┃"
)

func GetIndentSymbol(hideIndentSymbol bool) string {
	if hideIndentSymbol {
		return noIndent
	}

	if os.Getenv("TERM_PROGRAM") == "Apple_Terminal" {
		return compatibilityIndent
	}

	return normalIndent
}
