package categories

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestCategories(t *testing.T, csv string) *Categories {
	t.Helper()
	c, err := New(csv)
	require.NoError(t, err)
	return c
}

func TestNew_SingleCategory(t *testing.T) {
	c := newTestCategories(t, "top")
	assert.Equal(t, Top, c.GetCurrentCategory(false))
}

func TestNew_MultipleCategories(t *testing.T) {
	c := newTestCategories(t, "top,best,ask,show")
	cats := c.GetCategories(false)
	assert.Len(t, cats, 4)
	assert.Equal(t, Top, cats[0])
	assert.Equal(t, Best, cats[1])
	assert.Equal(t, Ask, cats[2])
	assert.Equal(t, Show, cats[3])
}

func TestNew_WhitespaceHandling(t *testing.T) {
	c := newTestCategories(t, " top , best ")
	cats := c.GetCategories(false)
	assert.Len(t, cats, 2)
	assert.Equal(t, Top, cats[0])
	assert.Equal(t, Best, cats[1])
}

func TestNew_CaseInsensitive(t *testing.T) {
	c := newTestCategories(t, "TOP,Best,ASK")
	cats := c.GetCategories(false)
	assert.Len(t, cats, 3)
	assert.Equal(t, Top, cats[0])
	assert.Equal(t, Best, cats[1])
	assert.Equal(t, Ask, cats[2])
}

func TestNext_WithoutFavorites(t *testing.T) {
	c := newTestCategories(t, "top,best,ask")

	assert.Equal(t, 0, c.GetCurrentIndex())
	assert.Equal(t, Top, c.GetCurrentCategory(false))

	c.Next(false)
	assert.Equal(t, 1, c.GetCurrentIndex())
	assert.Equal(t, Best, c.GetCurrentCategory(false))

	c.Next(false)
	assert.Equal(t, 2, c.GetCurrentIndex())
	assert.Equal(t, Ask, c.GetCurrentCategory(false))

	// Wraps around
	c.Next(false)
	assert.Equal(t, 0, c.GetCurrentIndex())
	assert.Equal(t, Top, c.GetCurrentCategory(false))
}

func TestPrev_WithoutFavorites(t *testing.T) {
	c := newTestCategories(t, "top,best,ask")

	assert.Equal(t, 0, c.GetCurrentIndex())

	// Wraps to last
	c.Prev(false)
	assert.Equal(t, 2, c.GetCurrentIndex())
	assert.Equal(t, Ask, c.GetCurrentCategory(false))

	c.Prev(false)
	assert.Equal(t, 1, c.GetCurrentIndex())
	assert.Equal(t, Best, c.GetCurrentCategory(false))
}

func TestNext_WithFavorites(t *testing.T) {
	c := newTestCategories(t, "top,best")

	// With favorites, there are 3 positions: top(0), best(1), favorites(2)
	c.Next(true)
	assert.Equal(t, Best, c.GetCurrentCategory(true))

	c.Next(true)
	assert.Equal(t, Favorites, c.GetCurrentCategory(true))

	// Wraps around
	c.Next(true)
	assert.Equal(t, Top, c.GetCurrentCategory(true))
}

func TestPrev_WithFavorites(t *testing.T) {
	c := newTestCategories(t, "top,best")

	// Wraps to favorites
	c.Prev(true)
	assert.Equal(t, Favorites, c.GetCurrentCategory(true))

	c.Prev(true)
	assert.Equal(t, Best, c.GetCurrentCategory(true))
}

func TestGetNextCategory(t *testing.T) {
	c := newTestCategories(t, "top,best,ask")

	assert.Equal(t, Best, c.GetNextCategory(false))

	c.SetIndex(2)
	assert.Equal(t, Top, c.GetNextCategory(false))
}

func TestGetPrevCategory(t *testing.T) {
	c := newTestCategories(t, "top,best,ask")

	assert.Equal(t, Ask, c.GetPrevCategory(false))

	c.SetIndex(2)
	assert.Equal(t, Best, c.GetPrevCategory(false))
}

func TestGetNextIndex(t *testing.T) {
	c := newTestCategories(t, "top,best,ask")

	assert.Equal(t, 1, c.GetNextIndex(false))

	c.SetIndex(2)
	assert.Equal(t, 0, c.GetNextIndex(false))
}

func TestGetPrevIndex(t *testing.T) {
	c := newTestCategories(t, "top,best,ask")

	assert.Equal(t, 2, c.GetPrevIndex(false))

	c.SetIndex(2)
	assert.Equal(t, 1, c.GetPrevIndex(false))
}

func TestGetCategories_WithFavorites(t *testing.T) {
	c := newTestCategories(t, "top,best")
	cats := c.GetCategories(true)
	assert.Len(t, cats, 3)
	assert.Equal(t, Favorites, cats[2])
}

func TestGetCategories_WithoutFavorites(t *testing.T) {
	c := newTestCategories(t, "top,best")
	cats := c.GetCategories(false)
	assert.Len(t, cats, 2)
}

func TestSetIndex(t *testing.T) {
	c := newTestCategories(t, "top,best,ask")

	c.SetIndex(2)
	assert.Equal(t, 2, c.GetCurrentIndex())
	assert.Equal(t, Ask, c.GetCurrentCategory(false))
}

func TestNew_EmptyString(t *testing.T) {
	_, err := New("")
	assert.Error(t, err)
}

func TestNew_InvalidCategory(t *testing.T) {
	_, err := New("top,invalid")
	assert.Error(t, err)
}
