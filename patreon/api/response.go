package api

type Response struct {
	Data     []ResponsePost    `json:"data"`
	Includes []ResponseInclude `json:"included"`
	Meta     ResponseMeta      `json:"meta"`
	Links    ResponseLinks     `json:"links"`
}

type ResponsePost struct {
	ID            string                    `json:"id"`
	Type          string                    `json:"type"`
	Attributes    ResponsePostAttributes    `json:"attributes"`
	RelationShips ResponsePostRelationships `json:"relationships"`
}

type ResponsePostAttributes struct {
	PostType           string               `json:"post_type"`
	Title              string               `json:"title"`
	PublishedAt        string               `json:"published_at"`
	URL                string               `json:"url"`
	CurrentUserCanView bool                 `json:"current_user_can_view"`
	TeaserText         string               `json:"teaser_text"`
	PostMetaData       ResponsePostMetaData `json:"post_metadata"`
}

type ResponsePostMetaData struct {
	ImageOrder []string `json:"image_order"`
}

type ResponsePostRelationships struct {
	Attachments ResponsePostRelationshipsAttachments `json:"attachments"`
	Images      ResponsePostRelationshipsImages      `json:"images"`
	Media       ResponsePostRelationshipsMedia       `json:"media"`
}

type ResponsePostRelationshipsAttachments struct {
	Data []any `json:"data"`
}

type ResponsePostRelationshipsImages struct {
	Data []ResponseMedia `json:"data"`
}

type ResponsePostRelationshipsMedia struct {
	Data []ResponseMedia `json:"data"`
}

type ResponseMedia struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

type ResponseInclude struct {
	ResponseMedia
	Attributes ResponseMediaAttributes `json:"attributes"`
}

type ResponseMediaAttributes struct {
	SizeBytes   int                    `json:"size_bytes"`
	MimeType    string                 `json:"mimetype"`
	DownloadURL string                 `json:"download_url"`
	ImageURLs   ResponseMediaImageURLs `json:"image_urls"`
	Metadata    ResponseMediaMetadata  `json:"metadata"`
}

type ResponseMediaImageURLs struct {
	URL                 string `json:"url"`
	Original            string `json:"original"`
	Default             string `json:"default"`
	DefaultBlurred      string `json:"default_blurred"`
	DefaultSmall        string `json:"default_small"`
	DefaultLarge        string `json:"default_large"`
	DefaultBlurredSmall string `json:"default_blurred_small"`
	Thumbnail           string `json:"thumbnail"`
	ThumbnailLarge      string `json:"thumbnail_large"`
	ThumbnailSmall      string `json:"thumbnail_small"`
}

type ResponseMediaMetadata struct {
	Dimensions ResponseMediaDimensions `json:"dimensions"`
}

type ResponseMediaDimensions struct {
	W int `json:"w"`
	H int `json:"h"`
}

type ResponseMeta struct {
	Pagination ResponsePagination `json:"pagination"`
}

type ResponsePagination struct {
	Total   int             `json:"total"`
	Cursors ResponseCursors `json:"cursors"`
}

type ResponseCursors struct {
	Next string `json:"next"`
}

type ResponseLinks struct {
	Next string `json:"next"`
}
