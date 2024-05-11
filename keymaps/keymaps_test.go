package keymaps_test

import (
	"testing"

	ansi "github.com/f01c33/circumflex/utils/strip-ansi"

	"github.com/f01c33/circumflex/keymaps"

	"github.com/stretchr/testify/assert"
)

func TestKeymaps(t *testing.T) {
	t.Parallel()

	keys := new(keymaps.List)
	keys.Init()

	keys.AddHeader("Header")
	keys.AddSeparator()
	keys.AddKeymap("Very long description", "x")
	keys.AddKeymap("Separate item", "xyz")
	keys.AddSeparator()
	keys.AddKeymap("Add item", "x")
	keys.AddKeymap("Delete item", "x")
	keys.AddSeparator()
	keys.AddHeader("Header")
	keys.AddSeparator()
	keys.AddKeymap("Delete item", "x")
	keys.AddKeymap("Item", "a + b")

	actual := keys.Print(80)

	expected := `                                     Header                                     

[1mx[0m[2m ........................................................ [0mVery long description
[1mxyz[0m[2m .............................................................. [0mSeparate item

[1mx[0m[2m ..................................................................... [0mAdd item
[1mx[0m[2m .................................................................. [0mDelete item

                                     Header                                     

[1mx[0m[2m .................................................................. [0mDelete item
[1ma + b[0m[2m ..................................................................... [0mItem
`

	// Workaround for a bug where lipgloss does not render ansi formatting during testing
	// Possibly related to https://github.com/charmbracelet/lipgloss/issues/176
	expectedWithoutAnsi := ansi.Strip(expected)

	assert.Equal(t, expectedWithoutAnsi, actual)
}
