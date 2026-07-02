package crawling

import (
	"fmt"

	"github.com/MatthiasHarzer/patreon-crawler/crawling/download"
	"github.com/MatthiasHarzer/patreon-crawler/patreon"
	"github.com/MatthiasHarzer/patreon-crawler/queue"
	"github.com/MatthiasHarzer/patreon-crawler/util/fsutils"
)

type GroupingStrategy string

const (
	GroupingStrategyNone   GroupingStrategy = "none"
	GroupingStrategyByPost GroupingStrategy = "by-post"
)

func getDownloadDir(baseDownloadDir, postTitle string, groupingStrategy GroupingStrategy) (string, error) {
	switch groupingStrategy {
	case GroupingStrategyByPost:
		return fmt.Sprintf("%s/%s", baseDownloadDir, fsutils.SanitizeFilename(postTitle)), nil
	case GroupingStrategyNone:
		return baseDownloadDir, nil
	default:
		return "", fmt.Errorf("invalid grouping strategy")
	}
}

type Downloader struct {
	baseDownloadDir  string
	groupingStrategy GroupingStrategy
	downloadQueue    *queue.Queue
}

func NewDownloader(baseDownloadDir string, concurrencyLimit int, groupingStrategy GroupingStrategy) (*Downloader, error) {
	downloadQueue, err := queue.New(concurrencyLimit)
	if err != nil {
		return nil, err
	}
	return &Downloader{
		baseDownloadDir:  baseDownloadDir,
		groupingStrategy: groupingStrategy,
		downloadQueue:    downloadQueue,
	}, nil
}

func (d *Downloader) Enqueue(creatorVanityID string, parentPost patreon.Post, media patreon.Media, onDone func(item download.ReportItem)) {
	d.downloadQueue.Enqueue(func() error {
		creatorDownloadDir := fmt.Sprintf("%s/%s", d.baseDownloadDir, fsutils.SanitizeFilename(creatorVanityID))
		postDownloadDir, err := getDownloadDir(creatorDownloadDir, parentPost.Title, d.groupingStrategy)
		if err != nil {
			return fmt.Errorf("failed to get download directory: %w", err)
		}

		reportItem := download.Media(media, postDownloadDir, parentPost.PublishedAt)

		onDone(reportItem)
		return nil
	})
}

func (d *Downloader) ProcessAll() error {
	return d.downloadQueue.ProcessAll()
}
