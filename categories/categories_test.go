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
	assert.Equal(t, Top, c.GetCurrentCategory())
}

func TestNew_MultipleCategories(t *testing.T) {
	c := newTestCategories(t, "top,best,ask,show")
	cats := c.GetCategories()
	assert.Len(t, cats, 4)
	assert.Equal(t, Top, cats[0])
	assert.Equal(t, Best, cats[1])
	assert.Equal(t, Ask, cats[2])
	assert.Equal(t, Show, cats[3])
}

func TestNew_WhitespaceHandling(t *testing.T) {
	c := newTestCategories(t, " top , best ")
	cats := c.GetCategories()
	assert.Len(t, cats, 2)
	assert.Equal(t, Top, cats[0])
	assert.Equal(t, Best, cats[1])
}

func TestNew_CaseInsensitive(t *testing.T) {
	c := newTestCategories(t, "TOP,Best,ASK")
	cats := c.GetCategories()
	assert.Len(t, cats, 3)
	assert.Equal(t, Top, cats[0])
	assert.Equal(t, Best, cats[1])
	assert.Equal(t, Ask, cats[2])
}

func TestNext_WithoutFavorites(t *testing.T) {
	c := newTestCategories(t, "top,best,ask")

	assert.Equal(t, 0, c.GetCurrentIndex())
	assert.Equal(t, Top, c.GetCurrentCategory())

	c.Next()
	assert.Equal(t, 1, c.GetCurrentIndex())
	assert.Equal(t, Best, c.GetCurrentCategory())

	c.Next()
	assert.Equal(t, 2, c.GetCurrentIndex())
	assert.Equal(t, Ask, c.GetCurrentCategory())

	// Wraps around
	c.Next()
	assert.Equal(t, 0, c.GetCurrentIndex())
	assert.Equal(t, Top, c.GetCurrentCategory())
}

func TestPrev_WithoutFavorites(t *testing.T) {
	c := newTestCategories(t, "top,best,ask")

	assert.Equal(t, 0, c.GetCurrentIndex())

	// Wraps to last
	c.Prev()
	assert.Equal(t, 2, c.GetCurrentIndex())
	assert.Equal(t, Ask, c.GetCurrentCategory())

	c.Prev()
	assert.Equal(t, 1, c.GetCurrentIndex())
	assert.Equal(t, Best, c.GetCurrentCategory())
}

func TestNext_WithFavorites(t *testing.T) {
	c := newTestCategories(t, "top,best")
	c.SetFavorites(true)

	// With favorites, there are 3 positions: top(0), best(1), favorites(2)
	c.Next()
	assert.Equal(t, Best, c.GetCurrentCategory())

	c.Next()
	assert.Equal(t, Favorites, c.GetCurrentCategory())

	// Wraps around
	c.Next()
	assert.Equal(t, Top, c.GetCurrentCategory())
}

func TestPrev_WithFavorites(t *testing.T) {
	c := newTestCategories(t, "top,best")
	c.SetFavorites(true)

	// Wraps to favorites
	c.Prev()
	assert.Equal(t, Favorites, c.GetCurrentCategory())

	c.Prev()
	assert.Equal(t, Best, c.GetCurrentCategory())
}

func TestGetNextCategory(t *testing.T) {
	c := newTestCategories(t, "top,best,ask")

	assert.Equal(t, Best, c.GetNextCategory())

	c.SetIndex(2)
	assert.Equal(t, Top, c.GetNextCategory())
}

func TestGetPrevCategory(t *testing.T) {
	c := newTestCategories(t, "top,best,ask")

	assert.Equal(t, Ask, c.GetPrevCategory())

	c.SetIndex(2)
	assert.Equal(t, Best, c.GetPrevCategory())
}

func TestGetNextIndex(t *testing.T) {
	c := newTestCategories(t, "top,best,ask")

	assert.Equal(t, 1, c.GetNextIndex())

	c.SetIndex(2)
	assert.Equal(t, 0, c.GetNextIndex())
}

func TestGetPrevIndex(t *testing.T) {
	c := newTestCategories(t, "top,best,ask")

	assert.Equal(t, 2, c.GetPrevIndex())

	c.SetIndex(2)
	assert.Equal(t, 1, c.GetPrevIndex())
}

func TestGetCategories_WithFavorites(t *testing.T) {
	c := newTestCategories(t, "top,best")
	c.SetFavorites(true)
	cats := c.GetCategories()
	assert.Len(t, cats, 3)
	assert.Equal(t, Favorites, cats[2])
}

func TestGetCategories_WithoutFavorites(t *testing.T) {
	c := newTestCategories(t, "top,best")
	cats := c.GetCategories()
	assert.Len(t, cats, 2)
}

func TestSetIndex(t *testing.T) {
	c := newTestCategories(t, "top,best,ask")

	c.SetIndex(2)
	assert.Equal(t, 2, c.GetCurrentIndex())
	assert.Equal(t, Ask, c.GetCurrentCategory())
}

func TestSetIndex_OutOfBounds(t *testing.T) {
	c := newTestCategories(t, "top,best,ask")

	c.SetIndex(1)
	c.SetIndex(99)
	assert.Equal(t, 1, c.GetCurrentIndex(), "should be no-op for out-of-range index")

	c.SetIndex(-1)
	assert.Equal(t, 1, c.GetCurrentIndex(), "should be no-op for negative index")
}

func TestNew_EmptyString(t *testing.T) {
	_, err := New("")
	assert.Error(t, err)
}

func TestNew_InvalidCategory(t *testing.T) {
	_, err := New("top,invalid")
	assert.Error(t, err)
}

func TestSetFavorites_Idempotent(t *testing.T) {
	c := newTestCategories(t, "top,best")

	c.SetFavorites(true)
	c.SetFavorites(true)
	assert.Len(t, c.GetCategories(), 3)
	assert.True(t, c.HasFavorites())

	c.SetFavorites(false)
	c.SetFavorites(false)
	assert.Len(t, c.GetCategories(), 2)
	assert.False(t, c.HasFavorites())
}

func TestSetFavorites_ClampIndex(t *testing.T) {
	c := newTestCategories(t, "top,best")
	c.SetFavorites(true)

	// Move to favorites (index 2)
	c.SetIndex(2)
	assert.Equal(t, Favorites, c.GetCurrentCategory())

	// Remove favorites — index should clamp to last valid (1)
	c.SetFavorites(false)
	assert.Equal(t, 1, c.GetCurrentIndex())
	assert.Equal(t, Best, c.GetCurrentCategory())
}

func TestBase(t *testing.T) {
	c := newTestCategories(t, "top,best,ask")
	c.SetFavorites(true)

	base := c.Base()
	assert.Len(t, base, 3)
	assert.Equal(t, Top, base[0])
	assert.Equal(t, Best, base[1])
	assert.Equal(t, Ask, base[2])
}

func TestHasFavorites(t *testing.T) {
	c := newTestCategories(t, "top,best")

	assert.False(t, c.HasFavorites())

	c.SetFavorites(true)
	assert.True(t, c.HasFavorites())

	c.SetFavorites(false)
	assert.False(t, c.HasFavorites())
}
