package categories

import (
	"fmt"
	"os"
	"strings"
)

type Category string

const (
	FrontPage Category = "frontpage"
	Newest    Category = "newest"
	Ask       Category = "ask"
	Show      Category = "show"
	Favorites Category = "favorites"
)

var CategoryMapping = map[Category]int{
	FrontPage: 0,
	Newest:    1,
	Ask:       2,
	Show:      3,
	Favorites: 4,
}

type Categories struct {
	categories   []int
	currentIndex int
}

func New(categoriesCSV string) *Categories {
	categories := strings.Split(categoriesCSV, ",")
	var validCategories []int
	for _, category := range categories {
		category = strings.TrimSpace(category)
		category = strings.ToLower(category)

		enumCategory := Category(category)
		value, exists := CategoryMapping[enumCategory]

		if !exists || enumCategory == Favorites {
			println(fmt.Sprintf("Unsupported category: %s", category))
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

func (c *Categories) GetCategories(hasFavorites bool) []int {
	if hasFavorites {
		return append(c.categories, CategoryMapping[Favorites])
	}

	return c.categories
}

func (c *Categories) GetCurrentCategory() int {
	return c.currentIndex
}

func (c *Categories) GetNextCategory(hasFavorites bool) int {
	nextIndex := c.currentIndex + 1
	if hasFavorites && nextIndex >= len(c.categories)+1 || !hasFavorites && nextIndex >= len(c.categories) {
		nextIndex = 0
	}
	if hasFavorites && nextIndex == len(c.categories) {
		return CategoryMapping[Favorites]
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
		return CategoryMapping[Favorites]
	}
	return c.categories[prevIndex]
}

func (c *Categories) SetIndex(index int) {
	c.currentIndex = index
}
