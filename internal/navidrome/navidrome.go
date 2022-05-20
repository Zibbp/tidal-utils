package navidrome

type Playlist struct {
	Name   string  `json:"name"`
	ID     string  `json:"id"`
	Tracks []Track `json:"tracks"`
}

type Track struct {
	Title  string `json:"title"`
	Artist string `json:"artist"`
	Album  string `json:"album"`
}
