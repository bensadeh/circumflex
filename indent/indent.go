package indent

const (
	noIndent     = " "
	altIndent    = "┃"
	normalIndent = "▎"
)

func GetIndentSymbol(hideIndentSymbol bool, useAlternateIndent bool) string {
	if hideIndentSymbol {
		return noIndent
	}

	if useAlternateIndent {
		return altIndent
	}

	return normalIndent
}
