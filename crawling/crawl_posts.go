package crawling

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"slices"
	"strings"
	"time"

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

func getFileExtension(mimeType string) (string, error) {
	// This is a very quick and dirty method, but it should work here
	mimeTypeSplits := strings.Split(mimeType, "/")
	if len(mimeTypeSplits) != 2 {
		return "", fmt.Errorf("invalid mime type: %s", mimeType)
	}
	return mimeTypeSplits[1], nil
}

func downloadMedia(media patreon.Media, downloadDir string) (string, error) {
	extension, err := getFileExtension(media.MimeType)
	if err != nil {
		return "", err
	}
	downloadedFilePath := fmt.Sprintf("%s/%s.%s", downloadDir, media.ID, extension)

	if _, err := os.Stat(downloadedFilePath); err == nil {
		fmt.Printf("\t- skiped %s (already downloaded)\n", media.ID)
		return downloadedFilePath, nil
	}

	response, err := http.Get(media.DownloadURL)
	if err != nil {
		return "", fmt.Errorf("failed to download media: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download media: %s", response.Status)
	}

	tempDownloadFilePath := downloadedFilePath + ".tmp"

	out, err := os.Create(tempDownloadFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	_, err = io.Copy(out, response.Body)
	if err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	out.Close()

	err = os.Rename(tempDownloadFilePath, downloadedFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to rename file: %w", err)
	}

	fmt.Printf("\t- saved %s\n", media.ID)
	return downloadedFilePath, nil
}

func adjustFileTime(file string, time time.Time) error {
	err := os.Chtimes(file, time, time)
	if err != nil {
		return fmt.Errorf("failed to adjust file time: %w", err)
	}
	return nil
}

func downloadPost(downloadDirectory string, post patreon.Post) error {
	if len(post.Media) == 0 {
		return nil
	}

	err := os.MkdirAll(downloadDirectory, 0700)
	if err != nil {
		return fmt.Errorf("failed to create download directory: %w", err)
	}

	for _, media := range post.Media {
		file, err := downloadMedia(media, downloadDirectory)
		if err != nil {
			return err
		}
		err = adjustFileTime(file, post.PublishedAt)
		if err != nil {
			return err
		}
	}
	return nil
}

func CrawlPosts(client *patreon.Client, downloadDir string, downloadInaccessibleMedia bool, groupingStrategy GroupingStrategy, downloadLimit int) error {
	posts := client.Posts()

	postsDownloaded := 0
	postsCrawled := 0
	for post, err := range posts {
		if err != nil {
			panic(err)
		}

		if postsDownloaded >= downloadLimit && downloadLimit != 0 {
			break
		}

		if len(post.Media) == 0 {
			fmt.Printf("[%d] Skipping post with no media '%s'\n", postsDownloaded, post.Title)
			continue
		}

		postsCrawled++

		if !post.CurrentUserCanView && !downloadInaccessibleMedia {
			fmt.Printf("[%d] Skipping inaccessible post '%s'\n", postsDownloaded, post.Title)
			continue
		}

		fmt.Printf("[%d] Saving post '%s'\n", postsDownloaded, post.Title)

		err = nil
		switch groupingStrategy {
		case GroupingStrategyByPost:
			postDownloadDirectory := fmt.Sprintf("%s/%s", downloadDir, sanitizeFilename(post.Title))
			err = downloadPost(postDownloadDirectory, post)
		case GroupingStrategyNone:
			err = downloadPost(downloadDir, post)
		default:
			panic("invalid grouping strategy")
		}
		if err != nil {
			panic(err)
		}

		postsDownloaded++
	}
	if postsCrawled == 0 {
		fmt.Println("No posts found")
	}

	downloadedFraction := float64(postsDownloaded) / float64(postsCrawled)
	if downloadedFraction < 0.8 {
		if postsDownloaded == 0 {
			fmt.Printf("Warning: No posts were downloaded. Did you provide a valid cookie string?\n")
		} else {
			fmt.Printf("Warning: Only %f%% of posts were downloaded. Did you provide a valid cookie string?\n", downloadedFraction*100)
		}
	}

	return nil
}
