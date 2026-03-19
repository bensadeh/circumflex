package firebase

type hnItem struct {
	ID          int    `json:"id"`
	Type        string `json:"type"`
	By          string `json:"by"`
	Time        int64  `json:"time"`
	Text        string `json:"text"`
	Title       string `json:"title"`
	URL         string `json:"url"`
	Score       int    `json:"score"`
	Kids        []int  `json:"kids"`
	Descendants int    `json:"descendants"`
	Parent      int    `json:"parent"`
	Parts       []int  `json:"parts"` // poll option IDs (polls only)
	Poll        int    `json:"poll"`  // parent poll ID (pollopts only)
	Deleted     bool   `json:"deleted"`
	Dead        bool   `json:"dead"`
}
