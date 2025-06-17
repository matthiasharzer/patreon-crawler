package patreon

type Post struct {
	ID                 string
	Title              string
	Media              []Media
	CurrentUserCanView bool
}

type Media struct {
	ID          string
	Height      int
	Width       int
	DownloadURL string
	MimeType    string
}
