package patreon

import (
	"iter"
	"time"

	"patreon-crawler/patreon/api"
)

type Crawler struct {
	apiClient  *api.Client
	campaignID string
}

func NewClient(apiClient *api.Client, creatorID string) (*Crawler, error) {
	campaignID, err := apiClient.GetCampaignID(creatorID)
	if err != nil {
		return nil, err
	}
	return &Crawler{
		apiClient:  apiClient,
		campaignID: campaignID,
	}, nil
}

func (c *Crawler) GetPosts(cursor string) ([]Post, string, error) {
	postsResponse, err := c.apiClient.GetPosts(c.campaignID, &cursor)
	if err != nil {
		return nil, "", err
	}

	medias := make(map[string]Media)
	for _, include := range postsResponse.Includes {
		if include.Type != "media" {
			continue
		}

		downloadURL := include.Attributes.DownloadURL
		if downloadURL == "" {
			downloadURL = include.Attributes.ImageURLs.Original
		}

		medias[include.ID] = Media{
			ID:          include.ID,
			DownloadURL: downloadURL,
			MimeType:    include.Attributes.MimeType,
			Height:      include.Attributes.Metadata.Dimensions.H,
			Width:       include.Attributes.Metadata.Dimensions.W,
		}
	}

	var posts []Post
	for _, responsePost := range postsResponse.Data {
		if responsePost.Type != "post" {
			continue
		}

		mediaIDs := responsePost.Attributes.PostMetaData.ImageOrder
		var media []Media
		for _, mediaID := range mediaIDs {
			m, ok := medias[mediaID]
			if !ok {
				continue
			}

			media = append(media, m)
		}

		publishedAt, err := time.Parse(time.RFC3339, responsePost.Attributes.PublishedAt)
		if err != nil {
			return nil, "", err
		}

		posts = append(posts, Post{
			ID:                 responsePost.ID,
			Title:              responsePost.Attributes.Title,
			Media:              media,
			PublishedAt:        publishedAt,
			CurrentUserCanView: responsePost.Attributes.CurrentUserCanView,
		})
	}

	nextCursor := postsResponse.Meta.Pagination.Cursors.Next

	return posts, nextCursor, nil
}

func (c *Crawler) Posts() iter.Seq2[Post, error] {
	return func(yield func(Post, error) bool) {
		var cursor string
		for {
			posts, nextCursor, err := c.GetPosts(cursor)
			if err != nil {
				yield(Post{}, err)
				return
			}

			for _, post := range posts {
				if !yield(post, nil) {
					return
				}
			}

			if nextCursor == "" {
				return
			}

			cursor = nextCursor
		}
	}
}
