package keymaps

import "strings"

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

func (k *List) Print(headerMargin int, width int) string {
	output := ""
	headerIndentation := strings.Repeat(" ", headerMargin)

	for _, item := range k.keymaps {
		switch item.category {
		case header:
			output += headerIndentation + item.header + newline
		case separator:
			output += newline
		case keymap:
			dots := getDotSeparators(item.description, item.key, width)
			output += item.description + dots + item.key + newline
		}
	}

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
