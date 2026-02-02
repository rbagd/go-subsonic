package subsonic

type Indexes struct {
	LastModified int64    `json:"lastModified"`
	IgnoredArticles string `json:"ignoredArticles"`
	Index        []Index  `json:"index"`
}

type ArtistsID3 struct {
	Index []Index `json:"index"`
}

type Index struct {
	Name   string   `json:"name"`
	Artist []Artist `json:"artist"`
}

type Artist struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	CoverArt  string `json:"coverArt"`
	AlbumCount int   `json:"albumCount"`
	Starred   string `json:"starred,omitempty"` // timestamp
}

type Directory struct {
	ID       string  `json:"id"`
	Parent   string  `json:"parent,omitempty"`
	Name     string  `json:"name"`
	Child    []Child `json:"child"`
}

type Child struct {
	ID          string `json:"id"`
	Parent      string `json:"parent"`
	Title       string `json:"title"`
	Artist      string `json:"artist"`
	Album       string `json:"album"`
	CoverArt    string `json:"coverArt,omitempty"`
	IsDir       bool   `json:"isDir"` // true for nested folders/albums, false for songs
	Year        int    `json:"year,omitempty"`
	Duration    int    `json:"duration,omitempty"`
	Track       int    `json:"track,omitempty"`
	Size        int64  `json:"size,omitempty"`
	ContentType string `json:"contentType,omitempty"`
	Suffix      string `json:"suffix,omitempty"`
	Path        string `json:"path,omitempty"`
}
