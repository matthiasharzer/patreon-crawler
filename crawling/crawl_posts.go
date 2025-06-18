package crawling

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"slices"
	"strings"

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

func getMediaFile(downloadDirectory string, media patreon.Media) (string, error) {
	extension, err := getFileExtension(media.MimeType)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s.%s", downloadDirectory, media.ID, extension), nil
}

func getFileExtension(mimeType string) (string, error) {
	// This is a very quick and dirty method, but it should work here
	mimeTypeSplits := strings.Split(mimeType, "/")
	if len(mimeTypeSplits) != 2 {
		return "", fmt.Errorf("invalid mime type: %s", mimeType)
	}
	return mimeTypeSplits[1], nil
}

func downloadMedia(media patreon.Media, downloadDir string) error {
	downloadedFilePath, err := getMediaFile(downloadDir, media)
	if err != nil {
		return err
	}

	_, err = os.Stat(downloadedFilePath)
	if err == nil {
		fmt.Printf("\t- skipped %s (already downloaded)\n", media.ID)
		return nil
	}

	response, err := http.Get(media.DownloadURL)
	if err != nil {
		return fmt.Errorf("failed to download media: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download media: %s", response.Status)
	}

	tempDownloadFilePath := downloadedFilePath + ".tmp"

	out, err := os.Create(tempDownloadFilePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	_, err = io.Copy(out, response.Body)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	out.Close()

	err = os.Rename(tempDownloadFilePath, downloadedFilePath)
	if err != nil {
		return fmt.Errorf("failed to rename file: %w", err)
	}

	fmt.Printf("\t- saved %s\n", media.ID)
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
		err := downloadMedia(media, downloadDirectory)
		if err != nil {
			return err
		}
	}
	return nil
}

func adjustPostsFileTime(downloadDirectory string, post patreon.Post) error {
	if len(post.Media) == 0 {
		return nil
	}
	for _, media := range post.Media {
		file, err := getMediaFile(downloadDirectory, media)
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
			fmt.Printf("[%d] Saving post '%s'\n", postsDownloaded, post.Title)

			err := downloadPost(downloadDir, post)
			if err != nil {
				return err
			}

			postsDownloaded++
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
