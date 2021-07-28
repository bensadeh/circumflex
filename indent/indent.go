package indent

func GetIndentSymbol(useAlternateIndent bool) string {
	if useAlternateIndent {
		return "┃"
	}

	return "▎"
}
