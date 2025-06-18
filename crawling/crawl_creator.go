package crawling

import (
	"fmt"

	"patreon-crawler/patreon"
	"patreon-crawler/patreon/api"
)

func CrawlCreator(apiClient *api.Client, creatorID string, downloadDir string, downloadInaccessibleMedia bool, groupingStrategy GroupingStrategy, downloadLimit int) error {

	creatorDownloadDir := fmt.Sprintf("%s/%s", downloadDir, sanitizeFilename(creatorID))

	fmt.Printf("Downloading posts from %s to %s\n", creatorID, creatorDownloadDir)

	client, err := patreon.NewClient(apiClient, creatorID)
	if err != nil {
		return err
	}

	err = CrawlPosts(client, creatorDownloadDir, downloadInaccessibleMedia, groupingStrategy, downloadLimit)
	if err != nil {
		return err
	}

	return nil
}
