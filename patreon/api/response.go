package api

type ResponseReference struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

type ResponseEntity[TAttributes any, TRelationShips any] struct {
	ID            string         `json:"id"`
	Type          string         `json:"type"`
	Attributes    TAttributes    `json:"attributes"`
	RelationShips TRelationShips `json:"relationships"`
}

type Response[TData any] struct {
	Data     TData         `json:"data"`
	Included []any         `json:"included,omitempty"`
	Links    ResponseLinks `json:"links,omitempty"`
	Meta     ResponseMeta  `json:"meta,omitempty"`
}

type ResponseLinks struct {
	Next    string `json:"next,omitempty"`
	Related string `json:"related,omitempty"`
	Self    string `json:"self,omitempty"`
}

type ResponseMeta struct {
	Pagination ResponsePagination `json:"pagination"`
}

type ResponsePagination struct {
	Total   int             `json:"total"`
	Cursors ResponseCursors `json:"cursors"`
}

type ResponseCursors struct {
	Next string `json:"next"`
}

type ResponseMember = ResponseEntity[ResponseMemberAttributes, ResponseMemberRelationships]

type ResponseMemberAttributes struct {
	PatronStatus                 string `json:"patron_status"`
	IsFollower                   bool   `json:"is_follower"`
	IsFreeMember                 bool   `json:"is_free_member"`
	IsFreeTrial                  bool   `json:"is_free_trial"`
	FullName                     string `json:"full_name"`
	PledgeAmountCents            int    `json:"pledge_amount_cents"`
	PledgeRelationshipStart      string `json:"pledge_relationship_start"`
	LifetimeSupportCents         int    `json:"lifetime_support_cents"`
	LastChargeDate               string `json:"last_charge_date"`
	LastChargeStatus             string `json:"last_charge_status"`
	CampaignCurrency             string `json:"campaign_currency"`
	CampaignPledgeAmountCents    int    `json:"campaign_pledge_amount_cents"`
	CampaignLifetimeSupportCents int    `json:"campaign_lifetime_support_cents"`
	AccessExpiresAt              string `json:"access_expires_at"`
}

type ResponseMemberRelationships struct {
	Campaign Response[ResponseReference] `json:"campaign"`
}

type ResponseCampaign = ResponseEntity[ResponseCampaignAttributes, ResponseCampaignRelationships]

type ResponseCampaignAttributes struct {
	Name        string `json:"name"`
	PublishedAt string `json:"published_at"`
	URL         string `json:"url"`
	Vanity      string `json:"vanity"`
}

type ResponseCampaignRelationships struct {
	Creator Response[ResponseReference]   `json:"creator"`
	Rewards Response[[]ResponseReference] `json:"rewards"`
}

type ResponseUser = ResponseEntity[ResponseUserAttributes, ResponseUserRelationShips]

type ResponseUserAttributes struct {
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	FullName  string `json:"full_name"`
	About     string `json:"about"`
	URL       string `json:"url"`
	ImageURL  string `json:"image_url"`
}

type ResponseUserRelationShips struct {
	Campaign Response[ResponseReference] `json:"campaign"`
}

type ResponseReward = ResponseEntity[ResponseRewardAttributes, ResponseRewardRelationships]

type ResponseRewardAttributes struct {
	Amount              int    `json:"amount"`
	AmountCents         int    `json:"amount_cents"`
	Remaining           int    `json:"remaining"`
	Description         string `json:"description"`
	RequiresShipping    bool   `json:"requires_shipping"`
	CreatedAt           string `json:"created_at"`
	URL                 string `json:"url"`
	DeclinedPatronCount int    `json:"declined_patron_count"`
	PostCount           int    `json:"post_count"`
	DiscordRoleIds      any    `json:"discord_role_ids"`
	Title               string `json:"title"`
	ImageURL            string `json:"image_url"`
	EditedAt            string `json:"edited_at"`
	Published           bool   `json:"published"`
	PublishedAt         string `json:"published_at"`
	UnpublishedAt       any    `json:"unpublished_at"`
	Currency            string `json:"currency"`
	PatronAmountCents   int    `json:"patron_amount_cents"`
	PatronCurrency      string `json:"patron_currency"`
	IsFreeTier          bool   `json:"is_free_tier"`
}

type ResponseRewardRelationships struct {
	Campaign Response[ResponseReference] `json:"campaign"`
}

type ResponsePost = ResponseEntity[ResponsePostAttributes, ResponsePostRelationships]

type ResponsePostAttributes struct {
	PostType           string               `json:"post_type"`
	Title              string               `json:"title"`
	PublishedAt        string               `json:"published_at"`
	URL                string               `json:"url"`
	CurrentUserCanView bool                 `json:"current_user_can_view"`
	TeaserText         string               `json:"teaser_text"`
	PostMetaData       ResponsePostMetaData `json:"post_metadata"`
}

type ResponsePostMetaData struct {
	ImageOrder []string `json:"image_order"`
}

type ResponsePostRelationships struct {
	Attachments ResponsePostRelationshipsAttachments `json:"attachments"`
	Images      ResponsePostRelationshipsImages      `json:"images"`
	Media       ResponsePostRelationshipsMedia       `json:"media"`
}

type ResponsePostRelationshipsAttachments struct {
	Data []any `json:"data"`
}

type ResponsePostRelationshipsImages struct {
	Data []ResponseReference `json:"data"`
}

type ResponsePostRelationshipsMedia struct {
	Data []ResponseReference `json:"data"`
}

type ResponseMedia = ResponseEntity[ResponseMediaAttributes, any]

type ResponseMediaAttributes struct {
	SizeBytes   int                    `json:"size_bytes"`
	MimeType    string                 `json:"mimetype"`
	DownloadURL string                 `json:"download_url"`
	ImageURLs   ResponseMediaImageURLs `json:"image_urls"`
	Metadata    ResponseMediaMetadata  `json:"metadata"`
}

type ResponseMediaImageURLs struct {
	URL                 string `json:"url"`
	Original            string `json:"original"`
	Default             string `json:"default"`
	DefaultBlurred      string `json:"default_blurred"`
	DefaultSmall        string `json:"default_small"`
	DefaultLarge        string `json:"default_large"`
	DefaultBlurredSmall string `json:"default_blurred_small"`
	Thumbnail           string `json:"thumbnail"`
	ThumbnailLarge      string `json:"thumbnail_large"`
	ThumbnailSmall      string `json:"thumbnail_small"`
}

type ResponseMediaMetadata struct {
	Dimensions ResponseMediaDimensions `json:"dimensions"`
}

type ResponseMediaDimensions struct {
	W int `json:"w"`
	H int `json:"h"`
}
