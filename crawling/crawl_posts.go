package crawling

import (
	"fmt"
	"os"
	"regexp"
	"slices"
	"strings"

	"patreon-crawler/crawling/download"
	"patreon-crawler/patreon"
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

func adjustPostsFileTime(downloadDirectory string, post patreon.Post) error {
	if len(post.Media) == 0 {
		return nil
	}
	for _, media := range post.Media {
		file, err := download.GetMediaFile(downloadDirectory, media)
		if err != nil {
			return err
		}

		_, err = os.Stat(file)
		if err != nil {
			// File does not exist
			return nil
		}

		err = os.Chtimes(file, post.PublishedAt, post.PublishedAt)
		if err != nil {
			return err
		}
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

func savePost(post patreon.Post, postsDownloaded int, downloadDir string) error {
	fmt.Printf("[%d] Saving post '%s'\n", postsDownloaded, post.Title)

	err := os.MkdirAll(downloadDir, 0700)
	if err != nil {
		return err
	}

	report := download.Post(downloadDir, post)
	for item := range report {
		switch item := item.(type) {
		case *download.ReportErrorItem:
			fmt.Printf("\t[error] %s: %s\n", item.Media.ID, item.Err)
		case *download.ReportSkippedItem:
			fmt.Printf("\t[skipped] %s (%s)\n", item.Media.ID, item.Reason)
		case *download.ReportSuccessItem:
			fmt.Printf("\t[downloaded] %s\n", item.Media.ID)
		}
	}
	return nil
}

func CrawlPosts(client *patreon.Client, baseDownloadDir string, downloadInaccessibleMedia bool, groupingStrategy GroupingStrategy, downloadLimit int) error {
	posts := client.Posts()

	postsDownloaded := 0
	postsCrawled := 0
	for post, err := range posts {
		if err != nil {
			return err
		}

		if postsDownloaded >= downloadLimit && downloadLimit != 0 {
			break
		}

		if len(post.Media) == 0 {
			fmt.Printf("[%d] Skipping post with no media '%s'\n", postsDownloaded, post.Title)
			continue
		}
		downloadDir, err := getDownloadDir(baseDownloadDir, post.Title, groupingStrategy)
		if err != nil {
			return err
		}

		postsCrawled++

		if post.CurrentUserCanView || downloadInaccessibleMedia {
			err := savePost(post, postsDownloaded, downloadDir)
			if err != nil {
				return err
			}

			postsDownloaded++
		} else {
			fmt.Printf("[%d] Skipping post '%s' (inaccessible media)\n", postsDownloaded, post.Title)
		}

		// This can be done independent to downloading
		err = adjustPostsFileTime(downloadDir, post)
		if err != nil {
			return err
		}
	}
	if postsCrawled == 0 {
		fmt.Println("No posts found")
	}

	downloadedFraction := float64(postsDownloaded) / float64(postsCrawled)
	if downloadedFraction < 0.8 {
		if postsDownloaded == 0 {
			fmt.Printf("Warning: No posts were downloaded. Did you provide the correct creator ID?\n")
		} else {
			fmt.Printf("Warning: Only %f%% of posts were downloaded. Did you provide correct creator ID?\n", downloadedFraction*100)
		}
	}

	return nil
}
