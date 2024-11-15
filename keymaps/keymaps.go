package keymaps

import (
	"strings"

	"charm.land/lipgloss/v2"

	termText "github.com/MichaelMure/go-term-text"
	"github.com/jedib0t/go-pretty/v6/text"
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
	var output strings.Builder

	for _, item := range k.keymaps {
		switch item.category {
		case header:
			centeredHeader := text.AlignCenter.Apply(item.header, screenWidth)
			output.WriteString(centeredHeader + newline)
		case separator:
			output.WriteString(newline)
		case keymap:
			dots := getDotSeparators(item.description, item.key, screenWidth)
			output.WriteString(lipgloss.NewStyle().Bold(true).Render(item.key) + lipgloss.NewStyle().Faint(true).Render(dots) + item.description + newline)
		}
	}

	return output.String()
}

func getDotSeparators(description string, key string, screenWidth int) string {
	descriptionLength := termText.Len(description)
	keyLength := termText.Len(key)
	space := " "
	spaceLength := termText.Len(space)
	numberOfDotSeparators := screenWidth - descriptionLength - keyLength - spaceLength - spaceLength

	if numberOfDotSeparators < 0 {
		return ""
	}

	return space + strings.Repeat(".", numberOfDotSeparators) + space
}
