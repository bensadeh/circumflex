package config

import "github.com/charmbracelet/glamour/ansi"

func GetStyleConfig() ansi.StyleConfig {
	return ansi.StyleConfig{
		Document:              ansi.StyleBlock{},
		BlockQuote:            ansi.StyleBlock{},
		Paragraph:             ansi.StyleBlock{},
		List:                  ansi.StyleList{},
		Heading:               ansi.StyleBlock{},
		H1:                    ansi.StyleBlock{},
		H2:                    ansi.StyleBlock{},
		H3:                    ansi.StyleBlock{},
		H4:                    ansi.StyleBlock{},
		H5:                    ansi.StyleBlock{},
		H6:                    ansi.StyleBlock{},
		Text:                  ansi.StylePrimitive{},
		Strikethrough:         ansi.StylePrimitive{},
		Emph:                  ansi.StylePrimitive{},
		Strong:                ansi.StylePrimitive{},
		HorizontalRule:        ansi.StylePrimitive{},
		Item:                  ansi.StylePrimitive{},
		Enumeration:           ansi.StylePrimitive{},
		Task:                  ansi.StyleTask{},
		Link:                  ansi.StylePrimitive{},
		LinkText:              ansi.StylePrimitive{},
		Image:                 ansi.StylePrimitive{},
		ImageText:             ansi.StylePrimitive{},
		Code:                  ansi.StyleBlock{},
		CodeBlock:             ansi.StyleCodeBlock{},
		Table:                 ansi.StyleTable{},
		DefinitionList:        ansi.StyleBlock{},
		DefinitionTerm:        ansi.StylePrimitive{},
		DefinitionDescription: ansi.StylePrimitive{},
		HTMLBlock:             ansi.StyleBlock{},
		HTMLSpan:              ansi.StyleBlock{},
	}
}
