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
	assert.Equal(t, Top, c.CurrentCategory())
}

func TestNew_MultipleCategories(t *testing.T) {
	c := newTestCategories(t, "top,best,ask,show")
	cats := c.ActiveCategories()
	assert.Len(t, cats, 4)
	assert.Equal(t, Top, cats[0])
	assert.Equal(t, Best, cats[1])
	assert.Equal(t, Ask, cats[2])
	assert.Equal(t, Show, cats[3])
}

func TestNew_WhitespaceHandling(t *testing.T) {
	c := newTestCategories(t, " top , best ")
	cats := c.ActiveCategories()
	assert.Len(t, cats, 2)
	assert.Equal(t, Top, cats[0])
	assert.Equal(t, Best, cats[1])
}

func TestNew_CaseInsensitive(t *testing.T) {
	c := newTestCategories(t, "TOP,Best,ASK")
	cats := c.ActiveCategories()
	assert.Len(t, cats, 3)
	assert.Equal(t, Top, cats[0])
	assert.Equal(t, Best, cats[1])
	assert.Equal(t, Ask, cats[2])
}

func TestNext_WithoutFavorites(t *testing.T) {
	c := newTestCategories(t, "top,best,ask")

	assert.Equal(t, 0, c.CurrentIndex())
	assert.Equal(t, Top, c.CurrentCategory())

	c.Next()
	assert.Equal(t, 1, c.CurrentIndex())
	assert.Equal(t, Best, c.CurrentCategory())

	c.Next()
	assert.Equal(t, 2, c.CurrentIndex())
	assert.Equal(t, Ask, c.CurrentCategory())

	// Wraps around
	c.Next()
	assert.Equal(t, 0, c.CurrentIndex())
	assert.Equal(t, Top, c.CurrentCategory())
}

func TestPrev_WithoutFavorites(t *testing.T) {
	c := newTestCategories(t, "top,best,ask")

	assert.Equal(t, 0, c.CurrentIndex())

	// Wraps to last
	c.Prev()
	assert.Equal(t, 2, c.CurrentIndex())
	assert.Equal(t, Ask, c.CurrentCategory())

	c.Prev()
	assert.Equal(t, 1, c.CurrentIndex())
	assert.Equal(t, Best, c.CurrentCategory())
}

func TestNext_WithFavorites(t *testing.T) {
	c := newTestCategories(t, "top,best")
	c.SetFavorites(true)

	// With favorites, there are 3 positions: top(0), best(1), favorites(2)
	c.Next()
	assert.Equal(t, Best, c.CurrentCategory())

	c.Next()
	assert.Equal(t, Favorites, c.CurrentCategory())

	// Wraps around
	c.Next()
	assert.Equal(t, Top, c.CurrentCategory())
}

func TestPrev_WithFavorites(t *testing.T) {
	c := newTestCategories(t, "top,best")
	c.SetFavorites(true)

	// Wraps to favorites
	c.Prev()
	assert.Equal(t, Favorites, c.CurrentCategory())

	c.Prev()
	assert.Equal(t, Best, c.CurrentCategory())
}

func TestNextCategory(t *testing.T) {
	c := newTestCategories(t, "top,best,ask")

	assert.Equal(t, Best, c.NextCategory())

	c.SetIndex(2)
	assert.Equal(t, Top, c.NextCategory())
}

func TestPrevCategory(t *testing.T) {
	c := newTestCategories(t, "top,best,ask")

	assert.Equal(t, Ask, c.PrevCategory())

	c.SetIndex(2)
	assert.Equal(t, Best, c.PrevCategory())
}

func TestNextIndex(t *testing.T) {
	c := newTestCategories(t, "top,best,ask")

	assert.Equal(t, 1, c.NextIndex())

	c.SetIndex(2)
	assert.Equal(t, 0, c.NextIndex())
}

func TestPrevIndex(t *testing.T) {
	c := newTestCategories(t, "top,best,ask")

	assert.Equal(t, 2, c.PrevIndex())

	c.SetIndex(2)
	assert.Equal(t, 1, c.PrevIndex())
}

func TestActiveCategories_WithFavorites(t *testing.T) {
	c := newTestCategories(t, "top,best")
	c.SetFavorites(true)
	cats := c.ActiveCategories()
	assert.Len(t, cats, 3)
	assert.Equal(t, Favorites, cats[2])
}

func TestActiveCategories_WithoutFavorites(t *testing.T) {
	c := newTestCategories(t, "top,best")
	cats := c.ActiveCategories()
	assert.Len(t, cats, 2)
}

func TestSetIndex(t *testing.T) {
	c := newTestCategories(t, "top,best,ask")

	c.SetIndex(2)
	assert.Equal(t, 2, c.CurrentIndex())
	assert.Equal(t, Ask, c.CurrentCategory())
}

func TestSetIndex_OutOfBounds(t *testing.T) {
	c := newTestCategories(t, "top,best,ask")

	c.SetIndex(1)
	c.SetIndex(99)
	assert.Equal(t, 1, c.CurrentIndex(), "should be no-op for out-of-range index")

	c.SetIndex(-1)
	assert.Equal(t, 1, c.CurrentIndex(), "should be no-op for negative index")
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
	assert.Len(t, c.ActiveCategories(), 3)
	assert.True(t, c.HasFavorites())

	c.SetFavorites(false)
	c.SetFavorites(false)
	assert.Len(t, c.ActiveCategories(), 2)
	assert.False(t, c.HasFavorites())
}

func TestSetFavorites_ClampIndex(t *testing.T) {
	c := newTestCategories(t, "top,best")
	c.SetFavorites(true)

	// Move to favorites (index 2)
	c.SetIndex(2)
	assert.Equal(t, Favorites, c.CurrentCategory())

	// Remove favorites — index should clamp to last valid (1)
	c.SetFavorites(false)
	assert.Equal(t, 1, c.CurrentIndex())
	assert.Equal(t, Best, c.CurrentCategory())
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
