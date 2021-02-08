package vim

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetItemDown(t *testing.T) {
	cases := []struct {
		Register    string
		CurrentItem int
		ItemCount   int
		Expected    int
	}{
		{
			"",
			0,
			10,
			1,
		},
		{
			"20",
			0,
			10,
			9,
		},
	}

	for _, c := range cases {
		actual := GetItemDown(c.Register, c.CurrentItem, c.ItemCount)

		assert.Equal(t, c.Expected, actual)
	}
}
