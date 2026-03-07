package categories

import (
	"clx/category"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew_SingleCategory(t *testing.T) {
	c := New("top")
	assert.Equal(t, category.Top, c.GetCurrentCategory(false))
}

func TestNew_MultipleCategories(t *testing.T) {
	c := New("top,best,ask,show")
	cats := c.GetCategories(false)
	assert.Len(t, cats, 4)
	assert.Equal(t, category.Top, cats[0])
	assert.Equal(t, category.Best, cats[1])
	assert.Equal(t, category.Ask, cats[2])
	assert.Equal(t, category.Show, cats[3])
}

func TestNew_WhitespaceHandling(t *testing.T) {
	c := New(" top , best ")
	cats := c.GetCategories(false)
	assert.Len(t, cats, 2)
	assert.Equal(t, category.Top, cats[0])
	assert.Equal(t, category.Best, cats[1])
}

func TestNew_CaseInsensitive(t *testing.T) {
	c := New("TOP,Best,ASK")
	cats := c.GetCategories(false)
	assert.Len(t, cats, 3)
	assert.Equal(t, category.Top, cats[0])
	assert.Equal(t, category.Best, cats[1])
	assert.Equal(t, category.Ask, cats[2])
}

func TestNext_WithoutFavorites(t *testing.T) {
	c := New("top,best,ask")

	assert.Equal(t, 0, c.GetCurrentIndex())
	assert.Equal(t, category.Top, c.GetCurrentCategory(false))

	c.Next(false)
	assert.Equal(t, 1, c.GetCurrentIndex())
	assert.Equal(t, category.Best, c.GetCurrentCategory(false))

	c.Next(false)
	assert.Equal(t, 2, c.GetCurrentIndex())
	assert.Equal(t, category.Ask, c.GetCurrentCategory(false))

	// Wraps around
	c.Next(false)
	assert.Equal(t, 0, c.GetCurrentIndex())
	assert.Equal(t, category.Top, c.GetCurrentCategory(false))
}

func TestPrev_WithoutFavorites(t *testing.T) {
	c := New("top,best,ask")

	assert.Equal(t, 0, c.GetCurrentIndex())

	// Wraps to last
	c.Prev(false)
	assert.Equal(t, 2, c.GetCurrentIndex())
	assert.Equal(t, category.Ask, c.GetCurrentCategory(false))

	c.Prev(false)
	assert.Equal(t, 1, c.GetCurrentIndex())
	assert.Equal(t, category.Best, c.GetCurrentCategory(false))
}

func TestNext_WithFavorites(t *testing.T) {
	c := New("top,best")

	// With favorites, there are 3 positions: top(0), best(1), favorites(2)
	c.Next(true)
	assert.Equal(t, category.Best, c.GetCurrentCategory(true))

	c.Next(true)
	assert.Equal(t, category.Favorites, c.GetCurrentCategory(true))

	// Wraps around
	c.Next(true)
	assert.Equal(t, category.Top, c.GetCurrentCategory(true))
}

func TestPrev_WithFavorites(t *testing.T) {
	c := New("top,best")

	// Wraps to favorites
	c.Prev(true)
	assert.Equal(t, category.Favorites, c.GetCurrentCategory(true))

	c.Prev(true)
	assert.Equal(t, category.Best, c.GetCurrentCategory(true))
}

func TestGetNextCategory(t *testing.T) {
	c := New("top,best,ask")

	assert.Equal(t, category.Best, c.GetNextCategory(false))

	c.SetIndex(2)
	assert.Equal(t, category.Top, c.GetNextCategory(false))
}

func TestGetPrevCategory(t *testing.T) {
	c := New("top,best,ask")

	assert.Equal(t, category.Ask, c.GetPrevCategory(false))

	c.SetIndex(2)
	assert.Equal(t, category.Best, c.GetPrevCategory(false))
}

func TestGetNextIndex(t *testing.T) {
	c := New("top,best,ask")

	assert.Equal(t, 1, c.GetNextIndex(false))

	c.SetIndex(2)
	assert.Equal(t, 0, c.GetNextIndex(false))
}

func TestGetPrevIndex(t *testing.T) {
	c := New("top,best,ask")

	assert.Equal(t, 2, c.GetPrevIndex(false))

	c.SetIndex(2)
	assert.Equal(t, 1, c.GetPrevIndex(false))
}

func TestGetCategories_WithFavorites(t *testing.T) {
	c := New("top,best")
	cats := c.GetCategories(true)
	assert.Len(t, cats, 3)
	assert.Equal(t, category.Favorites, cats[2])
}

func TestGetCategories_WithoutFavorites(t *testing.T) {
	c := New("top,best")
	cats := c.GetCategories(false)
	assert.Len(t, cats, 2)
}

func TestSetIndex(t *testing.T) {
	c := New("top,best,ask")

	c.SetIndex(2)
	assert.Equal(t, 2, c.GetCurrentIndex())
	assert.Equal(t, category.Ask, c.GetCurrentCategory(false))
}
