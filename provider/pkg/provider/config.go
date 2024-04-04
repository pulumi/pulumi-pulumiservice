package provider

import (
	"fmt"

	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/pulumi/pulumi/sdk/v3/go/common/workspace"
)

const (
	EnvVarPulumiAccessToken = "PULUMI_ACCESS_TOKEN"
	EnvVarPulumiBackendUrl  = "PULUMI_BACKEND_URL"
)

var ErrAccessTokenNotFound = fmt.Errorf("pulumi access token not found")

type Config struct {
	AccessToken string `pulumi:"accessToken,optional"`
	ServiceURL  string `pulumi:"serviceURL,optional"`
}

var _ infer.Annotated = (*Config)(nil)

func (c *Config) Annotate(a infer.Annotator) {
	a.Describe(&c.AccessToken, "Access Token to authenticate with Pulumi Cloud.")
	a.SetDefault(nil, EnvVarPulumiAccessToken)

	a.Describe(&c.ServiceURL, "The service URL used to reach Pulumi Cloud.")
	a.SetDefault("https://api.pulumi.com", EnvVarPulumiBackendUrl)
}

func (c *Config) getPulumiAccessToken() (string, error) {
	if len(c.AccessToken) > 0 {
		// found the token
		return c.AccessToken, nil
	}

	// attempt to grab credentials directly from the pulumi configuration on the machine
	creds, err := workspace.GetStoredCredentials()
	if err != nil {
		return "", ErrAccessTokenNotFound
	}
	if token, ok := creds.AccessTokens[creds.Current]; ok {
		return token, nil
	}
	return "", ErrAccessTokenNotFound
}
