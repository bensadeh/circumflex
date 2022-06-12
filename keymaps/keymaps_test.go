package keymaps_test

import (
	"clx/keymaps"
	"testing"

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

	expected := `[1m                                     Header                                     [0m

[1mx[0m[2m ........................................................ [0mVery long description
[1mxyz[0m[2m .............................................................. [0mSeparate item

[1mx[0m[2m ..................................................................... [0mAdd item
[1mx[0m[2m .................................................................. [0mDelete item

[1m                                     Header                                     [0m

[1mx[0m[2m .................................................................. [0mDelete item
[1ma + b[0m[2m ..................................................................... [0mItem
`

	assert.Equal(t, expected, actual)
}
