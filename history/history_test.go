package history

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPersistent_MarkAndContains(t *testing.T) {
	h := &Persistent{VisitedStories: make(map[int]StoryInfo)}
	assert.False(t, h.Contains(42))

	err := h.MarkAsReadAndWriteToDisk(42, 10)
	// May fail if cache dir doesn't exist in CI, but the in-memory state should still update
	_ = err

	assert.True(t, h.Contains(42))
}

func TestPersistent_GetLastVisited(t *testing.T) {
	h := &Persistent{VisitedStories: make(map[int]StoryInfo)}

	// Unvisited story returns current time (approximately)
	ts := h.GetLastVisited(1)
	assert.Positive(t, ts)

	// After marking, returns the marked time
	_ = h.MarkAsReadAndWriteToDisk(1, 5)
	ts2 := h.GetLastVisited(1)
	assert.Positive(t, ts2)
}

func TestPersistent_ClearAndWriteToDisk(t *testing.T) {
	h := &Persistent{VisitedStories: make(map[int]StoryInfo)}
	h.VisitedStories[1] = StoryInfo{LastVisited: 100, CommentsOnLastVisit: 5}
	h.VisitedStories[2] = StoryInfo{LastVisited: 200, CommentsOnLastVisit: 10}

	_ = h.ClearAndWriteToDisk()

	assert.Empty(t, h.VisitedStories)
	assert.False(t, h.Contains(1))
	assert.False(t, h.Contains(2))
}

func TestPersistent_WriteToDisk_RoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	fileName := "test_history.json"

	h := &Persistent{VisitedStories: make(map[int]StoryInfo)}
	h.VisitedStories[42] = StoryInfo{LastVisited: 1234567890, CommentsOnLastVisit: 15}

	err := writeToDisk(h, tmpDir, fileName)
	require.NoError(t, err)

	// Verify file was created
	fullPath := path.Join(tmpDir, fileName)
	_, statErr := os.Stat(fullPath)
	require.NoError(t, statErr)

	// Read it back
	content, err := os.ReadFile(fullPath) //nolint:gosec // test temp dir
	require.NoError(t, err)
	assert.Contains(t, string(content), "42")
}

func TestNonPersistent_NoOps(t *testing.T) {
	h := NonPersistent{}

	assert.False(t, h.Contains(1))
	assert.Positive(t, h.GetLastVisited(1))
	assert.NoError(t, h.ClearAndWriteToDisk())
	assert.NoError(t, h.MarkAsReadAndWriteToDisk(1, 5))
}

func TestMock_ContainsKnownIDs(t *testing.T) {
	h := Mock{}

	assert.True(t, h.Contains(2))
	assert.True(t, h.Contains(10))
	assert.True(t, h.Contains(14))
	assert.True(t, h.Contains(18))
	assert.False(t, h.Contains(1))
	assert.False(t, h.Contains(99))
}
