package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

const apiURL = "https://www.patreon.com/api"

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
	currentUser, err := c.GetCurrentUser()
	if err != nil {
		return "", err
	}

	for _, include := range currentUser.Included {
		switch include := include.(type) {
		case ResponseCampaign:
			if strings.EqualFold(include.Attributes.Vanity, creatorID) {
				return include.ID, nil
			}
		}
	}

	return "", fmt.Errorf("failed to find campaign ID")
}

func (c *Client) GetCurrentUser() (UserResponse, error) {
	options := map[string]string{
		"include":          "active_memberships.campaign",
		"fields[campaign]": "name,published_at,url,vanity",
		"json-api-version": "1.0",
	}
	response, err := c.doAPIRequest("/current_user", options)
	if err != nil {
		return UserResponse{}, err
	}
	defer response.Body.Close()

	var userResponse UserResponse
	err = UnmarshalResponse(response.Body, &userResponse)
	if err != nil {
		return UserResponse{}, err
	}

	return userResponse, nil
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
	defer response.Body.Close()

	var postsResponse PostsResponse
	err = UnmarshalResponse(response.Body, &postsResponse)
	if err != nil {
		return PostsResponse{}, err
	}

	return postsResponse, nil
}

func (c *Client) IsAuthenticated() (bool, error) {
	response, err := c.doAPIRequest("/current_user", nil)
	if err != nil {
		return false, fmt.Errorf("failed to get current user: %w", err)
	}
	defer response.Body.Close()

	var errorResponse ErrorResponse
	err = json.NewDecoder(response.Body).Decode(&errorResponse)
	if err != nil {
		return false, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(errorResponse.Errors) > 0 {
		return false, nil
	}

	return true, nil
}
