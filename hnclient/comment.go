package hnclient

type Comment struct {
	By     string `json:"by"`
	Id     int    `json:"id"`
	Kids   []int  `json:"kids"`
	Parent int    `json:"parent"`
	Text   string `json:"text"`
	Time   int    `json:"time"`
}
