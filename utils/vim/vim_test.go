package vim

import (
	"strconv"
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
		{
			"g",
			0,
			10,
			1,
		},
	}

	for _, c := range cases {
		actual := GetItemDown(c.Register, c.CurrentItem, c.ItemCount)

		assert.Equal(t, c.Expected, actual)
	}
}

func TestIsNumberWithGAppended(t *testing.T) {
	cases := []struct {
		Input    string
		Expected bool
	}{
		{Input: "23g", Expected: true},
		{Input: "g", Expected: false},
		{Input: "", Expected: false},
	}

	for _, c := range cases {
		actual := IsNumberWithGAppended(c.Input)

		assert.Equal(t, c.Expected, actual, c.Input+" should return "+strconv.FormatBool(c.Expected))
	}
}
