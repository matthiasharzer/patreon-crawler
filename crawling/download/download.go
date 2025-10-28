package download

import (
	"fmt"
	"io"
	"iter"
	"net/http"
	"os"
	"strings"

	"patreon-crawler/patreon"
)

func GetMediaFile(downloadDirectory string, media patreon.Media) (string, error) {
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

func downloadMedia(media patreon.Media, downloadDir string) ReportItem {
	downloadedFilePath, err := GetMediaFile(downloadDir, media)
	if err != nil {
		return NewErrorItem(media, err)
	}

	_, err = os.Stat(downloadedFilePath)
	if err == nil {
		return NewSkippedItem(media, "already downloaded")
	}

	response, err := http.Get(media.DownloadURL)
	if err != nil {
		return NewErrorItem(media, err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return NewErrorItem(media, fmt.Errorf("unexpected status code: %s", response.Status))
	}

	tempDownloadFilePath := downloadedFilePath + ".tmp"

	out, err := os.Create(tempDownloadFilePath)
	if err != nil {
		return NewErrorItem(media, fmt.Errorf("failed to create file: %w", err))
	}
	defer out.Close()

	_, err = io.Copy(out, response.Body)
	if err != nil {
		return NewErrorItem(media, fmt.Errorf("failed to write file: %w", err))
	}

	out.Close()

	err = os.Rename(tempDownloadFilePath, downloadedFilePath)
	if err != nil {
		return NewErrorItem(media, fmt.Errorf("failed to rename file: %w", err))
	}

	return NewSuccessItem(media)
}

func Post(downloadDirectory string, post patreon.Post) iter.Seq[ReportItem] {
	return func(yield func(ReportItem) bool) {
		if len(post.Media) == 0 {
			return
		}

		for _, media := range post.Media {
			if media.MimeType == "" {
				item := NewSkippedItem(media, "no mime type")
				if !yield(item) {
					return
				}
				continue
			}

			item := downloadMedia(media, downloadDirectory)
			if !yield(item) {
				return
			}
		}
	}
}
