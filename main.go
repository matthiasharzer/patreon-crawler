package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"slices"
	"strings"

	"patreon-crawler/patreon"
	"patreon-crawler/patreon/api"
)

var version = "version unknown"

var windowsReservedNames = []string{
	"CON", "PRN", "AUX", "NUL",
	"COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9",
	"LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9",
}

var invalidChars = regexp.MustCompile(`[<>:"/\\|?*\x00]`)

type GroupingStrategy string

const (
	GroupingStrategyNone   GroupingStrategy = "none"
	GroupingStrategyByPost GroupingStrategy = "by-post"
)

var argCreatorID string
var argCookie string
var argDownloadDir string
var argDownloadLimit int
var argDownloadInaccessibleMedia bool
var argGroupingStrategy string

func init() {
	flag.StringVar(&argCreatorID, "creator", "", "The creator ID to crawl")
	flag.StringVar(&argCookie, "cookie", "", "The cookie to use for authentication")
	flag.StringVar(&argDownloadDir, "download-dir", "", "The directory to download posts to")
	flag.IntVar(&argDownloadLimit, "download-limit", 0, "The maximum number of media files to download")
	flag.BoolVar(&argDownloadInaccessibleMedia, "download-inaccessible-media", false, "Whether to download inaccessible media")
	flag.StringVar(&argGroupingStrategy, "grouping", "none", "The grouping strategy to use. Must be one of: none, by-post")
	flag.Parse()
}

func isGroupingStrategy(strategy GroupingStrategy) bool {
	switch strategy {
	case GroupingStrategyNone, GroupingStrategyByPost:
		return true
	default:
		return false
	}
}

func getCookie() (string, error) {
	if argCookie != "" {
		return argCookie, nil
	}

	fmt.Println("Please enter your cookie from the patreon website: ")
	reader := bufio.NewReader(os.Stdin)
	cookieString, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read cookie: %w", err)
	}
	cookieString = cookieString[:len(cookieString)-1]
	return cookieString, nil
}

func getDownloadDir() (string, error) {
	if argDownloadDir != "" {
		return argDownloadDir, nil
	}

	fmt.Println("Please enter the download directory: ")
	reader := bufio.NewReader(os.Stdin)
	downloadDir, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read download directory: %w", err)
	}
	downloadDir = downloadDir[:len(downloadDir)-1]
	return downloadDir, nil
}

func sanitizeFilename(name string) string {
	name = invalidChars.ReplaceAllString(name, "_")
	name = strings.TrimRight(name, " .")
	upper := strings.ToUpper(name)
	if slices.Contains(windowsReservedNames, upper) {
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

func downloadMedia(media patreon.Media, downloadDir string) error {
	extension, err := getFileExtension(media.MimeType)
	if err != nil {
		return err
	}
	downloadedFilePath := fmt.Sprintf("%s/%s.%s", downloadDir, media.ID, extension)

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

func main() {
	if len(os.Args) == 1 {
		fmt.Printf("patreon-crawler %s\n", version)
		flag.PrintDefaults()

		fmt.Println("\nUsage:")
		fmt.Println("  patreon-crawler --creator <creator ID> --download-dir <download directory>")
		return
	}

	if argCreatorID == "" {
		panic("creator ID is required")
	}

	var groupingStrategy GroupingStrategy
	if argGroupingStrategy == "" {
		groupingStrategy = GroupingStrategyNone
	} else if isGroupingStrategy(GroupingStrategy(argGroupingStrategy)) {
		groupingStrategy = GroupingStrategy(argGroupingStrategy)
	} else {
		panic("invalid grouping strategy. Must be one of: none, by-post")
	}

	cookie, err := getCookie()
	if err != nil {
		panic(err)
	}
	downloadDir, err := getDownloadDir()
	if err != nil {
		panic(err)
	}
	creatorDownloadDir := fmt.Sprintf("%s/%s", downloadDir, sanitizeFilename(argCreatorID))

	client, err := patreon.NewClient(api.NewClient(cookie), argCreatorID)
	if err != nil {
		panic(err)
	}

	posts := client.Posts()

	postsDownloaded := 0
	postsCrawled := 0
	for post, err := range posts {
		if err != nil {
			panic(err)
		}

		if postsDownloaded >= argDownloadLimit && argDownloadLimit != 0 {
			break
		}

		postsCrawled++

		if !post.CurrentUserCanView && !argDownloadInaccessibleMedia {
			fmt.Printf("[%d] Skipping inaccessible post '%s'\n", postsDownloaded, post.Title)
			continue
		}

		if len(post.Media) == 0 {
			fmt.Printf("[%d] Skipping post '%s' with no media\n", postsDownloaded, post.Title)
			continue
		}

		err = nil
		switch groupingStrategy {
		case GroupingStrategyByPost:
			postDownloadDirectory := fmt.Sprintf("%s/%s", creatorDownloadDir, sanitizeFilename(post.Title))
			err = downloadPost(postDownloadDirectory, post)
		case GroupingStrategyNone:
			err = downloadPost(creatorDownloadDir, post)
		default:
			panic("invalid grouping strategy")
		}
		if err != nil {
			panic(err)
		}

		postsDownloaded++
		fmt.Printf("[%d] Downloaded post (%d image(s)) '%s'\n", postsDownloaded, len(post.Media), post.Title)
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
}
