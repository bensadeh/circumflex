package favorites

import (
	"clx/item"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFavorites_AddAndGetItems(t *testing.T) {
	f := &Favorites{}
	assert.False(t, f.HasItems())
	assert.Empty(t, f.GetItems())

	f.Add(&item.Story{ID: 1, Title: "First"})
	assert.True(t, f.HasItems())
	assert.Len(t, f.GetItems(), 1)

	f.Add(&item.Story{ID: 2, Title: "Second"})
	assert.Len(t, f.GetItems(), 2)
}

func TestFavorites_Remove(t *testing.T) {
	f := &Favorites{}
	f.Add(&item.Story{ID: 1, Title: "First"})
	f.Add(&item.Story{ID: 2, Title: "Second"})
	f.Add(&item.Story{ID: 3, Title: "Third"})

	err := f.Remove(1)
	require.NoError(t, err)
	assert.Len(t, f.GetItems(), 2)
	assert.Equal(t, 1, f.GetItems()[0].ID)
	assert.Equal(t, 3, f.GetItems()[1].ID)
}

func TestFavorites_Remove_First(t *testing.T) {
	f := &Favorites{}
	f.Add(&item.Story{ID: 1, Title: "First"})
	f.Add(&item.Story{ID: 2, Title: "Second"})

	err := f.Remove(0)
	require.NoError(t, err)
	assert.Len(t, f.GetItems(), 1)
	assert.Equal(t, 2, f.GetItems()[0].ID)
}

func TestFavorites_Remove_Last(t *testing.T) {
	f := &Favorites{}
	f.Add(&item.Story{ID: 1, Title: "First"})
	f.Add(&item.Story{ID: 2, Title: "Second"})

	err := f.Remove(1)
	require.NoError(t, err)
	assert.Len(t, f.GetItems(), 1)
	assert.Equal(t, 1, f.GetItems()[0].ID)
}

func TestFavorites_Remove_OutOfBounds(t *testing.T) {
	f := &Favorites{}
	f.Add(&item.Story{ID: 1, Title: "First"})

	err := f.Remove(-1)
	require.Error(t, err)

	err = f.Remove(1)
	require.Error(t, err)

	err = f.Remove(5)
	require.Error(t, err)
}

func TestFavorites_Remove_Empty(t *testing.T) {
	f := &Favorites{}

	err := f.Remove(0)
	assert.Error(t, err)
}

func TestUnmarshal_ValidJSON(t *testing.T) {
	data := []byte(`[{"ID":1,"Title":"Test","Points":100}]`)
	items, err := unmarshal(data)
	require.NoError(t, err)
	assert.Len(t, items, 1)
	assert.Equal(t, 1, items[0].ID)
	assert.Equal(t, "Test", items[0].Title)
}

func TestUnmarshal_InvalidJSON(t *testing.T) {
	data := []byte(`not json`)
	items, err := unmarshal(data)
	require.Error(t, err)
	assert.Nil(t, items)
}

func TestUnmarshal_EmptyArray(t *testing.T) {
	data := []byte(`[]`)
	items, err := unmarshal(data)
	require.NoError(t, err)
	assert.Empty(t, items)
}

func TestFavorites_UpdateStoryAndWriteToDisk(t *testing.T) {
	f := &Favorites{}
	f.Add(&item.Story{ID: 1, Title: "Old Title", Points: 10, CommentsCount: 5})

	newItem := &item.Story{ID: 1, Title: "New Title", Points: 100, CommentsCount: 50}
	_ = f.UpdateStoryAndWriteToDisk(newItem)

	assert.Equal(t, "New Title", f.GetItems()[0].Title)
	assert.Equal(t, 100, f.GetItems()[0].Points)
	assert.Equal(t, 50, f.GetItems()[0].CommentsCount)
}

func TestFavorites_UpdateStoryAndWriteToDisk_NoMatch(t *testing.T) {
	f := &Favorites{}
	f.Add(&item.Story{ID: 1, Title: "Original"})

	newItem := &item.Story{ID: 99, Title: "Different"}
	_ = f.UpdateStoryAndWriteToDisk(newItem)

	assert.Equal(t, "Original", f.GetItems()[0].Title)
}
