package handler

import (
	"clx/constants/categories"
	"clx/endpoints"
	"clx/favorites"
	"clx/file"
	"clx/header"
	"clx/history"
	"clx/utils/filter"
	"clx/utils/formatter"
	"clx/utils/http"
	"encoding/json"
	"fmt"

	"github.com/gdamore/tcell/v2"

	"code.rocketnine.space/tslocum/cview"
)

const (
	totalNumberOfCategories = 5

	frontPageMaxPages = 2
	newMaxPages       = 2
	askMaxPages       = 0
	showMaxPages      = 0
	favoritesMaxPages = 2
)

type StoryHandler struct {
	sc      []*storyCategory
	history *history.History
}

type storyCategory struct {
	maxPages           int
	pageToFetchFromAPI int
	stories            []*endpoints.Story
}

func (r *StoryHandler) GetStories(category int, page int, visibleStories int, highlightHeadlines int,
	hideYCJobs bool) ([]*cview.ListItem, error) {
	if category == categories.Favorites {
		return getFavoritesStories(page, visibleStories, highlightHeadlines, r.sc[category], r.history)
	}

	return getOnlineStories(category, page, visibleStories, highlightHeadlines, hideYCJobs, r.sc[category], r.history)
}

func getFavoritesStories(page int, visibleStories int, highlightHeadlines int, sc *storyCategory,
	his *history.History) ([]*cview.ListItem,
	error) {
	storiesToShow := min(visibleStories, len(sc.stories))
	firstItemToDisplay := page * storiesToShow
	lastItemToDisplay := min(firstItemToDisplay+storiesToShow, len(sc.stories))

	listItems := convert(sc.stories[firstItemToDisplay:lastItemToDisplay], his, highlightHeadlines, true)

	return listItems, nil
}

func getOnlineStories(category int, page int, visibleStories int, highlightHeadlines int, hideYCJobs bool,
	sc *storyCategory, his *history.History) ([]*cview.ListItem, error) {
	overriddenYCJobsStatus := getOverriddenYCJobsStatus(visibleStories, hideYCJobs)
	smallestItemToDisplay := page * visibleStories
	largestItemToDisplay := (page * visibleStories) + visibleStories

	downloadedStories := len(sc.stories)
	pageHasEnoughStoriesToView := downloadedStories > largestItemToDisplay

	if pageHasEnoughStoriesToView {
		listItems := convert(sc.stories[smallestItemToDisplay:largestItemToDisplay], his, highlightHeadlines, false)

		return listItems, nil
	}

	sc.pageToFetchFromAPI++

	newStories, err := http.FetchStories(sc.pageToFetchFromAPI, category)
	if err != nil {
		return nil, fmt.Errorf("could not fetch storyCategory: %w", err)
	}

	filteredStories := filter.Filter(newStories, overriddenYCJobsStatus)
	sc.stories = append(sc.stories, filteredStories...)

	listItems := convert(sc.stories[smallestItemToDisplay:largestItemToDisplay], his, highlightHeadlines, false)

	return listItems, nil
}

func getOverriddenYCJobsStatus(visibleStories int, hideYCJobs bool) bool {
	if visibleStories >= 28 {
		return false
	}

	return hideYCJobs
}

func (r *StoryHandler) Init(fav *favorites.Favorites, his *history.History) {
	r.sc = make([]*storyCategory, totalNumberOfCategories)

	r.sc[categories.FrontPage] = new(storyCategory)
	r.sc[categories.New] = new(storyCategory)
	r.sc[categories.Ask] = new(storyCategory)
	r.sc[categories.Show] = new(storyCategory)
	r.sc[categories.Favorites] = new(storyCategory)

	r.sc[categories.FrontPage].maxPages = frontPageMaxPages
	r.sc[categories.New].maxPages = newMaxPages
	r.sc[categories.Ask].maxPages = askMaxPages
	r.sc[categories.Show].maxPages = showMaxPages
	r.sc[categories.Favorites].maxPages = favoritesMaxPages

	r.sc[categories.Favorites].stories = fav.Items
	r.history = his
}

func (r *StoryHandler) Reset() {
	r.sc[categories.FrontPage].pageToFetchFromAPI = 0
	r.sc[categories.New].pageToFetchFromAPI = 0
	r.sc[categories.Ask].pageToFetchFromAPI = 0
	r.sc[categories.Show].pageToFetchFromAPI = 0

	r.sc[categories.FrontPage].stories = nil
	r.sc[categories.New].stories = nil
	r.sc[categories.Ask].stories = nil
	r.sc[categories.Show].stories = nil
}

