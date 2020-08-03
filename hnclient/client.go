package hnclient

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Client struct {
	BaseUrl      string
	UserSuffix   string
	ItemSuffix   string
	MaxSuffix    string
	TopSuffix    string
	NewSuffix    string
	JobSuffix    string
	AskSuffix    string
	ShowSuffix   string
	UpdateSuffix string
}

func NewClient() *Client {
	var c Client
	c.BaseUrl = "https://hacker-news.firebaseio.com/v0/"
	c.UserSuffix = "user/%s.json"
	c.ItemSuffix = "item/%d.json"
	c.MaxSuffix = "maxitem.json"
	c.TopSuffix = "topstories.json"
	c.NewSuffix = "newstories.json"
	c.JobSuffix = "jobstories.json"
	c.AskSuffix = "askstories.json"
	c.ShowSuffix = "showstories.json"
	c.UpdateSuffix = "updates.json"
	return &c
}

// Interesting. This function returns EOF on occasion (~1/10 times) - the below GetResource never does.
// Cursory exploration leads me to believe that the `req.Close = true` line is the differentiating factor
// func (c *Client) OldGetResource(url string) ([]byte, error) {
// 	response, err := http.Get(url)
// 	if err != nil {
// 		return nil, err
// 	} else {
// 		defer response.Body.Close()
// 		contents, err := ioutil.ReadAll(response.Body)
// 		if err != nil {
// 			return nil, err
// 		}
// 		return contents, err
// 	}
// }

func (c *Client) GetResource(url string) ([]byte, error) {
	httpClient := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	req.Close = true
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	response, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return response, err
}

// GetItem returns an Item struct with the information corresponding to the item with the provided id
func (c *Client) GetItem(id int) (Item, error) {
	url := c.BaseUrl + fmt.Sprintf(c.ItemSuffix, id)
	rep, err := c.GetResource(url)

	var i Item
	if err != nil {
		return i, err
	}

	err = json.Unmarshal(rep, &i)
	return i, err
}

// GetUser returns a User struct with the information of a user corresponding to the provided username
func (c *Client) GetUser(username string) (User, error) {
	url := c.BaseUrl + fmt.Sprintf(c.UserSuffix, username)
	rep, err := c.GetResource(url)

	var user User
	if err != nil {
		return user, err
	}

	err = json.Unmarshal(rep, &user)
	return user, err
}

// GetStory returns a Story struct with the information of a story corresponding to the provided id
func (c *Client) GetStory(id int) (Story, error) {
	item, err := c.GetItem(id)
	var story Story
	if err != nil {
		return story, err
	}
	if item.Type != "story" {
		return story, fmt.Errorf("Item with id %d is not a story", id)
	}
	return c.ItemToStory(item), nil
}

func (c *Client) ItemToStory(item Item) Story {
	story := Story{
		By:          item.By,
		Descendants: item.Descendants,
		Kids:        item.Kids,
		Score:       item.Score,
		Time:        item.Time,
		Title:       item.Title,
		Url:         item.Url,
	}
	return story
}

// GetComment returns a Comment struct with the information of a comment corresponding to the provided id
func (c *Client) GetComment(id int) (Comment, error) {
	item, err := c.GetItem(id)
	var comment Comment
	if err != nil {
		return comment, err
	}
	if item.Type != "comment" {
		return comment, fmt.Errorf("Item with id %d is not a comment", id)
	}
	return c.ItemToComment(item), nil
}

func (c *Client) ItemToComment(item Item) Comment {
	comment := Comment{
		By:     item.By,
		Id:     item.Id,
		Kids:   item.Kids,
		Parent: item.Parent,
		Text:   item.Text,
		Time:   item.Time,
	}
	return comment
}

// GetPoll returns a Poll struct with the information of a poll corresponding to the provided id
func (c *Client) GetPoll(id int) (Poll, error) {
	item, err := c.GetItem(id)
	var poll Poll
	if err != nil {
		return poll, err
	}
	if item.Type != "poll" {
		return poll, fmt.Errorf("Item with id %d is not a poll", id)
	}
	return c.ItemToPoll(item), nil
}

func (c *Client) ItemToPoll(item Item) Poll {
	poll := Poll{
		By:          item.By,
		Descendants: item.Descendants,
		Id:          item.Id,
		Kids:        item.Kids,
		Parts:       item.Parts,
		Text:        item.Text,
		Time:        item.Time,
		Title:       item.Title,
	}
	return poll
}

