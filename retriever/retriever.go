package retriever

import (
	"clx/core"
	"clx/sub"
	"fmt"
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
	hideYCJobs bool) ([]*core.ListItem, error) {
	largestItemToDisplay := (page * visibleStories) + visibleStories
	smallestItemToDisplay := page * visibleStories

	subs := r.Submissions[category]
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

func (r *Retriever) Init() {
	var subs []*Submissions

	subs = append(subs, &Submissions{MaxPages: 2}) // Front Page
	subs = append(subs, &Submissions{MaxPages: 2}) // new
	subs = append(subs, &Submissions{MaxPages: 0}) // Ask HN
	subs = append(subs, &Submissions{MaxPages: 0}) // Show HN

	r.Submissions = subs
}

func convert(subs []*core.Submission, highlightHeadlines int) []*core.ListItem {
	var listItems []*core.ListItem

	for _, s := range subs {
		item := new(core.ListItem)
		item.Main = sub.FormatSubMain(s.Title, s.Domain, highlightHeadlines)
		item.Secondary = sub.FormatSubSecondary(s.Points, s.Author, s.Time, s.CommentsCount)

		listItems = append(listItems, item)
	}

	return listItems
}

func (r *Retriever) GetStory(category int, index int) *core.Submission {
	return r.Submissions[category].Entries[index]
}

func (r *Retriever) GetMaxPages(category int) int {
	return r.Submissions[category].MaxPages
}
