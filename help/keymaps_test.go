package help

import (
	"testing"

	"github.com/bensadeh/circumflex/ansi"

	"github.com/stretchr/testify/assert"
)

func TestKeymaps(t *testing.T) {
	t.Parallel()

	keys := new(keyList)

	keys.addHeader("Header")
	keys.addSeparator()
	keys.addKeymap("Very long description", "x")
	keys.addKeymap("Separate item", "xyz")
	keys.addSeparator()
	keys.addKeymap("Add item", "x")
	keys.addKeymap("Delete item", "x")
	keys.addSeparator()
	keys.addHeader("Header")
	keys.addSeparator()
	keys.addKeymap("Delete item", "x")
	keys.addKeymap("Item", "a + b")

	actual := keys.print(80)

	expected := `Header

[1mx[0m[2m ........................................................ [0mVery long description
[1mxyz[0m[2m .............................................................. [0mSeparate item

[1mx[0m[2m ..................................................................... [0mAdd item
[1mx[0m[2m .................................................................. [0mDelete item

Header

[1mx[0m[2m .................................................................. [0mDelete item
[1ma + b[0m[2m ..................................................................... [0mItem
`

	assert.Equal(t, ansi.Strip(expected), ansi.Strip(actual))
}
