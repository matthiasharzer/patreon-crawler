package main

import (
	"bufio"
	"flag"
	"fmt"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
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

func cookieCacheFile() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	tokenCacheDir := filepath.Join(usr.HomeDir, ".credentials")
	err = os.MkdirAll(tokenCacheDir, 0700)
	if err != nil {
		return "", err
	}
	return filepath.Join(tokenCacheDir,
		url.QueryEscape("patreon.cookie")), nil
}

func saveCookieToFile(cookie string) error {
	cookieFile, err := cookieCacheFile()
	if err != nil {
		return err
	}
	return os.WriteFile(cookieFile, []byte(cookie), 0600)
}

func readCookieFromFile() (string, error) {
	cookieFile, err := cookieCacheFile()
	if err != nil {
		return "", err
	}

	cookie, err := os.ReadFile(cookieFile)
	if err != nil {
		return "", err
	}
	return string(cookie), nil
}

func readCookieFromStdin() (string, error) {
	reader := bufio.NewReader(os.Stdin)
	cookie, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(cookie), nil
}

func getAPIClientFromStdIn() (*api.Client, string, error) {
	var apiClient *api.Client
	var cookie string
	var err error
	authenticated := false

	for !authenticated {
		fmt.Println("Please enter your cookie from the patreon website: ")
		cookie, err = readCookieFromStdin()
		if err != nil {
			return nil, "", err
		}
		apiClient = api.NewClient(cookie)
		authenticated, err = apiClient.IsAuthenticated()
		if err != nil {
			return nil, "", err
		}
		if !authenticated {
			fmt.Println("Unable to authenticate with the provided cookie. Please try again.")
		}
	}

	return apiClient, cookie, nil
}

func getAPIClient() (*api.Client, error) {
	if argCookie != "" {
		apiClient := api.NewClient(argCookie)
		authenticated, err := apiClient.IsAuthenticated()
		if err != nil {
			return nil, err
		}
		if authenticated {
			return apiClient, nil
		}
		return nil, fmt.Errorf("failed to authenticate with cookie provided via --cookie")
	}

	cookie, err := readCookieFromFile()
	if err == nil {
		apiClient := api.NewClient(cookie)
		authenticated, err := apiClient.IsAuthenticated()
		if err != nil {
			return nil, err
		}
		if authenticated {
			return apiClient, nil
		}
	}

	apiClient, cookie, err := getAPIClientFromStdIn()
	if err != nil {
		return nil, err
	}

	err = saveCookieToFile(cookie)
	if err != nil {
		return nil, err
	}

	return apiClient, nil
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
		fmt.Println("  patreon-crawler --creator <creator-dir> --download-dir <directory>")
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

	downloadDir, err := getDownloadDir()
	if err != nil {
		panic(err)
	}

	apiClient, err := getAPIClient()
	if err != nil {
		panic(err)
	}

	err = crawling.CrawlCreator(apiClient, argCreatorID, downloadDir, argDownloadInaccessibleMedia, groupingStrategy, argDownloadLimit)
	if err != nil {
		panic(err)
	}
}
