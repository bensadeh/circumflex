package scrape

import (
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/go-hackernews/client/feed"
	"github.com/gocolly/colly"
)

func NewTestServer(expectedOutput string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		io.WriteString(w, expectedOutput)
	}))

}

func TestHNClient_ScrapeHNFeed(t *testing.T) {

	type args struct {
		targetAmount int
		expected     string
	}

	tests := []struct {
		name string
		args args
		want *[]feed.Item

		wantErr bool
	}{{"standard test", args{targetAmount: 1, expected: `
	<table id="hnmain">
	<tr>
	<td><table class="itemlist">
	<tr class="athing">
	<td align="right" valign="top" class="title"><span class="rank">1.</span></td>      <td valign="top" class="votelinks"><center><a id='up_20335198' href='vote?id=20335198&amp;how=up&amp;goto=news%3Fp%3D1'><div class='votearrow' title='upvote'></div></a></center></td><td class="title"><a href="http://test.com" class="storylink">Minnesota</a></td></tr><tr><td colspan="2"></td><td class="subtext">
	<span class="score" id="score_20335198">178 points</span> by <a href="user?id=pseudolus" class="hnuser">pseudolus</a> <span class="age"><a href="item?id=20335198">3 hours ago</a></span> <span id="unv_20335198"></span> | <a href="hide?id=20335198&amp;goto=news%3Fp%3D1">hide</a> | <a href="item?id=20335198">58&nbsp;comments</a>              </td></tr>
	</tr>
	</table>
	</td>
	</tr>
	</table>
	`},
		&[]feed.Item{{Author: "pseudolus", Title: "Minnesota", URI: "http://test.com", Points: 178, Rank: 1, Comments: 58}}, false},
		{"standard multiple rows", args{targetAmount: 2, expected: `
	<table id="hnmain">
	<tr>
	<td><table class="itemlist">
	<tr class="athing">
	<td align="right" valign="top" class="title"><span class="rank">1.</span></td>      <td valign="top" class="votelinks"><center><a id='up_20335198' href='vote?id=20335198&amp;how=up&amp;goto=news%3Fp%3D1'><div class='votearrow' title='upvote'></div></a></center></td><td class="title"><a href="http://test.com" class="storylink">Minnesota</a></td></tr><tr><td colspan="2"></td><td class="subtext">
	<span class="score" id="score_20335198">178 points</span> by <a href="user?id=pseudolus" class="hnuser">pseudolus</a> <span class="age"><a href="item?id=20335198">3 hours ago</a></span> <span id="unv_20335198"></span> | <a href="hide?id=20335198&amp;goto=news%3Fp%3D1">hide</a> | <a href="item?id=20335198">58&nbsp;comments</a>              </td></tr>
	</tr>
	<tr class="spacer"></tr>
	<tr class="athing">
	<td align="right" valign="top" class="title"><span class="rank">1.</span></td>      <td valign="top" class="votelinks"><center><a id='up_20335198' href='vote?id=20335198&amp;how=up&amp;goto=news%3Fp%3D1'><div class='votearrow' title='upvote'></div></a></center></td><td class="title"><a href="http://test.com" class="storylink">Minnesota</a></td></tr><tr><td colspan="2"></td><td class="subtext">
	<span class="score" id="score_20335198">178 points</span> by <a href="user?id=pseudolus" class="hnuser">pseudolus</a> <span class="age"><a href="item?id=20335198">3 hours ago</a></span> <span id="unv_20335198"></span> | <a href="hide?id=20335198&amp;goto=news%3Fp%3D1">hide</a> | <a href="item?id=20335198">58&nbsp;comments</a>              </td></tr>
	</tr>
	</table>
	</td>
	</tr>
	</table>
		`},
			&[]feed.Item{{Author: "pseudolus", Title: "Minnesota", URI: "http://test.com", Points: 178, Rank: 1, Comments: 58}, {Author: "pseudolus", Title: "Minnesota", URI: "http://test.com", Points: 178, Rank: 1, Comments: 58}}, false},
		{"malformed data", args{targetAmount: 1, expected: `
		<table id="hnmain">

		</table>
		`},
			nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testServer := NewTestServer(tt.args.expected)
			defer testServer.Close()
			scraper := NewScraper(testServer.URL)

			got, err := scraper.ScrapeHNFeed(tt.args.targetAmount)

			if (err != nil) != tt.wantErr {
				t.Errorf("Scraper.ScrapeHNFeed() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(&got, &tt.want) {
				t.Errorf("Scraper.ScrapeHNFeed() = %v, want %v", got, tt.want)
			}
		})
	}
}

//With more time investigate how to mocks colly.HTMLElement
func Test_processItem(t *testing.T) {
	type args struct {
		e *colly.HTMLElement
	}
	tests := []struct {
		name    string
		args    args
		want    *feed.Item
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := processFeedItem(tt.args.e)
			if (err != nil) != tt.wantErr {
				t.Errorf("processItem() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("processItem() = %v, want %v", got, tt.want)
			}
		})
	}
}
