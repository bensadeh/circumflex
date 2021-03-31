package retriever

import (
	"clx/constants/submissions"
	"clx/core"
	"clx/sub"
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
	r.Submissions = make([]*Submissions, submissions.TotalNumberOfCategories)

	r.Submissions[submissions.FrontPage] = new(Submissions)
	r.Submissions[submissions.New] = new(Submissions)
	r.Submissions[submissions.Ask] = new(Submissions)
	r.Submissions[submissions.Show] = new(Submissions)

	r.Submissions[submissions.FrontPage].MaxPages = submissions.FrontPageMaxPages
	r.Submissions[submissions.New].MaxPages = submissions.NewMaxPages
	r.Submissions[submissions.Ask].MaxPages = submissions.AskMaxPages
	r.Submissions[submissions.Show].MaxPages = submissions.ShowMaxPages
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

func (r *Retriever) GetStory(category int, index int) *core.Submission {
	return r.Submissions[category].Entries[index]
}

func (r *Retriever) GetMaxPages(category int) int {
	return r.Submissions[category].MaxPages
}
