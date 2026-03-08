package categories

import (
	"fmt"
	"os"
	"strings"
)

const (
	Top       = 0
	Newest    = 1
	Ask       = 2
	Show      = 3
	Best      = 4
	Favorites = 5
	Buffer    = 6
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
	categories   []int
	currentIndex int
}

func New(categoriesCSV string) *Categories {
	if categoriesCSV == "" {
		println("Need at least one category")
		os.Exit(1)
	}
	parts := strings.Split(categoriesCSV, ",")
	var validCategories []int
	for _, part := range parts {
		part = strings.TrimSpace(part)
		part = strings.ToLower(part)

		value, exists := categoryFromName(part)

		if !exists {
			println(fmt.Sprintf("Unsupported category: %s", part))
			os.Exit(1)
		}

		validCategories = append(validCategories, value)
	}

	return &Categories{
		categories:   validCategories,
		currentIndex: 0,
	}
}

func (c *Categories) Next(hasFavorites bool) {
	c.currentIndex++
	if hasFavorites && c.currentIndex >= len(c.categories)+1 || !hasFavorites && c.currentIndex >= len(c.categories) {
		c.currentIndex = 0
	}
}

func (c *Categories) Prev(hasFavorites bool) {
	c.currentIndex--
	if c.currentIndex < 0 {
		if hasFavorites {
			c.currentIndex = len(c.categories)
		} else {
			c.currentIndex = len(c.categories) - 1
		}
	}
}

func (c *Categories) GetNextIndex(hasFavorites bool) int {
	nextIndex := c.currentIndex + 1
	if hasFavorites && nextIndex >= len(c.categories)+1 || !hasFavorites && nextIndex >= len(c.categories) {
		nextIndex = 0
	}

	return nextIndex
}

func (c *Categories) GetPrevIndex(hasFavorites bool) int {
	prevIndex := c.currentIndex - 1
	if prevIndex < 0 {
		if hasFavorites {
			prevIndex = len(c.categories)
		} else {
			prevIndex = len(c.categories) - 1
		}
	}

	return prevIndex
}

func (c *Categories) GetCategories(hasFavorites bool) []int {
	if hasFavorites {
		categoriesWithFavorites := make([]int, len(c.categories), len(c.categories)+1)
		copy(categoriesWithFavorites, c.categories)
		return append(categoriesWithFavorites, Favorites)
	}

	return c.categories
}

func (c *Categories) GetCurrentCategory(hasFavorites bool) int {
	if hasFavorites {
		categoriesWithFavorites := make([]int, len(c.categories), len(c.categories)+1)
		copy(categoriesWithFavorites, c.categories)
		categoriesWithFavorites = append(categoriesWithFavorites, Favorites)
		return categoriesWithFavorites[c.currentIndex]
	}

	return c.categories[c.currentIndex]
}

func (c *Categories) GetCurrentIndex() int {
	return c.currentIndex
}

func (c *Categories) GetNextCategory(hasFavorites bool) int {
	nextIndex := c.currentIndex + 1
	if hasFavorites && nextIndex >= len(c.categories)+1 || !hasFavorites && nextIndex >= len(c.categories) {
		nextIndex = 0
	}

	if hasFavorites && nextIndex == len(c.categories) {
		return Favorites
	}

	return c.categories[nextIndex]
}

func (c *Categories) GetPrevCategory(hasFavorites bool) int {
	prevIndex := c.currentIndex - 1
	if prevIndex < 0 {
		if hasFavorites {
			prevIndex = len(c.categories)
		} else {
			prevIndex = len(c.categories) - 1
		}
	}
	if hasFavorites && prevIndex == len(c.categories) {
		return Favorites
	}
	return c.categories[prevIndex]
}

func (c *Categories) SetIndex(index int) {
	c.currentIndex = index
}
