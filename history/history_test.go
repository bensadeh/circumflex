package history

import (
	"os"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestPersistent(t *testing.T) *Persistent {
	t.Helper()

	return &Persistent{
		filePath:       path.Join(t.TempDir(), "history.json"),
		VisitedStories: make(map[int]StoryInfo),
	}
}

func TestPersistent_MarkAndContains(t *testing.T) {
	h := newTestPersistent(t)
	assert.False(t, h.Contains(42))

	err := h.MarkAsReadAndWriteToDisk(42, 10)
	require.NoError(t, err)

	assert.True(t, h.Contains(42))
}

func TestPersistent_LastVisited(t *testing.T) {
	h := newTestPersistent(t)

	// Unvisited story returns current time (approximately)
	ts := h.LastVisited(1)
	assert.Positive(t, ts)

	// After marking, returns the marked time
	_ = h.MarkAsReadAndWriteToDisk(1, 5)
	ts2 := h.LastVisited(1)
	assert.Positive(t, ts2)
}

func TestPersistent_ClearAndWriteToDisk(t *testing.T) {
	h := newTestPersistent(t)
	h.VisitedStories[1] = StoryInfo{LastVisited: 100, CommentsOnLastVisit: 5}
	h.VisitedStories[2] = StoryInfo{LastVisited: 200, CommentsOnLastVisit: 10}

	_ = h.ClearAndWriteToDisk()

	assert.Empty(t, h.VisitedStories)
	assert.False(t, h.Contains(1))
	assert.False(t, h.Contains(2))
}

func TestPersistent_WriteToDisk_RoundTrip(t *testing.T) {
	filePath := path.Join(t.TempDir(), "test_history.json")

	h := &Persistent{
		filePath:       filePath,
		VisitedStories: make(map[int]StoryInfo),
	}
	h.VisitedStories[42] = StoryInfo{LastVisited: 1234567890, CommentsOnLastVisit: 15}

	err := writeToDisk(h, filePath)
	require.NoError(t, err)

	// Verify file was created
	_, statErr := os.Stat(filePath)
	require.NoError(t, statErr)

	// Read it back
	content, err := os.ReadFile(filePath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "42")
}

func TestNonPersistent_NoOps(t *testing.T) {
	h := NonPersistent{}

	assert.False(t, h.Contains(1))
	assert.Positive(t, h.LastVisited(1))
	assert.NoError(t, h.ClearAndWriteToDisk())
	assert.NoError(t, h.MarkAsReadAndWriteToDisk(1, 5))
}

func TestPersistent_MarkAsRead_SkipsWithinThreshold(t *testing.T) {
	h := newTestPersistent(t)

	// First mark sets the timestamp
	err := h.MarkAsReadAndWriteToDisk(1, 5)
	require.NoError(t, err)

	firstVisit := h.LastVisited(1)

	// Marking again within 5 minutes should not update the timestamp
	err = h.MarkAsReadAndWriteToDisk(1, 10)
	require.NoError(t, err)

	assert.Equal(t, firstVisit, h.LastVisited(1))
	assert.Equal(t, 5, h.VisitedStories[1].CommentsOnLastVisit)
}

func TestPersistent_MarkAsRead_UpdatesAfterThreshold(t *testing.T) {
	h := newTestPersistent(t)

	// Set a timestamp 6 minutes in the past
	h.VisitedStories[1] = StoryInfo{
		LastVisited:         time.Now().Unix() - 6*60,
		CommentsOnLastVisit: 5,
	}

	err := h.MarkAsReadAndWriteToDisk(1, 15)
	require.NoError(t, err)

	assert.Equal(t, 15, h.VisitedStories[1].CommentsOnLastVisit)
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
