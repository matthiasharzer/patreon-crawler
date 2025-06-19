package api

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testCase struct {
	data   string
	entity any
}

var entity = struct {
	media    testCase
	member   testCase
	user     testCase
	post     testCase
	reward   testCase
	campaign testCase
}{
	media: testCase{
		data: `{
					"type": "media",
					"id": "some-id",
					"attributes": {
						"size_bytes": 123456,
						"mimetype": "image/png",
						"download_url": "https://example.com/download",
						"image_urls": {
							"url": "https://example.com/image.png",
							"original": "https://example.com/original.png",
							"default": "https://example.com/default.png",
							"default_blurred": "https://example.com/default_blurred.png",
							"default_small": "https://example.com/default_small.png",
							"default_large": "https://example.com/default_large.png",
							"default_blurred_small": "https://example.com/default_blurred_small.png",
							"thumbnail": "https://example.com/thumbnail.png",
							"thumbnail_large": "https://example.com/thumbnail_large.png",
							"thumbnail_small": "https://example.com/thumbnail_small.png"
						},
						"metadata": {
							"dimensions": {
								"w": 800,
								"h": 600
							}
						}
					}
				}`,
		entity: ResponseMedia{

			ID:   "some-id",
			Type: "media",
			Attributes: ResponseMediaAttributes{
				SizeBytes:   123456,
				MimeType:    "image/png",
				DownloadURL: "https://example.com/download",
				ImageURLs: ResponseMediaImageURLs{
					URL:                 "https://example.com/image.png",
					Original:            "https://example.com/original.png",
					Default:             "https://example.com/default.png",
					DefaultBlurred:      "https://example.com/default_blurred.png",
					DefaultSmall:        "https://example.com/default_small.png",
					DefaultLarge:        "https://example.com/default_large.png",
					DefaultBlurredSmall: "https://example.com/default_blurred_small.png",
					Thumbnail:           "https://example.com/thumbnail.png",
					ThumbnailLarge:      "https://example.com/thumbnail_large.png",
					ThumbnailSmall:      "https://example.com/thumbnail_small.png",
				},
				Metadata: ResponseMediaMetadata{
					Dimensions: ResponseMediaDimensions{
						W: 800,
						H: 600,
					},
				},
			},
		},
	},
	member: testCase{
		data: `{
					"type": "member",
					"id": "member-id",
					"attributes": {
						"patron_status": "active_patron",
						"is_follower": true,
						"if_free_member": false,
						"is_free_trial": false,
						"full_name": "John Doe",
						"pledge_amount_cents": 5000,
						"pledge_relationship_start": "2025-01-01T00:00:00Z",
						"lifetime_support_cents": 10000,
						"last_charge_date": "2025-01-15T00:00:00Z",
						"last_charge_status": "Paid",
						"campaign_currency": "USD",
						"campaign_pledge_amount_cents": 5000,
						"campaign_lifetime_support_cents": 10000,
						"access_expires_at": "2025-12-31T23:59:59Z"
					},
					"relationships": {
						"campaign": {
							"data": {
								"type": "campaign",
								"id": "campaign-id"
							}
						}
					}
				}`,
		entity: ResponseMember{
			ID:   "member-id",
			Type: "member",
			Attributes: ResponseMemberAttributes{
				PatronStatus:                 "active_patron",
				IsFollower:                   true,
				IsFreeMember:                 false,
				IsFreeTrial:                  false,
				FullName:                     "John Doe",
				PledgeAmountCents:            5000,
				PledgeRelationshipStart:      "2025-01-01T00:00:00Z",
				LifetimeSupportCents:         10000,
				LastChargeDate:               "2025-01-15T00:00:00Z",
				LastChargeStatus:             "Paid",
				CampaignCurrency:             "USD",
				CampaignPledgeAmountCents:    5000,
				CampaignLifetimeSupportCents: 10000,
				AccessExpiresAt:              "2025-12-31T23:59:59Z",
			},
			RelationShips: ResponseMemberRelationships{
				Campaign: Response[ResponseReference]{
					Data: ResponseReference{
						ID:   "campaign-id",
						Type: "campaign",
					},
				},
			},
		},
	},
	user: testCase{
		data: `{
					"type": "user",
					"id": "user-id",
					"attributes": {
						"email": "john@doe.com",
						"first_name": "John",
						"last_name": "Doe",
						"full_name": "John Doe",
						"about": "This is a test user.",
						"url": "https://example.com/john_doe",
						"image_url": "https://example.com/john_doe.png"
					},
					"relationships": {
						"campaign": {
							"data": {
								"type": "campaign",
								"id": "campaign-id"
							},
							"links": {
								"related": "https://example.com/campaigns/campaign-id"
							}
						}
					}
				}`,
		entity: ResponseUser{
			ID:   "user-id",
			Type: "user",
			Attributes: ResponseUserAttributes{
				Email:     "john@doe.com",
				FirstName: "John",
				LastName:  "Doe",
				FullName:  "John Doe",
				About:     "This is a test user.",
				URL:       "https://example.com/john_doe",
				ImageURL:  "https://example.com/john_doe.png",
			},
			RelationShips: ResponseUserRelationShips{
				Campaign: Response[ResponseReference]{
					Data: ResponseReference{
						ID:   "campaign-id",
						Type: "campaign",
					},
					Links: ResponseLinks{
						Related: "https://example.com/campaigns/campaign-id",
					},
				},
			},
		},
	},
	post: testCase{
		data: `{
					"type": "post",
					"id": "post-id",
					"attributes": {
						"post_type": "text",
						"title": "Test Post",
						"published_at": "2025-01-01T00:00:00Z",
						"url": "https://example.com/posts/post-id",
						"current_user_can_view": true,
						"teaser_text": "This is a teaser for the test post.",
						"post_metadata": {
							"image_order": ["media-id-1", "media-id-2"]
						}
					},
					"relationships": {
						"attachments": {
							"data": []
						},
						"images": {
							"data": [{
								"type": "media",
								"id": "media-id-1"
							}, {
								"type": "media",
								"id": "media-id-2"
							}]
						},
						"media": {
							"data": [{
								"type": "media",
								"id": "media-id-1"
							}, {
								"type": "media",
								"id": "media-id-2"
							}]
						}
					}
				}`,
		entity: ResponsePost{
			ID:   "post-id",
			Type: "post",
			Attributes: ResponsePostAttributes{
				PostType:           "text",
				Title:              "Test Post",
				PublishedAt:        "2025-01-01T00:00:00Z",
				URL:                "https://example.com/posts/post-id",
				CurrentUserCanView: true,
				TeaserText:         "This is a teaser for the test post.",
				PostMetaData: ResponsePostMetaData{
					ImageOrder: []string{"media-id-1", "media-id-2"},
				},
			},
			RelationShips: ResponsePostRelationships{
				Attachments: ResponsePostRelationshipsAttachments{
					Data: []any{},
				},
				Images: ResponsePostRelationshipsImages{
					Data: []ResponseReference{
						{Type: "media", ID: "media-id-1"},
						{Type: "media", ID: "media-id-2"},
					},
				},
				Media: ResponsePostRelationshipsMedia{
					Data: []ResponseReference{
						{Type: "media", ID: "media-id-1"},
						{Type: "media", ID: "media-id-2"},
					},
				},
			},
		},
	},
	reward: testCase{
		data: `{
					"type": "reward",
					"id": "reward-id",
					"attributes": {
						"amount": 1000,
						"amount_cents": 1000,
						"remaining": 10,
						"description": "This is a test reward.",
						"requires_shipping": false,
						"created_at": "2025-01-01T00:00:00Z",
						"url": "https://example.com/rewards/reward-id",
						"declined_patron_count": 5,
						"post_count": 2,
						"discord_role_ids": null,
						"title": "Test Reward",
						"image_url": "https://example.com/rewards/reward-id.png",
						"edited_at": "2025-01-02T00:00:00Z",
						"published": true,
						"published_at": "2025-01-01T00:00:00Z",
						"unpublished_at": null,
						"currency": "USD",
						"patron_amount_cents": 1000,
						"patron_currency": "USD",
						"is_free_tier": false
					},
					"relationships": {
						"campaign": {
							"data": {
								"type": "campaign",
								"id": "campaign-id"
							}
						}
					}
				}`,
		entity: ResponseReward{
			ID:   "reward-id",
			Type: "reward",
			Attributes: ResponseRewardAttributes{
				Amount:              1000,
				AmountCents:         1000,
				Remaining:           10,
				Description:         "This is a test reward.",
				RequiresShipping:    false,
				CreatedAt:           "2025-01-01T00:00:00Z",
				URL:                 "https://example.com/rewards/reward-id",
				DeclinedPatronCount: 5,
				PostCount:           2,
				DiscordRoleIds:      nil,
				Title:               "Test Reward",
				ImageURL:            "https://example.com/rewards/reward-id.png",
				EditedAt:            "2025-01-02T00:00:00Z",
				Published:           true,
				PublishedAt:         "2025-01-01T00:00:00Z",
				UnpublishedAt:       nil,
				Currency:            "USD",
				PatronAmountCents:   1000,
				PatronCurrency:      "USD",
				IsFreeTier:          false,
			},
			RelationShips: ResponseRewardRelationships{
				Campaign: Response[ResponseReference]{
					Data: ResponseReference{
						ID:   "campaign-id",
						Type: "campaign",
					},
				},
			},
		},
	},
	campaign: testCase{
		data: `{
					"type": "campaign",
					"id": "campaign-id",
					"attributes": {
						"name": "Test Campaign",
						"published_at": "2025-01-01T00:00:00Z",
						"url": "https://example.com/campaigns/campaign-id",
						"vanity": "test-campaign"
					},
					"relationships": {
						"creator": {
							"data": {
								"type": "user",
								"id": "user-id"
							}
						},
						"rewards": {
							"data": [{
								"type": "reward",
								"id": "reward-id-1"
							}, {
								"type": "reward",
								"id": "reward-id-2"
							}]
						}
					}
				}`,
		entity: ResponseCampaign{
			ID:   "campaign-id",
			Type: "campaign",
			Attributes: ResponseCampaignAttributes{
				Name:        "Test Campaign",
				PublishedAt: "2025-01-01T00:00:00Z",
				URL:         "https://example.com/campaigns/campaign-id",
				Vanity:      "test-campaign",
			},
			RelationShips: ResponseCampaignRelationships{
				Creator: Response[ResponseReference]{
					Data: ResponseReference{
						ID:   "user-id",
						Type: "user",
					},
				},
				Rewards: Response[[]ResponseReference]{
					Data: []ResponseReference{
						{Type: "reward", ID: "reward-id-1"},
						{Type: "reward", ID: "reward-id-2"},
					},
				},
			},
		},
	},
}

