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

	c.Next()
	assert.Equal(t, 0, c.CurrentIndex())
	assert.Equal(t, Top, c.CurrentCategory())
}

func TestPrev_WithoutFavorites(t *testing.T) {
	c := newTestCategories(t, "top,best,ask")

	assert.Equal(t, 0, c.CurrentIndex())

	c.Prev()
	assert.Equal(t, 2, c.CurrentIndex())
	assert.Equal(t, Ask, c.CurrentCategory())

	c.Prev()
	assert.Equal(t, 1, c.CurrentIndex())
	assert.Equal(t, Best, c.CurrentCategory())
}

func TestNext_WithFavorites(t *testing.T) {
	c := newTestCategories(t, "top,best,favorites")

	c.Next()
	assert.Equal(t, Best, c.CurrentCategory())

	c.Next()
	assert.Equal(t, Favorites, c.CurrentCategory())

	c.Next()
	assert.Equal(t, Top, c.CurrentCategory())
}

func TestPrev_WithFavorites(t *testing.T) {
	c := newTestCategories(t, "top,best,favorites")

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
	c := newTestCategories(t, "top,best,favorites")
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

func TestNew_Favorites(t *testing.T) {
	c := newTestCategories(t, "favorites")
	assert.Equal(t, Favorites, c.CurrentCategory())
}

func TestDefault_IncludesFavorites(t *testing.T) {
	c := newTestCategories(t, Default)
	assert.Contains(t, c.ActiveCategories(), Favorites)
}

func TestAvailableNames_IncludesFavorites(t *testing.T) {
	assert.Contains(t, AvailableNames(), "favorites")
}

func TestEndpoint(t *testing.T) {
	assert.Equal(t, "topstories", Endpoint(Top))
	assert.Equal(t, "newstories", Endpoint(Newest))
	assert.Equal(t, "beststories", Endpoint(Best))
	assert.Empty(t, Endpoint(Favorites), "favorites is local and has no endpoint")
}

func TestPolicy(t *testing.T) {
	assert.Equal(t, MultiPage, Policy(Top))
	assert.Equal(t, SinglePage, Policy(Ask))
	assert.Equal(t, SinglePage, Policy(Show))
}

func TestIsFavorites(t *testing.T) {
	assert.True(t, IsFavorites(Favorites))
	assert.False(t, IsFavorites(Top))
	assert.False(t, IsFavorites(Ask))
}

func TestCount_MatchesNamedCategories(t *testing.T) {
	assert.Equal(t, int(Favorites)+1, Count())
}

// TestCategoryTable_Consistent guards future additions: every category must
// have a name, and only favorites (served locally) may omit an endpoint.
func TestCategoryTable_Consistent(t *testing.T) {
	for i := range Count() {
		cat := Category(i)

		assert.NotEmptyf(t, Name(cat), "category %d has no name", i)
		assert.NotEqualf(t, "unknown", Name(cat), "category %d falls through to unknown", i)

		if IsFavorites(cat) {
			assert.Emptyf(t, Endpoint(cat), "favorites is local and must not have an endpoint")
		} else {
			assert.NotEmptyf(t, Endpoint(cat), "fetched category %q must have an endpoint", Name(cat))
		}

		got, ok := categoryFromName(Name(cat))
		assert.Truef(t, ok, "Name/categoryFromName round-trip failed for %q", Name(cat))
		assert.Equal(t, cat, got)
	}
}
