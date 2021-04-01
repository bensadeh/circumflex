package retriever

import (
	"clx/comment"
	"clx/constants/categories"
	"clx/core"
	"clx/favorites"
	"clx/file"
	"clx/sub"
	"encoding/json"
	"fmt"

	"gitlab.com/tslocum/cview"
)

const (
	totalNumberOfCategories = 5
	frontPageMaxPages       = 2
	newMaxPages             = 2
	askMaxPages             = 0
	showMaxPages            = 0
)

type Retriever struct {
	Submissions []*Submissions
}

type Submissions struct {
	MaxPages           int
	PageToFetchFromAPI int
	Entries            []*core.Submission
}

func (r *Retriever) GetSubmissions(category int, page int, visibleStories int, highlightHeadlines int,
	hideYCJobs bool) ([]*cview.ListItem, error) {
	if category == categories.Favorites {
		return getOfflineSubmissions(page, visibleStories, highlightHeadlines, r.Submissions[category])
	}

	return getOnlineSubmissions(category, page, visibleStories, highlightHeadlines, hideYCJobs, r.Submissions[category])
}

func getOfflineSubmissions(page int, visibleStories int, highlightHeadlines int,
	subs *Submissions) ([]*cview.ListItem, error) {
	storiesToShow := min(visibleStories, len(subs.Entries))
	smallestItemToDisplay := page * storiesToShow
	largestItemToDisplay := (page * storiesToShow) + storiesToShow

	listItems := convert(subs.Entries[smallestItemToDisplay:largestItemToDisplay], highlightHeadlines)

	return listItems, nil
}

func getOnlineSubmissions(category int, page int, visibleStories int, highlightHeadlines int, hideYCJobs bool,
	subs *Submissions) ([]*cview.ListItem, error) {
	smallestItemToDisplay := page * visibleStories
	largestItemToDisplay := (page * visibleStories) + visibleStories

	downloadedSubmissions := len(subs.Entries)
	pageHasEnoughSubmissionsToView := downloadedSubmissions > largestItemToDisplay

	if pageHasEnoughSubmissionsToView {
		listItems := convert(subs.Entries[smallestItemToDisplay:largestItemToDisplay], highlightHeadlines)

		return listItems, nil
	}

	subs.PageToFetchFromAPI++

	newSubmissions, err := sub.FetchSubmissions(subs.PageToFetchFromAPI, category)
	if err != nil {
		return nil, fmt.Errorf("could not fetch submissions: %w", err)
	}

	filteredSubmissions := sub.Filter(newSubmissions, hideYCJobs)
	subs.Entries = append(subs.Entries, filteredSubmissions...)

	listItems := convert(subs.Entries[smallestItemToDisplay:largestItemToDisplay], highlightHeadlines)

	return listItems, nil
}

func (r *Retriever) Init(fav *favorites.Favorites) {
	r.Submissions = make([]*Submissions, totalNumberOfCategories)

	r.Submissions[categories.FrontPage] = new(Submissions)
	r.Submissions[categories.New] = new(Submissions)
	r.Submissions[categories.Ask] = new(Submissions)
	r.Submissions[categories.Show] = new(Submissions)
	r.Submissions[categories.Favorites] = new(Submissions)

	r.Submissions[categories.FrontPage].MaxPages = frontPageMaxPages
	r.Submissions[categories.New].MaxPages = newMaxPages
	r.Submissions[categories.Ask].MaxPages = askMaxPages
	r.Submissions[categories.Show].MaxPages = showMaxPages

	r.Submissions[categories.Favorites].Entries = fav.Items
}

func (r *Retriever) Reset() {
	r.Submissions[categories.FrontPage].PageToFetchFromAPI = 0
	r.Submissions[categories.New].PageToFetchFromAPI = 0
	r.Submissions[categories.Ask].PageToFetchFromAPI = 0
	r.Submissions[categories.Show].PageToFetchFromAPI = 0

	r.Submissions[categories.FrontPage].Entries = nil
	r.Submissions[categories.New].Entries = nil
	r.Submissions[categories.Ask].Entries = nil
	r.Submissions[categories.Show].Entries = nil
}

func convert(subs []*core.Submission, highlightHeadlines int) []*cview.ListItem {
	listItems := make([]*cview.ListItem, len(subs))

	for i, s := range subs {
		main := sub.FormatSubMain(s.Title, s.Domain, highlightHeadlines)
		secondary := sub.FormatSubSecondary(s.Points, s.Author, s.Time, s.CommentsCount)

		item := cview.NewListItem(main)
		item.SetSecondaryText(secondary)

		listItems[i] = item
	}

	return listItems
}

func (r *Retriever) GetStory(category, currentItemIndex, submissionsToShow, currentPage int) *core.Submission {
	index := currentItemIndex + submissionsToShow*(currentPage)

	return r.Submissions[category].Entries[index]
}

func (r *Retriever) GetMaxPages(category int) int {
	return r.Submissions[category].MaxPages
}

func (r *Retriever) AddItemToFavorites(story *core.Submission) {
	r.Submissions[categories.Favorites].Entries = append(r.Submissions[categories.Favorites].Entries, story)
}

func (r *Retriever) GetFavoritesJSON() ([]byte, error) {
	b, err := json.MarshalIndent(r.Submissions[categories.Favorites].Entries, "", "    ")
	if err != nil {
		return nil, fmt.Errorf("could not serialize favorites struct: %w", err)
	}

	return b, nil
}

func (r *Retriever) UpdateFavoriteStoryAndWriteToDisk(updatedStory *comment.Comments) {
	for i, s := range r.Submissions[categories.Favorites].Entries {
		if s.ID == updatedStory.ID {
			isFieldsUpdated := s.Title != updatedStory.Title || s.Points != updatedStory.Points ||
				s.CommentsCount != updatedStory.CommentsCount || s.URL != updatedStory.URL ||
				s.Domain != updatedStory.Domain

			if isFieldsUpdated {
				r.Submissions[categories.Favorites].Entries[i].Title = updatedStory.Title
				r.Submissions[categories.Favorites].Entries[i].Points = updatedStory.Points
				r.Submissions[categories.Favorites].Entries[i].CommentsCount = updatedStory.CommentsCount
				r.Submissions[categories.Favorites].Entries[i].URL = updatedStory.URL
				r.Submissions[categories.Favorites].Entries[i].Domain = updatedStory.Domain

				bytes, _ := r.GetFavoritesJSON()

				err := file.WriteToFile(file.PathToFavoritesFile(), string(bytes))
				if err != nil {
					panic(err)
				}
			}
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}

	return b
}
