package authentication

import (
	"context"
	"fmt"

	rest "github.com/SKF/go-rest-utility/client"
	"github.com/SKF/go-utility/v2/stages"
)

type SSOClient struct {
	*rest.Client
}

func NewSSOClient(stage string) *SSOClient {
	baseURL := rest.WithBaseURL(fmt.Sprintf("https://sso-api.%s.users.enlight.skf.com", stage))
	if stage == stages.StageProd {
		baseURL = rest.WithBaseURL("https://sso-api.users.enlight.skf.com")
	}

	return &SSOClient{
		Client: rest.NewClient(baseURL),
	}
}

// nolint: unused
func (c *SSOClient) getUserIDFromAccessToken(ctx context.Context, accessToken string) (string, error) {
	request := rest.Get("users/me").
		SetHeader("Authorization", accessToken).
		SetHeader("Accept", "application/json")

	var response struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}

	if err := c.DoAndUnmarshal(ctx, request, &response); err != nil {
		return "", err
	}

	return response.Data.ID, nil
}
