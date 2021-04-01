package retriever

import (
	"clx/constants/submissions"
	"clx/core"
	"clx/favorites"
	"clx/sub"
	"encoding/json"
	"fmt"

	"gitlab.com/tslocum/cview"
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
	if category == submissions.Favorites {
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
	r.Submissions = make([]*Submissions, submissions.TotalNumberOfCategories)

	r.Submissions[submissions.FrontPage] = new(Submissions)
	r.Submissions[submissions.New] = new(Submissions)
	r.Submissions[submissions.Ask] = new(Submissions)
	r.Submissions[submissions.Show] = new(Submissions)
	r.Submissions[submissions.Favorites] = new(Submissions)

	r.Submissions[submissions.FrontPage].MaxPages = submissions.FrontPageMaxPages
	r.Submissions[submissions.New].MaxPages = submissions.NewMaxPages
	r.Submissions[submissions.Ask].MaxPages = submissions.AskMaxPages
	r.Submissions[submissions.Show].MaxPages = submissions.ShowMaxPages

	r.Submissions[submissions.Favorites].Entries = fav.Items
}

func (r *Retriever) Reset() {
	r.Submissions[submissions.FrontPage].PageToFetchFromAPI = 0
	r.Submissions[submissions.New].PageToFetchFromAPI = 0
	r.Submissions[submissions.Ask].PageToFetchFromAPI = 0
	r.Submissions[submissions.Show].PageToFetchFromAPI = 0

	r.Submissions[submissions.FrontPage].Entries = nil
	r.Submissions[submissions.New].Entries = nil
	r.Submissions[submissions.Ask].Entries = nil
	r.Submissions[submissions.Show].Entries = nil
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
	r.Submissions[submissions.Favorites].Entries = append(r.Submissions[submissions.Favorites].Entries, story)
}

func (r *Retriever) GetFavoritesJSON() ([]byte, error) {
	b, err := json.MarshalIndent(r.Submissions[submissions.Favorites].Entries, "", "    ")
	if err != nil {
		return nil, fmt.Errorf("could not serialize favorites struct: %w", err)
	}

	return b, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}

	return b
}
