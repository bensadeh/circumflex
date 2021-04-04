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

	actual := keys.Print(5, 20)

	expected := `     Header

Very long descriptionx
Separate item .. xyz

Add item ......... x
Delete item ...... x

     Header

Delete item ...... x
Item ......... a + b
`

	assert.Equal(t, expected, actual)
}