func TestUnmarshalEntity(t *testing.T) {
	t.Run("fails if unknown entity type is given", func(t *testing.T) {
		entityBytes := []byte(`{"type": "unknown"}`)
		_, err := UnmarshalEntity(entityBytes)
		assert.ErrorContains(t, err, "unknown entity type")
	})

	t.Run("fails if entity bytes are not valid JSON", func(t *testing.T) {
		malformedEntityBytes := []byte(`{"type": "`)
		_, err := UnmarshalEntity(malformedEntityBytes)
		assert.Error(t, err)
	})

	t.Run("succeeds with expected entity", func(t *testing.T) {
		tests := []struct {
			name     string
			testCase testCase
		}{
			{
				name:     "media",
				testCase: entity.media,
			},
			{
				name:     "member",
				testCase: entity.member,
			},
			{
				name:     "user",
				testCase: entity.user,
			},
			{
				name:     "post",
				testCase: entity.post,
			},
			{
				name:     "reward",
				testCase: entity.reward,
			},
			{
				name:     "campaign",
				testCase: entity.campaign,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				entityBytes := []byte(tt.testCase.data)
				entity, err := UnmarshalEntity(entityBytes)
				require.NoError(t, err)
				assert.Equal(t, tt.testCase.entity, entity)
			})
		}
	})
}

