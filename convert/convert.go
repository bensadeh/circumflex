package convert

import (
	"clx/comment"
	"clx/item"
)

// StoryToThread maps an item.Story (API layer) to a comment.Thread (rendering layer).
func StoryToThread(s *item.Story) *comment.Thread {
	return &comment.Thread{
		ID:            s.ID,
		Title:         s.Title,
		Author:        s.User,
		URL:           s.URL,
		Domain:        s.Domain,
		Points:        s.Points,
		TimeAgo:       s.TimeAgo,
		Content:       s.Content,
		CommentsCount: s.CommentsCount,
		Comments:      mapComments(s.Comments),
	}
}

func mapComments(stories []*item.Story) []*comment.Comment {
	if stories == nil {
		return nil
	}

	result := make([]*comment.Comment, 0, len(stories))

	for _, s := range stories {
		result = append(result, mapComment(s))
	}

	return result
}

func mapComment(s *item.Story) *comment.Comment {
	return &comment.Comment{
		ID:       s.ID,
		Author:   s.User,
		Content:  s.Content,
		Time:     s.Time,
		TimeAgo:  s.TimeAgo,
		Children: mapComments(s.Comments),
	}
}
