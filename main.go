package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"patreon-crawler/crawling"
	"patreon-crawler/patreon/api"
)

var version = "version unknown"

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

func isGroupingStrategy(strategy crawling.GroupingStrategy) bool {
	switch strategy {
	case crawling.GroupingStrategyNone, crawling.GroupingStrategyByPost:
		return true
	default:
		return false
	}
}

func getCookie() (string, error) {
	if argCookie != "" {
		return strings.TrimSpace(argCookie), nil
	}

	fmt.Println("Please enter your cookie from the patreon website: ")
	reader := bufio.NewReader(os.Stdin)
	cookieString, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read cookie: %w", err)
	}
	cookieString = cookieString[:len(cookieString)-1]
	return strings.TrimSpace(cookieString), nil
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

	var groupingStrategy crawling.GroupingStrategy
	if argGroupingStrategy == "" {
		groupingStrategy = crawling.GroupingStrategyNone
	} else if isGroupingStrategy(crawling.GroupingStrategy(argGroupingStrategy)) {
		groupingStrategy = crawling.GroupingStrategy(argGroupingStrategy)
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

	apiClient := api.NewClient(cookie)

	err = crawling.CrawlCreator(apiClient, argCreatorID, downloadDir, argDownloadInaccessibleMedia, groupingStrategy, argDownloadLimit)
	if err != nil {
		panic(err)
	}
}
