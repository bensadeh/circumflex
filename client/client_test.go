package client

import (
	"reflect"
	"testing"

	"github.com/go-hackernews/client/feed"
	"github.com/go-hackernews/client/mocks"
	"github.com/golang/mock/gomock"
)

func TestHNClient_GetTopStories(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockScraper := mocks.NewMockScraper(ctrl)

	tests := []struct {
		name         string
		targetAmount int
		want         *[]feed.Item
		wantErr      bool
	}{{"basic test", 10, &[]feed.Item{{Author: "asd", Title: "asda", URI: "http://google.com", Rank: 1, Points: 1, Comments: 1}}, false},
		{"zero num stories", 0, nil, true},
		{"over 100 stories", 101, nil, true},
		{"100 stories", 100, nil, false}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hnc := &HNClient{
				scraper: mockScraper,
			}

			numTimes := 1
			if tt.wantErr {
				numTimes = 0
			}

			mockScraper.EXPECT().
				ScrapeHNFeed(tt.targetAmount).
				Return(tt.want, nil).Times(numTimes)

			got, err := hnc.GetTopStories(tt.targetAmount)
			if (err != nil) != tt.wantErr {
				t.Errorf("HNClient.GetTopStories() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("HNClient.GetTopStories() = %v, want %v", got, tt.want)
			}
		})
	}
}
