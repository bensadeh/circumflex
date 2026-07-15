package categories

import (
	"fmt"
	"strings"
)

type Category int

const (
	Top Category = iota
	Newest
	Ask
	Show
	Best
	Jobs
	Favorites
)

// FetchPolicy describes how many stories to fetch for a category.
type FetchPolicy int

const (
	// MultiPage fetches PerPage * PageMultiplier IDs; used by the large pools
	// (top, new, best) where paging deep is worthwhile.
	MultiPage FetchPolicy = iota
	// SinglePage fetches one page; the Ask, Show and Jobs pools are only
	// ~150-200 IDs, so fetching more would exceed the pool and waste requests.
	SinglePage
)

type info struct {
	name        string
	endpoint    string
	fetchPolicy FetchPolicy
}

// categoryInfo is the single source of truth for per-category facts. Adding a
// category means adding an enum value above and one row here; everything else
// (names, parsing, endpoints, fetch sizing, Count) derives from this table.
var categoryInfo = [...]info{
	Top:    {name: "top", endpoint: "topstories", fetchPolicy: MultiPage},
	Newest: {name: "new", endpoint: "newstories", fetchPolicy: MultiPage},
	Ask:    {name: "ask", endpoint: "askstories", fetchPolicy: SinglePage},
	Show:   {name: "show", endpoint: "showstories", fetchPolicy: SinglePage},
	Best:   {name: "best", endpoint: "beststories", fetchPolicy: MultiPage},
	Jobs:   {name: "jobs", endpoint: "jobstories", fetchPolicy: SinglePage},
	// favorites is served locally; it is never fetched, so fetchPolicy is unused.
	Favorites: {name: "favorites", endpoint: "", fetchPolicy: SinglePage},
}

// Count is the number of defined categories.
func Count() int { return len(categoryInfo) }

func (cat Category) valid() bool { return cat >= 0 && int(cat) < len(categoryInfo) }

func Name(cat Category) string {
	if !cat.valid() {
		return "unknown"
	}

	return categoryInfo[cat].name
}

// Endpoint returns the Firebase endpoint used to fetch cat's stories. It is
// empty for favorites, which is served locally rather than fetched.
func Endpoint(cat Category) string {
	if !cat.valid() {
		return ""
	}

	return categoryInfo[cat].endpoint
}

// Policy returns how many stories to fetch for cat.
func Policy(cat Category) FetchPolicy {
	if !cat.valid() {
		return SinglePage
	}

	return categoryInfo[cat].fetchPolicy
}

// IsFavorites reports whether cat is the favorites view, which is served from
// saved items on disk rather than fetched over the network.
func IsFavorites(cat Category) bool { return cat == Favorites }

// Default is the default value for the --categories flag.
const Default = "top,best,ask,show,favorites"

// AvailableNames returns the names accepted by the --categories flag.
func AvailableNames() []string {
	names := make([]string, len(categoryInfo))
	for i, inf := range categoryInfo {
		names[i] = inf.name
	}

	return names
}

func categoryFromName(name string) (Category, bool) {
	for i, inf := range categoryInfo {
		if inf.name == name {
			return Category(i), true
		}
	}

	return 0, false
}

type Categories struct {
	list         []Category
	currentIndex int
}

func New(categoriesCSV string) (*Categories, error) {
	if categoriesCSV == "" {
		return nil, fmt.Errorf("need at least one category")
	}

	parts := strings.Split(categoriesCSV, ",")

	var validCategories []Category

	for _, part := range parts {
		part = strings.TrimSpace(part)
		part = strings.ToLower(part)

		value, exists := categoryFromName(part)

		if !exists {
			return nil, fmt.Errorf("unsupported category: %q (available: %s)", part, strings.Join(AvailableNames(), ", "))
		}

		validCategories = append(validCategories, value)
	}

	return &Categories{
		list:         validCategories,
		currentIndex: 0,
	}, nil
}

func (c *Categories) Next() {
	c.currentIndex++
	if c.currentIndex >= len(c.list) {
		c.currentIndex = 0
	}
}

func (c *Categories) Prev() {
	c.currentIndex--
	if c.currentIndex < 0 {
		c.currentIndex = len(c.list) - 1
	}
}

func (c *Categories) NextIndex() int {
	nextIndex := c.currentIndex + 1
	if nextIndex >= len(c.list) {
		nextIndex = 0
	}

	return nextIndex
}

func (c *Categories) PrevIndex() int {
	prevIndex := c.currentIndex - 1
	if prevIndex < 0 {
		prevIndex = len(c.list) - 1
	}

	return prevIndex
}

func (c *Categories) ActiveCategories() []Category {
	return c.list
}

func (c *Categories) CurrentCategory() Category {
	return c.list[c.currentIndex]
}

func (c *Categories) CurrentIndex() int {
	return c.currentIndex
}

func (c *Categories) NextCategory() Category {
	return c.list[c.NextIndex()]
}

func (c *Categories) PrevCategory() Category {
	return c.list[c.PrevIndex()]
}

func (c *Categories) SetIndex(index int) {
	if index < 0 || index >= len(c.list) {
		return
	}

	c.currentIndex = index
}