func TestUnmarshalResponse(t *testing.T) {
	t.Run("unmarshals users response with data", func(t *testing.T) {
		responseData := `{
			"data": {
				"type": "user",
				"id": "user-id",
				"attributes": {
					"email": "john@doe.com",
					"first_name": "John",
					"last_name": "Doe",
					"full_name": "John Doe",
					"about": "This is a test user.",
					"url": "https://example.com/john_doe",
					"image_url": "https://example.com/john_doe.png"
				}
			},
			"included": [
` + entity.media.data + `,
` + entity.member.data + `,
` + entity.user.data + `,
` + entity.post.data + `,
` + entity.reward.data + `,
` + entity.campaign.data + `
			]
		}`
		var response Response[ResponseUser]
		err := UnmarshalResponse(strings.NewReader(responseData), &response)
		require.NoError(t, err)

		assert.Equal(t, "user-id", response.Data.ID)
		assert.Equal(t, "user", response.Data.Type)
		assert.Equal(t, "john@doe.com", response.Data.Attributes.Email)
		assert.Equal(t, "John", response.Data.Attributes.FirstName)
		assert.Equal(t, "Doe", response.Data.Attributes.LastName)
		assert.Equal(t, "John Doe", response.Data.Attributes.FullName)
		assert.Equal(t, "This is a test user.", response.Data.Attributes.About)
		assert.Equal(t, "https://example.com/john_doe", response.Data.Attributes.URL)
		assert.Equal(t, "https://example.com/john_doe.png", response.Data.Attributes.ImageURL)

		require.Len(t, response.Included, 6)
		assert.Equal(t, entity.media.entity, response.Included[0])
		assert.Equal(t, entity.member.entity, response.Included[1])
		assert.Equal(t, entity.user.entity, response.Included[2])
		assert.Equal(t, entity.post.entity, response.Included[3])
		assert.Equal(t, entity.reward.entity, response.Included[4])
		assert.Equal(t, entity.campaign.entity, response.Included[5])
	})
}
