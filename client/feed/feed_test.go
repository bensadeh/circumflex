package feed

import (
	"reflect"
	"testing"
)

func TestNewFeedItem(t *testing.T) {

	longStr := `Lorem ipsum dolor sit amet, consectetuer adipiscing elit. Aenean commodo ligula eget dolor. Aenean massa. Cum sociis natoque penatibus et magnis dis parturient montes, nascetur ridiculus mus. Donec quam felis, ultricies nec, pellentesque eu, pretium quis, s`
	tests := []struct {
		name string
		want Item

		wantErr bool
	}{{"basic test", Item{Author: "asd", Title: "asda", URI: "http://google.com", Rank: 1, Points: 1, Comments: 1}, false},
		{"invalid uri", Item{Author: "asd", Title: "asda", URI: "a.com", Rank: 1, Points: 1, Comments: 1}, true},
		{"empty author", Item{Author: "", Title: "asda", URI: "http://a.com", Rank: 1, Points: 1, Comments: 1}, true},
		{"empty title", Item{Author: "asdsa", Title: "", URI: "http://a.com", Rank: 1, Points: 1, Comments: 1}, true},
		{"<0 rank", Item{Author: "asdsa", Title: "asda", URI: "http://a.com", Rank: -1, Points: 1, Comments: 1}, true},
		{"<0 points", Item{Author: "asd", Title: "asda", URI: "http://a.com", Rank: 1, Points: -1, Comments: 1}, true},
		{"<0 comments", Item{Author: "asd", Title: "asda", URI: "http://a.com", Rank: 1, Points: 1, Comments: -1}, true},
		{"author with more than 256 characters", Item{Author: longStr, Title: "asda", URI: "http://a.com", Rank: 1, Points: 1, Comments: -1}, true},
		{"title with more than 256 characters", Item{Author: "asd", Title: longStr, URI: "http://a.com", Rank: 1, Points: 1, Comments: -1}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewItem(tt.want.Title, tt.want.URI, tt.want.Author, tt.want.Points, tt.want.Comments, tt.want.Rank)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewFeedItem() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if !reflect.DeepEqual(&got, &tt.want) {
					t.Errorf("NewFeedItem() = %v, want %v", got, tt.want)
				}
			}

			if len(got.Author) > 256 {
				t.Errorf("author cannot be greater than 256 characters, got %d", len(got.Author))
			}
			if len(got.Title) > 256 {
				t.Errorf("title cannot be greater than 256 characters, got %d", len(got.Title))
			}
		})
	}

}
