package hnclient

type PollOpt struct {
	By     string `json:"by"`
	Id     int    `json:"id"`
	Parent int    `json:"parent"`
	Score  int    `json:"score"`
	Text   string `json:"text"`
	Time   int    `json:"time"`
}
