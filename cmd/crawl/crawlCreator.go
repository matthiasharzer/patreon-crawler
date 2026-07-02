package crawl

import (
	"fmt"
	"sync"

	"github.com/MatthiasHarzer/patreon-crawler/crawling"
	"github.com/MatthiasHarzer/patreon-crawler/crawling/download"
	"github.com/MatthiasHarzer/patreon-crawler/patreon"
	"github.com/MatthiasHarzer/patreon-crawler/patreon/api"
	"github.com/fatih/color"
)

type mediaPair struct {
	post  patreon.Post
	media patreon.Media
}

func crawlCreator(creatorID string, apiClient api.Client, downloader *crawling.Downloader, downloadLimit int, downloadInaccessibleMedia bool) error {
	client, err := patreon.NewClient(apiClient, creatorID)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	vanityID := client.VanityID()

	var mediaPairs []mediaPair
	totalPostsDiscovered := 0
	inaccessiblePostsSkipped := 0
	for post, err := range client.Posts() {
		if err != nil {
			return err
		}
		fmt.Printf("\rDiscovered %s posts with %s media files.", color.GreenString("%d", totalPostsDiscovered), color.GreenString("%d", len(mediaPairs)))
		totalPostsDiscovered++

		if !post.CurrentUserCanView && !downloadInaccessibleMedia {
			inaccessiblePostsSkipped++
			continue
		}

		for _, media := range post.Media {
			mediaPairs = append(mediaPairs, mediaPair{post: post, media: media})
		}

		if downloadLimit > 0 && len(mediaPairs) >= downloadLimit {
			if len(mediaPairs) > downloadLimit {
				mediaPairs = mediaPairs[:downloadLimit]
			}
			break
		}
	}

	fmt.Printf("\rDiscovered %s posts with %s media files.\n", color.GreenString("%d", totalPostsDiscovered), color.GreenString("%d", len(mediaPairs)))
	if inaccessiblePostsSkipped > 0 {
		fmt.Printf("Skipped %s inaccessible posts.\n", color.YellowString("%d", inaccessiblePostsSkipped))
	}

	fmt.Println("Downloading media...")

	printMutex := sync.Mutex{}
	for _, pair := range mediaPairs {
		downloader.Enqueue(vanityID, pair.post, pair.media, func(reportItem download.ReportItem) {
			printMutex.Lock()
			defer printMutex.Unlock()
			switch item := reportItem.(type) {
			case *download.ReportErrorItem:
				fmt.Printf("[%s] %s from post \"%s\": %s\n", color.RedString("error"), item.Media.ID, color.RedString(pair.post.Title), item.Err)
			case *download.ReportSkippedItem:
				fmt.Printf("[%s] %s from post \"%s\" (%s)\n", color.YellowString("skipped"), item.Media.ID, color.YellowString(pair.post.Title), color.RGB(100, 100, 100).Sprintf(item.Reason))
			case *download.ReportSuccessItem:
				fmt.Printf("[%s] %s from post \"%s\"\n", color.GreenString("downloaded"), item.Media.ID, color.GreenString(pair.post.Title))
			}
		})
	}

	err = downloader.ProcessAll()
	if err != nil {
		return err
	}

	return nil
}
