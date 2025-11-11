package download_test

import (
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/MatthiasHarzer/patreon-crawler/crawling/download"
	"github.com/MatthiasHarzer/patreon-crawler/patreon"
	"github.com/MatthiasHarzer/patreon-crawler/util/fsutils"
	"github.com/MatthiasHarzer/patreon-crawler/util/testutils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDownload(t *testing.T) {
	t.Run("downloads medias", func(t *testing.T) {
		url, cleanup := testutils.HTTPServer(map[string]http.HandlerFunc{
			"/media1.jpg": func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, err := w.Write([]byte("media1 content"))
				require.NoError(t, err)
			},
			"/media2.png": func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, err := w.Write([]byte("media2 content"))
				require.NoError(t, err)
			},
		})
		defer cleanup()

		medias := []patreon.Media{
			{
				ID:          "media1",
				DownloadURL: url.String() + "media1.jpg",
				MimeType:    "image/jpeg",
			},
			{
				ID:          "media2",
				DownloadURL: url.String() + "media2.png",
				MimeType:    "image/png",
			},
		}

		downloadDir, dirCleanup, err := fsutils.TemporaryDirectory()
		require.NoError(t, err)
		defer dirCleanup()

		for _, media := range medias {
			reportItem := download.Media(media, downloadDir)
			require.IsType(t, &download.ReportSuccessItem{}, reportItem)

			successItem := reportItem.(*download.ReportSuccessItem)
			assert.Equal(t, media.ID, successItem.Media.ID)
		}

		downloadDirMedia1Path := fmt.Sprintf("%s/media1.jpeg", downloadDir)
		downloadDirMedia2Path := fmt.Sprintf("%s/media2.png", downloadDir)
		content1, err := os.ReadFile(downloadDirMedia1Path)
		require.NoError(t, err)
		assert.Equal(t, "media1 content", string(content1))

		content2, err := os.ReadFile(downloadDirMedia2Path)
		require.NoError(t, err)
		assert.Equal(t, "media2 content", string(content2))
	})

	t.Run("skips already downloaded medias", func(t *testing.T) {
		url, cleanup := testutils.HTTPServer(map[string]http.HandlerFunc{
			"/media1.jpg": func(w http.ResponseWriter, r *http.Request) {
				assert.Fail(t, "When media is already downloaded, this handler should not be called")
			},
		})
		defer cleanup()

		media := patreon.Media{
			ID:          "media1",
			DownloadURL: url.String() + "media1.jpg",
			MimeType:    "image/jpeg",
		}

		downloadDir, dirCleanup, err := fsutils.TemporaryDirectory()
		require.NoError(t, err)
		defer dirCleanup()

		// Pre-create the file to simulate already downloaded media
		downloadedFilePath, err := download.GetMediaFile(downloadDir, media)
		require.NoError(t, err)
		err = os.WriteFile(downloadedFilePath, []byte("existing content"), 0644)
		require.NoError(t, err)

		reportItem := download.Media(media, downloadDir)
		require.IsType(t, &download.ReportSkippedItem{}, reportItem)

		skippedItem := reportItem.(*download.ReportSkippedItem)
		assert.Equal(t, media.ID, skippedItem.Media.ID)
		assert.Equal(t, "already downloaded", skippedItem.Reason)
	})

	t.Run("skips medias with no mime type", func(t *testing.T) {
		media := patreon.Media{
			ID:          "media1",
			DownloadURL: "http://example.com/media1.jpg",
			MimeType:    "",
		}

		downloadDir, dirCleanup, err := fsutils.TemporaryDirectory()
		require.NoError(t, err)
		defer dirCleanup()

		reportItem := download.Media(media, downloadDir)
		require.IsType(t, &download.ReportSkippedItem{}, reportItem)

		skippedItem := reportItem.(*download.ReportSkippedItem)
		assert.Equal(t, media.ID, skippedItem.Media.ID)
		assert.Equal(t, "no mime type", skippedItem.Reason)
	})

	t.Run("handles url errors", func(t *testing.T) {
		media := patreon.Media{
			ID:          "media1",
			DownloadURL: "http://invalid-url",
			MimeType:    "image/jpeg",
		}

		downloadDir, dirCleanup, err := fsutils.TemporaryDirectory()
		require.NoError(t, err)
		defer dirCleanup()

		reportItem := download.Media(media, downloadDir)
		require.IsType(t, &download.ReportErrorItem{}, reportItem)

		errorItem := reportItem.(*download.ReportErrorItem)
		assert.Equal(t, media.ID, errorItem.Media.ID)
		assert.NotNil(t, errorItem.Err)
	})

	t.Run("handles mime type errors", func(t *testing.T) {
		media := patreon.Media{
			ID:          "media1",
			DownloadURL: "http://example.com/media1.unknown",
			MimeType:    "invalid-mime-type",
		}

		downloadDir, dirCleanup, err := fsutils.TemporaryDirectory()
		require.NoError(t, err)
		defer dirCleanup()

		reportItem := download.Media(media, downloadDir)
		require.IsType(t, &download.ReportErrorItem{}, reportItem)

		errorItem := reportItem.(*download.ReportErrorItem)
		assert.Equal(t, media.ID, errorItem.Media.ID)
		assert.NotNil(t, errorItem.Err)
		assert.Equal(t, errorItem.Err.Error(), "invalid mime type: invalid-mime-type")
	})
}
