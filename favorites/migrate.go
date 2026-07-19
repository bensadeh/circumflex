package favorites

import (
	"encoding/json"
	"fmt"
	"os"
)

// legacyItem decodes the pre-5.0 favorites.json format, which used
// PascalCase keys and "User" for the author.
type legacyItem struct {
	ID            int    `json:"ID"`
	Title         string `json:"Title"`
	Points        int    `json:"Points"`
	Author        string `json:"User"`
	Time          int64  `json:"Time"`
	URL           string `json:"URL"`
	Domain        string `json:"Domain"`
	CommentsCount int    `json:"CommentsCount"`
}

func loadLegacyJSON(path string) ([]*Item, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read favorites: %w", err)
	}

	var legacy []*legacyItem
	if err := json.Unmarshal(data, &legacy); err != nil {
		return nil, fmt.Errorf("could not parse favorites (file may be corrupted): %w", err)
	}

	items := make([]*Item, len(legacy))
	for i, l := range legacy {
		items[i] = &Item{
			ID:            l.ID,
			Title:         l.Title,
			Points:        l.Points,
			Author:        l.Author,
			Time:          l.Time,
			URL:           l.URL,
			Domain:        l.Domain,
			CommentsCount: l.CommentsCount,
		}
		sanitizeItem(items[i])
	}

	return items, nil
}
