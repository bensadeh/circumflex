package keymaps

import (
	"clx/constants/margins"
	"clx/utils/formatter"
	"strings"

	text "github.com/MichaelMure/go-term-text"
)

const (
	header    = 0
	separator = 1
	keymap    = 2

	newline = "\n"
)

type List struct {
	keymaps []*entry
}

type entry struct {
	header      string
	description string
	key         string
	category    int
}

func (k *List) Init() {
	var entries []*entry
	k.keymaps = entries
}

func (k *List) AddHeader(text string) {
	item := new(entry)
	item.header = text
	item.category = header

	k.keymaps = append(k.keymaps, item)
}

func (k *List) AddSeparator() {
	item := new(entry)
	item.category = separator

	k.keymaps = append(k.keymaps, item)
}

func (k *List) AddKeymap(description string, key string) {
	item := new(entry)
	item.description = description
	item.key = key
	item.category = keymap

	k.keymaps = append(k.keymaps, item)
}

func (k *List) Print(screenWidth int) string {
	output := ""

	for _, item := range k.keymaps {
		switch item.category {
		case header:
			padding := k.getLongestLineLength(screenWidth)/2 - len(item.header) + len(item.header)/2
			padToCenterAlign := strings.Repeat(" ", padding)

			output += padToCenterAlign + formatter.Bold(item.header) + newline
		case separator:
			output += newline
		case keymap:
			dots := getDotSeparators(item.description, item.key, screenWidth-margins.LeftMargin*8)
			output += item.description + dots + item.key + newline
		}
	}

	pad := strings.Repeat(" ", margins.LeftMargin*3)
	output, _ = text.WrapWithPad(output, screenWidth, pad)

	return output
}

func getDotSeparators(description string, key string, width int) string {
	descriptionLength := len(description)
	keyLength := len(key)
	space := " "
	spaceLength := len(space)
	numberOfDotSeparators := width - descriptionLength - keyLength - spaceLength - spaceLength

	if numberOfDotSeparators < 0 {
		return ""
	}

	return space + strings.Repeat(".", numberOfDotSeparators) + space
}

func (k *List) getLongestLineLength(screenWidth int) int {
	allLines := ""

	for _, item := range k.keymaps {
		if item.category == keymap {
			dots := getDotSeparators(item.description, item.key, screenWidth-margins.LeftMargin*8)
			allLines += item.description + dots + item.key + newline
		}
	}

	return text.MaxLineLen(allLines)
}
