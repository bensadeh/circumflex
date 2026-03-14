package categories

import (
	"fmt"
	"slices"
	"strings"
)

const (
	Top       = 0
	Newest    = 1
	Ask       = 2
	Show      = 3
	Best      = 4
	Favorites = 5
)

// Name returns the display name for a category constant.
func Name(cat int) string {
	switch cat {
	case Top:
		return "top"
	case Newest:
		return "new"
	case Ask:
		return "ask"
	case Show:
		return "show"
	case Best:
		return "best"
	case Favorites:
		return "favorites"
	default:
		return "unknown"
	}
}

func categoryFromName(name string) (int, bool) {
	switch name {
	case "top":
		return Top, true
	case "new":
		return Newest, true
	case "ask":
		return Ask, true
	case "show":
		return Show, true
	case "best":
		return Best, true
	default:
		return 0, false
	}
}

type Categories struct {
	base         []int
	active       []int
	currentIndex int
}

func New(categoriesCSV string) (*Categories, error) {
	if categoriesCSV == "" {
		return nil, fmt.Errorf("need at least one category")
	}

	parts := strings.Split(categoriesCSV, ",")

	var validCategories []int

	for _, part := range parts {
		part = strings.TrimSpace(part)
		part = strings.ToLower(part)

		value, exists := categoryFromName(part)

		if !exists {
			return nil, fmt.Errorf("unsupported category: %s", part)
		}

		validCategories = append(validCategories, value)
	}

	return &Categories{
		base:         validCategories,
		active:       validCategories,
		currentIndex: 0,
	}, nil
}

func (c *Categories) SetFavorites(has bool) {
	if has {
		c.active = append(slices.Clone(c.base), Favorites)
	} else {
		c.active = c.base
		if c.currentIndex >= len(c.active) {
			c.currentIndex = len(c.active) - 1
		}
	}
}

func (c *Categories) HasFavorites() bool {
	return slices.Contains(c.active, Favorites)
}

func (c *Categories) Base() []int {
	return c.base
}

func (c *Categories) Next() {
	c.currentIndex++
	if c.currentIndex >= len(c.active) {
		c.currentIndex = 0
	}
}

func (c *Categories) Prev() {
	c.currentIndex--
	if c.currentIndex < 0 {
		c.currentIndex = len(c.active) - 1
	}
}

func (c *Categories) NextIndex() int {
	nextIndex := c.currentIndex + 1
	if nextIndex >= len(c.active) {
		nextIndex = 0
	}

	return nextIndex
}

func (c *Categories) PrevIndex() int {
	prevIndex := c.currentIndex - 1
	if prevIndex < 0 {
		prevIndex = len(c.active) - 1
	}

	return prevIndex
}

func (c *Categories) ActiveCategories() []int {
	return c.active
}

func (c *Categories) CurrentCategory() int {
	return c.active[c.currentIndex]
}

func (c *Categories) CurrentIndex() int {
	return c.currentIndex
}

func (c *Categories) NextCategory() int {
	return c.active[c.NextIndex()]
}

func (c *Categories) PrevCategory() int {
	return c.active[c.PrevIndex()]
}

func (c *Categories) SetIndex(index int) {
	if index < 0 || index >= len(c.active) {
		return
	}

	c.currentIndex = index
}
