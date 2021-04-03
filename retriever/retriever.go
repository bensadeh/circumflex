package retriever

import (
	"clx/comment"
	"clx/constants/categories"
	"clx/core"
	"clx/favorites"
	"clx/file"
	"clx/header"
	"clx/sub"
	"encoding/json"
	"fmt"

	"github.com/gdamore/tcell/v2"

	"gitlab.com/tslocum/cview"
)

const (
	totalNumberOfCategories = 5

	frontPageMaxPages = 2
	newMaxPages       = 2
	askMaxPages       = 0
	showMaxPages      = 0
	favoritesMaxPages = 2
)

type Retriever struct {
	submissions []*Submissions
}

type Submissions struct {
	maxPages           int
	pageToFetchFromAPI int
	entries            []*core.Submission
}

func (r *Retriever) GetSubmissions(category int, page int, visibleStories int, highlightHeadlines int,
	hideYCJobs bool) ([]*cview.ListItem, error) {
	if category == categories.Favorites {
		return getOfflineSubmissions(page, visibleStories, highlightHeadlines, r.submissions[categories.Favorites])
	}

	return getOnlineSubmissions(category, page, visibleStories, highlightHeadlines, hideYCJobs, r.submissions[category])
}

func getOfflineSubmissions(page int, visibleStories int, highlightHeadlines int,
	subs *Submissions) ([]*cview.ListItem, error) {
	storiesToShow := min(visibleStories, len(subs.entries))
	firstItemToDisplay := page * storiesToShow
	lastItemToDisplay := min(firstItemToDisplay+storiesToShow, len(subs.entries))

	listItems := convert(subs.entries[firstItemToDisplay:lastItemToDisplay], highlightHeadlines)

	return listItems, nil
}

func getOnlineSubmissions(category int, page int, visibleStories int, highlightHeadlines int, hideYCJobs bool,
	subs *Submissions) ([]*cview.ListItem, error) {
	smallestItemToDisplay := page * visibleStories
	largestItemToDisplay := (page * visibleStories) + visibleStories

	downloadedSubmissions := len(subs.entries)
	pageHasEnoughSubmissionsToView := downloadedSubmissions > largestItemToDisplay

	if pageHasEnoughSubmissionsToView {
		listItems := convert(subs.entries[smallestItemToDisplay:largestItemToDisplay], highlightHeadlines)

		return listItems, nil
	}

	subs.pageToFetchFromAPI++

	newSubmissions, err := sub.FetchSubmissions(subs.pageToFetchFromAPI, category)
	if err != nil {
		return nil, fmt.Errorf("could not fetch submissions: %w", err)
	}

	filteredSubmissions := sub.Filter(newSubmissions, hideYCJobs)
	subs.entries = append(subs.entries, filteredSubmissions...)

	listItems := convert(subs.entries[smallestItemToDisplay:largestItemToDisplay], highlightHeadlines)

	return listItems, nil
}

func (r *Retriever) Init(fav *favorites.Favorites) {
	r.submissions = make([]*Submissions, totalNumberOfCategories)

	r.submissions[categories.FrontPage] = new(Submissions)
	r.submissions[categories.New] = new(Submissions)
	r.submissions[categories.Ask] = new(Submissions)
	r.submissions[categories.Show] = new(Submissions)
	r.submissions[categories.Favorites] = new(Submissions)

	r.submissions[categories.FrontPage].maxPages = frontPageMaxPages
	r.submissions[categories.New].maxPages = newMaxPages
	r.submissions[categories.Ask].maxPages = askMaxPages
	r.submissions[categories.Show].maxPages = showMaxPages
	r.submissions[categories.Favorites].maxPages = favoritesMaxPages

	r.submissions[categories.Favorites].entries = fav.Items
}

func (r *Retriever) Reset() {
	r.submissions[categories.FrontPage].pageToFetchFromAPI = 0
	r.submissions[categories.New].pageToFetchFromAPI = 0
	r.submissions[categories.Ask].pageToFetchFromAPI = 0
	r.submissions[categories.Show].pageToFetchFromAPI = 0

	r.submissions[categories.FrontPage].entries = nil
	r.submissions[categories.New].entries = nil
	r.submissions[categories.Ask].entries = nil
	r.submissions[categories.Show].entries = nil
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
	index := getIndex(currentItemIndex, submissionsToShow, currentPage)

	return r.submissions[category].entries[index]
}

