package patreon

import "time"

type Post struct {
	ID                 string
	Title              string
	Media              []Media
	PublishedAt        time.Time
	CurrentUserCanView bool
}

type Media struct {
	ID          string
	Height      int
	Width       int
	DownloadURL string
	MimeType    string
}
