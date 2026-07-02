package crawl

import (
	"bufio"
	"fmt"
	"math"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/MatthiasHarzer/patreon-crawler/crawling"
	"github.com/MatthiasHarzer/patreon-crawler/patreon/api"
	"github.com/spf13/cobra"
)

var argCookie string
var argDownloadDir string
var argDownloadLimit = math.MaxInt
var argDownloadInaccessibleMedia bool
var argGroupingStrategy string
var argConcurrencyLimit = 4

func init() {
	Command.Flags().StringVarP(&argCookie, "cookie", "c", argCookie, "The cookie to use for authentication")
	Command.Flags().StringVarP(&argDownloadDir, "download-dir", "d", argDownloadDir, "The directory to download posts to")
	Command.Flags().IntVarP(&argDownloadLimit, "download-limit", "l", argDownloadLimit, "The maximum number of posts to download")
	Command.Flags().BoolVarP(&argDownloadInaccessibleMedia, "download-inaccessible-media", "", argDownloadInaccessibleMedia, "Whether to download inaccessible media")
	Command.Flags().StringVarP(&argGroupingStrategy, "grouping", "g", argGroupingStrategy, "The grouping strategy to use. Must be one of: none, by-post")
	Command.Flags().IntVarP(&argConcurrencyLimit, "concurrency", "", argConcurrencyLimit, "The number of concurrent downloads")
}

func isValidGroupingStrategy(strategy crawling.GroupingStrategy) bool {
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

func getAPIClientFromStdIn() (api.Client, string, error) {
	var apiClient api.Client
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

func getAPIClient(cookie string) (api.Client, error) {
	if cookie != "" {
		apiClient := api.NewClient(cookie)
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

func getDownloadDir(defaultDownloadDir string) (string, error) {
	if defaultDownloadDir != "" {
		return defaultDownloadDir, nil
	}

	fmt.Println("Please enter the download directory: ")
	reader := bufio.NewReader(os.Stdin)
	downloadDir, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read download directory: %w", err)
	}
	downloadDir = strings.TrimSpace(downloadDir)
	return downloadDir, nil
}

var Command = &cobra.Command{
	Use:   "crawl <creator-id>",
	Short: "Crawl a patreon creator and download their posts",
	Args:  cobra.ExactArgs(1),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if argDownloadLimit < 0 {
			return fmt.Errorf("download limit must be non-negative")
		}
		if argConcurrencyLimit <= 0 {
			return fmt.Errorf("concurrency limit must be positive")
		}
		if argGroupingStrategy != "" && !isValidGroupingStrategy(crawling.GroupingStrategy(argGroupingStrategy)) {
			return fmt.Errorf("invalid grouping strategy. Must be one of: none, by-post")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		var groupingStrategy crawling.GroupingStrategy
		if argGroupingStrategy == "" {
			groupingStrategy = crawling.GroupingStrategyNone
		} else {
			groupingStrategy = crawling.GroupingStrategy(argGroupingStrategy)
		}

		downloadDir, err := getDownloadDir(argDownloadDir)
		if err != nil {
			return fmt.Errorf("failed to get download directory: %w", err)
		}

		apiClient, err := getAPIClient(argCookie)
		if err != nil {
			return fmt.Errorf("failed to get API client: %w", err)
		}

		creatorID := args[0]

		downloader, err := crawling.NewDownloader(downloadDir, argConcurrencyLimit, groupingStrategy)
		if err != nil {
			return fmt.Errorf("failed to create downloader: %w", err)
		}

		err = crawlCreator(creatorID, apiClient, downloader, argDownloadLimit, argDownloadInaccessibleMedia)
		if err != nil {
			return fmt.Errorf("failed to crawl: %w", err)
		}

		return nil
	},
}
