package core

type Artist struct {
	ID         string
	Name       string
	AlbumCount int
}

type Album struct {
	ID     string
	Title  string
	Artist string
	Year   int
}

type Track struct {
	ID       string
	Title    string
	Artist   string
	Album    string
	Year     int
	Track    int
	Duration int // seconds
}
