package authentication

import (
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