func (r *Retriever) DeleteStoryAndWriteToFile(category, currentItemIndex, submissionsToShow, currentPage int) {
	index := getIndex(currentItemIndex, submissionsToShow, currentPage)
	r.submissions[category].entries = removeIndex(r.submissions[category].entries, index)
	write(r)
}

func getIndex(currentItemIndex, submissionsToShow, currentPage int) int {
	return currentItemIndex + submissionsToShow*(currentPage)
}

func removeIndex(s []*core.Submission, index int) []*core.Submission {
	return append(s[:index], s[index+1:]...)
}

func (r *Retriever) GetMaxPages(category int, submissionsToShow int) int {
	if category == categories.Favorites {
		fav := r.submissions[categories.Favorites].entries
		favItems := len(fav) - 1
		availablePages := favItems / submissionsToShow

		return min(availablePages, favoritesMaxPages)
	}

	return r.submissions[category].maxPages
}

func (r *Retriever) AddItemToFavoritesAndWriteToFile(story *core.Submission) error {
	r.submissions[categories.Favorites].entries = append(r.submissions[categories.Favorites].entries, story)

	bytes, _ := r.GetFavoritesJSON()
	filePath := file.PathToFavoritesFile()

	err := file.WriteToFile(filePath, string(bytes))
	if err != nil {
		return fmt.Errorf("could not write to file: %w", err)
	}

	return nil
}

func (r *Retriever) GetFavoritesJSON() ([]byte, error) {
	b, err := json.MarshalIndent(r.submissions[categories.Favorites].entries, "", "    ")
	if err != nil {
		return nil, fmt.Errorf("could not serialize favorites struct: %w", err)
	}

	return b, nil
}

func (r *Retriever) UpdateFavoriteStoryAndWriteToDisk(newStory *comment.Comments) {
	for i, s := range r.submissions[categories.Favorites].entries {
		if s.ID == newStory.ID {
			isFieldsUpdated := s.Title != newStory.Title || s.Points != newStory.Points ||
				s.Time != newStory.Time || s.Author != newStory.User ||
				s.CommentsCount != newStory.CommentsCount || s.URL != newStory.URL ||
				s.Domain != newStory.Domain

			if isFieldsUpdated {
				r.submissions[categories.Favorites].entries[i].Title = newStory.Title
				r.submissions[categories.Favorites].entries[i].Points = newStory.Points
				r.submissions[categories.Favorites].entries[i].Time = newStory.Time
				r.submissions[categories.Favorites].entries[i].Author = newStory.User
				r.submissions[categories.Favorites].entries[i].CommentsCount = newStory.CommentsCount
				r.submissions[categories.Favorites].entries[i].URL = newStory.URL
				r.submissions[categories.Favorites].entries[i].Domain = newStory.Domain

				write(r)
			}
		}
	}
}

func (r *Retriever) GetHackerNewsHeader(currentCategory int) string {
	fav := r.submissions[categories.Favorites].entries
	showFavorites := len(fav) != 0

	return header.GetHackerNewsHeader(currentCategory, showFavorites)
}

func (r *Retriever) GetNewCategory(event *tcell.EventKey, appState *core.ApplicationState) int {
	if event.Key() == tcell.KeyBacktab {
		return r.getPreviousCategory(appState.CurrentCategory)
	}

	return r.getNextCategory(appState.CurrentCategory)
}

func (r *Retriever) getNextCategory(currentCategory int) int {
	isAtLastCategory := currentCategory == (r.getTotalNumberOfCategories() - 1)

	if isAtLastCategory {
		return 0
	}

	return currentCategory + 1
}

func (r *Retriever) getPreviousCategory(currentCategory int) int {
	isAtFirstCategory := currentCategory == 0

	if isAtFirstCategory {
		return r.getTotalNumberOfCategories() - 1
	}

	return currentCategory - 1
}

func (r *Retriever) getTotalNumberOfCategories() int {
	fav := r.submissions[categories.Favorites].entries
	hasFavorites := len(fav) != 0

	if hasFavorites {
		return totalNumberOfCategories
	}

	return totalNumberOfCategories - 1
}

func write(r *Retriever) {
	bytes, _ := r.GetFavoritesJSON()

	err := file.WriteToFile(file.PathToFavoritesFile(), string(bytes))
	if err != nil {
		panic(err)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}

	return b
}
