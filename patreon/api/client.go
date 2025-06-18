package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
)

var patreonCampaignRegex = regexp.MustCompile(`patreon-media/p/campaign/(\d+)/.`)

const apiURL = "https://www.patreon.com/api"

func creatorURL(creatorID string) string {
	return "https://www.patreon.com/" + creatorID
}

type Client struct {
	cookie string
}

func NewClient(cookie string) *Client {
	return &Client{
		cookie: cookie,
	}
}

func (c *Client) doAPIRequest(path string, options map[string]string) (*http.Response, error) {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	urlBuilder := bytes.Buffer{}
	urlBuilder.WriteString(apiURL)
	urlBuilder.WriteString(path)
	urlBuilder.WriteString("?")

	for key, value := range options {
		urlBuilder.WriteString(key)
		urlBuilder.WriteString("=")
		urlBuilder.WriteString(value)
		urlBuilder.WriteString("&")
	}
	urlBuilder.Truncate(urlBuilder.Len() - 1)

	requestURL := urlBuilder.String()

	request, err := http.NewRequest(http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, err
	}

	request.Header.Add("Cookie", c.cookie)

	client := &http.Client{}
	return client.Do(request)
}

func (c *Client) GetCampaignID(creatorID string) (string, error) {
	campaignURL := creatorURL(creatorID)
	response, err := http.Get(campaignURL)
	if err != nil {
		return "", fmt.Errorf("failed to get campaign ID: %s", err)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read campaign ID response: %s", err)
	}

	matches := patreonCampaignRegex.FindStringSubmatch(string(body))
	if len(matches) == 0 {
		return "", fmt.Errorf("failed to find campaign ID")
	}

	return matches[1], nil
}

func (c *Client) GetPosts(campaignID string, cursor *string) (PostsResponse, error) {
	options := map[string]string{
		"include":                          "attachments,images,media",
		"fields[post]":                     "teaser_text,current_user_can_view,post_metadata,published_at,post_type,title,url,view_count",
		"fields[media]":                    "id,image_urls,download_url,metadata,mimetype,name,size_bytes",
		"filter[contains_exclusive_posts]": "true",
		"filter[is_draft]":                 "false",
		"sort":                             "-published_at",
		"json-api-version":                 "1.0",
		"filter[campaign_id]":              campaignID,
	}
	if cursor != nil {
		options["page[cursor]"] = *cursor
	}

	response, err := c.doAPIRequest("/posts", options)
	if err != nil {
		return PostsResponse{}, err
	}

	var postsResponse PostsResponse
	err = json.NewDecoder(response.Body).Decode(&postsResponse)
	if err != nil {
		return PostsResponse{}, err
	}

	return postsResponse, nil
}

func (c *Client) IsAuthenticated() (bool, error) {
	options := map[string]string{
		"include":          "active_memberships.campaign",
		"fields[campaign]": "avatar_photo_image_urls,name,published_at,url,vanity,is_nsfw,url_for_current_user",
		"fields[member]":   "is_free_member,is_free_trial",
		"json-api-version": "1.0",
	}

	response, err := c.doAPIRequest("/current_user", options)
	if err != nil {
		return false, err
	}

	var userErrorResponse UserErrorResponse
	err = json.NewDecoder(response.Body).Decode(&userErrorResponse)
	if err != nil {
		return false, err
	}

	if len(userErrorResponse.Errors) > 0 {
		return false, nil
	}

	return true, nil
}
