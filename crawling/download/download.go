package download

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/MatthiasHarzer/patreon-crawler/patreon"
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

func adjustFileTime(filePath string, publishedAt time.Time) error {
	_, err := os.Stat(filePath)
	if err != nil {
		// File does not exist
		return nil
	}

	err = os.Chtimes(filePath, publishedAt, publishedAt)
	if err != nil {
		return err
	}

	return nil
}

func Media(media patreon.Media, downloadDir string, modTime time.Time) ReportItem {
	if media.MimeType == "" {
		return NewSkippedItem(media, "no mime type")
	}

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

	err = os.MkdirAll(downloadDir, os.ModePerm)
	if err != nil {
		return NewErrorItem(media, fmt.Errorf("failed to create directory: %w", err))
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

	err = adjustFileTime(downloadedFilePath, modTime)
	if err != nil {
		return NewErrorItem(media, fmt.Errorf("failed to adjust file time: %w", err))
	}

	return NewSuccessItem(media)
}