// GetPoll returns a Poll struct with the information of a poll corresponding to the provided id
func (c *Client) GetPollOpt(id int) (PollOpt, error) {
	item, err := c.GetItem(id)
	var pollopt PollOpt
	if err != nil {
		return pollopt, err
	}
	if item.Type != "pollopt" {
		return pollopt, fmt.Errorf("Item with id %d is not a pollopt", id)
	}
	pollopt = PollOpt{
		By:     item.By,
		Id:     item.Id,
		Parent: item.Parent,
		Score:  item.Score,
		Text:   item.Text,
		Time:   item.Time,
	}
	return c.ItemToPollOpt(item), nil
}

func (c *Client) ItemToPollOpt(item Item) PollOpt {
	pollopt := PollOpt{
		By:     item.By,
		Id:     item.Id,
		Parent: item.Parent,
		Score:  item.Score,
		Text:   item.Text,
		Time:   item.Time,
	}
	return pollopt
}

// GetTopStories takes an int number and returns an array of up to number ints that represent the current top stories.
// Constraints: 0 <= number <= 500
func (c *Client) GetTopStories(number int) ([]int, error) {
	var top500 []int
	if number > 500 {
		return top500, fmt.Errorf("Number %d greater than maximum 500 items allowed", number)
	}

	url := c.BaseUrl + c.TopSuffix
	rep, err := c.GetResource(url)

	err = json.Unmarshal(rep, &top500)

	if err != nil {
		return nil, err
	}

	return top500[:number], nil
}

// GetNewStories takes an int number and returns an array of up to number ints that represent the newest stories.
// Constraints: 0 <= number <= 500
func (c *Client) GetNewStories(number int) ([]int, error) {
	var top500 []int
	if number > 500 {
		return top500, fmt.Errorf("Number %d greater than maximum 500 items allowed", number)
	}

	url := c.BaseUrl + c.NewSuffix
	rep, err := c.GetResource(url)

	err = json.Unmarshal(rep, &top500)

	if err != nil {
		return nil, err
	}

	return top500[:number], nil
}

// GetRecentAskStories takes an int number and returns an array of up to number ints that represent the most recent Ask stories
// Constraints: 0 <= number <= 200
func (c *Client) GetRecentAskStories(number int) ([]int, error) {
	var top200 []int
	if number > 200 {
		return top200, fmt.Errorf("Number %d greater than maximum 500 items allowed", number)
	}

	url := c.BaseUrl + c.AskSuffix
	rep, err := c.GetResource(url)

	err = json.Unmarshal(rep, &top200)

	if err != nil {
		return nil, err
	}

	return top200[:number], nil
}

// GetRecentShowStories takes an int number and returns an array of up to number ints that represent the most recent Show stories
// Constraints: 0 <= number <= 200
func (c *Client) GetRecentShowStories(number int) ([]int, error) {
	var top200 []int
	if number > 200 {
		return top200, fmt.Errorf("Number %d greater than maximum 500 items allowed", number)
	}

	url := c.BaseUrl + c.ShowSuffix
	rep, err := c.GetResource(url)

	err = json.Unmarshal(rep, &top200)

	if err != nil {
		return nil, err
	}

	return top200[:number], nil
}

// GetRecentJobStories takes an int number and returns an array of up to number ints that represent the most recent Job stories
// Constraints: 0 <= number <= 200
func (c *Client) GetRecentJobStories(number int) ([]int, error) {
	var top200 []int
	if number > 200 {
		return top200, fmt.Errorf("Number %d greater than maximum 500 items allowed", number)
	}

	url := c.BaseUrl + c.JobSuffix
	rep, err := c.GetResource(url)

	err = json.Unmarshal(rep, &top200)

	if err != nil {
		return nil, err
	}

	return top200[:number], nil
}

// GetRecentChanges takes an int number and returns an array of up to number ints that represent the most recent Job stories
// Constraints: 0 <= number <= 200
func (c *Client) GetRecentChanges() (Changes, error) {
	var changes Changes
	url := c.BaseUrl + c.UpdateSuffix
	rep, err := c.GetResource(url)

	err = json.Unmarshal(rep, &changes)

	return changes, err
}

// GetMaxId returns the maximum id represented by an int
func (c *Client) GetMaxId() (int, error) {
	var max int
	url := c.BaseUrl + c.MaxSuffix
	rep, err := c.GetResource(url)
	err = json.Unmarshal(rep, &max)
	return max, err
}
