package crawling

import (
	"fmt"
	"iter"
	"os"
	"regexp"
	"slices"
	"strings"
	"time"

	"patreon-crawler/crawling/download"
	"patreon-crawler/patreon"
	"patreon-crawler/patreon/api"
	"patreon-crawler/queue"
)

type GroupingStrategy string

const (
	GroupingStrategyNone   GroupingStrategy = "none"
	GroupingStrategyByPost GroupingStrategy = "by-post"
)

var fileNameWindowsReservedNames = []string{
	"CON", "PRN", "AUX", "NUL",
	"COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9",
	"LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9",
}

var fileNameInvalidChars = regexp.MustCompile(`[<>:"/\\|?*\x00]`)

func sanitizeFilename(name string) string {
	name = fileNameInvalidChars.ReplaceAllString(name, "_")
	name = strings.TrimRight(name, " .")
	upper := strings.ToUpper(name)
	if slices.Contains(fileNameWindowsReservedNames, upper) {
		name = "_" + name
	}
	if len(name) > 255 {
		name = name[:255]
	}
	if name == "" {
		name = "_"
	}
	return name
}

func adjustMediaFileTime(downloadDirectory string, media patreon.Media, publishedAt time.Time) error {
	file, err := download.GetMediaFile(downloadDirectory, media)
	if err != nil {
		return err
	}

	_, err = os.Stat(file)
	if err != nil {
		// File does not exist
		return nil
	}

	err = os.Chtimes(file, publishedAt, publishedAt)
	if err != nil {
		return err
	}

	return nil
}

func getDownloadDir(baseDownloadDir, postTitle string, groupingStrategy GroupingStrategy) (string, error) {
	switch groupingStrategy {
	case GroupingStrategyByPost:
		return fmt.Sprintf("%s/%s", baseDownloadDir, sanitizeFilename(postTitle)), nil
	case GroupingStrategyNone:
		return baseDownloadDir, nil
	default:
		return "", fmt.Errorf("invalid grouping strategy")
	}
}

type mediaPair struct {
	post  patreon.Post
	media patreon.Media
}

type Crawler struct {
	apiClient                 *api.Client
	downloadInaccessibleMedia bool
	groupingStrategy          GroupingStrategy
	downloadLimit             int
	concurrencyLimit          int
}

func NewCrawler(apiClient *api.Client, downloadInaccessibleMedia bool, groupingStrategy GroupingStrategy, downloadLimit int, concurrencyLimit int) *Crawler {
	return &Crawler{
		apiClient:                 apiClient,
		downloadInaccessibleMedia: downloadInaccessibleMedia,
		groupingStrategy:          groupingStrategy,
		downloadLimit:             downloadLimit,
		concurrencyLimit:          concurrencyLimit,
	}
}

func (c *Crawler) enumerateMedia(posts iter.Seq2[patreon.Post, error]) iter.Seq2[mediaPair, error] {
	return func(yield func(mediaPair, error) bool) {
		for post, err := range posts {
			if err != nil {
				if !yield(mediaPair{}, err) {
					return
				}
				continue
			}

			if !post.CurrentUserCanView && !c.downloadInaccessibleMedia {
				continue
			}

			for _, media := range post.Media {
				if !yield(mediaPair{post: post, media: media}, nil) {
					return
				}
			}
		}
	}
}

func (c *Crawler) downloadMedia(media patreon.Media, downloadDir string, parentPost patreon.Post) error {
	reportItem := download.Media(media, downloadDir)
	switch item := reportItem.(type) {
	case *download.ReportErrorItem:
		fmt.Printf("\t[error] %s from `%s`: %s\n", item.Media.ID, parentPost.Title, item.Err)
	case *download.ReportSkippedItem:
		fmt.Printf("\t[skipped] %s from `%s` (%s)\n", item.Media.ID, parentPost.Title, item.Reason)
	case *download.ReportSuccessItem:
		fmt.Printf("\t[downloaded] %s from `%s`\n", item.Media.ID, parentPost.Title)
	}
	err := adjustMediaFileTime(downloadDir, media, parentPost.PublishedAt)
	if err != nil {
		fmt.Printf("\t[error] adjusting file time for %s from %s: %s\n", media.ID, parentPost.Title, err)
	}
	return nil
}

func (c *Crawler) CrawlPosts(client *patreon.Client, baseDownloadDir string) error {
	posts := client.Posts()
	media := c.enumerateMedia(posts)

	q := queue.New[patreon.Media](c.concurrencyLimit)
	discoveredMediaCount := 0

	for pair, err := range media {
		if err != nil {
			return err
		}

		downloadDir, err := getDownloadDir(baseDownloadDir, pair.post.Title, c.groupingStrategy)
		if err != nil {
			return err
		}

		q.Enqueue(pair.media, func(media patreon.Media) error {
			return c.downloadMedia(media, downloadDir, pair.post)
		})

		discoveredMediaCount++
	}

	fmt.Printf("Discovered %d media items to download.\n", discoveredMediaCount)

	err := q.ProcessAll()

	return err
}

func (c *Crawler) CrawlCreator(creatorID string, downloadDir string) error {
	creatorDownloadDir := fmt.Sprintf("%s/%s", downloadDir, sanitizeFilename(creatorID))

	fmt.Printf("Downloading posts from %s to %s\n", creatorID, creatorDownloadDir)

	client, err := patreon.NewClient(c.apiClient, creatorID)
	if err != nil {
		return err
	}

	err = c.CrawlPosts(client, creatorDownloadDir)
	if err != nil {
		return err
	}

	return nil
}