func convert(subs []*endpoints.Story, his *history.History, highlightHeadlines int,
	isOnFavorites bool) []*cview.ListItem {
	listItems := make([]*cview.ListItem, len(subs))

	for i, s := range subs {
		_ = his.Contains(s.ID) && !isOnFavorites

		main := formatter.FormatMain(s.Title, s.Domain, s.Author, highlightHeadlines)
		secondary := formatter.FormatSecondary(s.Points, s.Author, s.Time, s.CommentsCount, highlightHeadlines)

		item := cview.NewListItem(main)
		item.SetSecondaryText(secondary)

		listItems[i] = item
	}

	return listItems
}

func (r *StoryHandler) GetStory(category, currentItemIndex, storiesToShow, currentPage int) *endpoints.Story {
	index := getIndex(currentItemIndex, storiesToShow, currentPage)

	return r.sc[category].stories[index]
}

func (r *StoryHandler) DeleteStoryAndWriteToFile(category, currentItemIndex, storiesToShow, currentPage int) {
	index := getIndex(currentItemIndex, storiesToShow, currentPage)
	r.sc[category].stories = removeIndex(r.sc[category].stories, index)
	write(r)
}

func getIndex(currentItemIndex, storiesToShow, currentPage int) int {
	return currentItemIndex + storiesToShow*(currentPage)
}

func removeIndex(s []*endpoints.Story, index int) []*endpoints.Story {
	return append(s[:index], s[index+1:]...)
}

func (r *StoryHandler) GetMaxPages(category int, storiesToShow int) int {
	if category == categories.Favorites {
		fav := r.sc[categories.Favorites].stories
		favItems := len(fav) - 1
		availablePages := favItems / storiesToShow

		return min(availablePages, favoritesMaxPages)
	}

	return r.sc[category].maxPages
}

func (r *StoryHandler) AddItemToFavoritesAndWriteToFile(story *endpoints.Story) error {
	r.sc[categories.Favorites].stories = append(r.sc[categories.Favorites].stories, story)

	bytes, _ := r.GetFavoritesJSON()
	filePath := file.PathToFavoritesFile()

	err := file.WriteToFileInConfigDir(filePath, string(bytes))
	if err != nil {
		return fmt.Errorf("could not write to file: %w", err)
	}

	return nil
}

func (r *StoryHandler) GetFavoritesJSON() ([]byte, error) {
	b, err := json.MarshalIndent(r.sc[categories.Favorites].stories, "", "    ")
	if err != nil {
		return nil, fmt.Errorf("could not serialize favorites struct: %w", err)
	}

	return b, nil
}

func (r *StoryHandler) UpdateFavoriteStoryAndWriteToDisk(newStory *endpoints.Comments) {
	for i, s := range r.sc[categories.Favorites].stories {
		if s.ID == newStory.ID {
			isFieldsUpdated := s.Title != newStory.Title || s.Points != newStory.Points ||
				s.Time != newStory.Time || s.Author != newStory.User ||
				s.CommentsCount != newStory.CommentsCount || s.URL != newStory.URL ||
				s.Domain != newStory.Domain

			if isFieldsUpdated {
				r.sc[categories.Favorites].stories[i].Title = newStory.Title
				r.sc[categories.Favorites].stories[i].Points = newStory.Points
				r.sc[categories.Favorites].stories[i].Time = newStory.Time
				r.sc[categories.Favorites].stories[i].Author = newStory.User
				r.sc[categories.Favorites].stories[i].CommentsCount = newStory.CommentsCount
				r.sc[categories.Favorites].stories[i].URL = newStory.URL
				r.sc[categories.Favorites].stories[i].Domain = newStory.Domain

				write(r)
			}
		}
	}
}

func (r *StoryHandler) GetHackerNewsHeader(currentCategory int) string {
	fav := r.sc[categories.Favorites].stories
	showFavorites := len(fav) != 0

	return header.GetHackerNewsHeader(currentCategory, showFavorites)
}

func (r *StoryHandler) GetNewCategory(event *tcell.EventKey, currentCategory int) int {
	if event.Key() == tcell.KeyBacktab {
		return r.getPreviousCategory(currentCategory)
	}

	return r.getNextCategory(currentCategory)
}

func (r *StoryHandler) getNextCategory(currentCategory int) int {
	isAtLastCategory := currentCategory == (r.getTotalNumberOfCategories() - 1)

	if isAtLastCategory {
		return 0
	}

	return currentCategory + 1
}

func (r *StoryHandler) getPreviousCategory(currentCategory int) int {
	isAtFirstCategory := currentCategory == 0

	if isAtFirstCategory {
		return r.getTotalNumberOfCategories() - 1
	}

	return currentCategory - 1
}

func (r *StoryHandler) getTotalNumberOfCategories() int {
	fav := r.sc[categories.Favorites].stories
	hasFavorites := len(fav) != 0

	if hasFavorites {
		return totalNumberOfCategories
	}

	return totalNumberOfCategories - 1
}

func write(r *StoryHandler) {
	bytes, _ := r.GetFavoritesJSON()

	err := file.WriteToFileInConfigDir(file.PathToFavoritesFile(), string(bytes))
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
