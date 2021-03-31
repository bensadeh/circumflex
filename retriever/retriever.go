package retriever

import (
	"clx/constants/submissions"
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

	subs = append(subs, new(Submissions))
	subs = append(subs, new(Submissions))
	subs = append(subs, new(Submissions))
	subs = append(subs, new(Submissions))

	subs[submissions.FrontPage].MaxPages = submissions.FrontPageMaxPages
	subs[submissions.New].MaxPages = submissions.NewMaxPages
	subs[submissions.Ask].MaxPages = submissions.AskMaxPages
	subs[submissions.Show].MaxPages = submissions.ShowMaxPages

	r.Submissions = subs
}

func convert(subs []*core.Submission, highlightHeadlines int) []*core.ListItem {
	listItems := make([]*core.ListItem, len(subs))

	for i, s := range subs {
		item := new(core.ListItem)
		item.Main = sub.FormatSubMain(s.Title, s.Domain, highlightHeadlines)
		item.Secondary = sub.FormatSubSecondary(s.Points, s.Author, s.Time, s.CommentsCount)

		listItems[i] = item
	}

	return listItems
}

func (r *Retriever) GetStory(category int, index int) *core.Submission {
	return r.Submissions[category].Entries[index]
}

func (r *Retriever) GetMaxPages(category int) int {
	return r.Submissions[category].MaxPages
}
