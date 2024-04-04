package provider

import (
	"fmt"
	"net/http"
	"time"

	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/pulumi/pulumi/sdk/v3/go/common/workspace"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/internal/pulumiapi"
)

const (
	EnvVarPulumiAccessToken = "PULUMI_ACCESS_TOKEN"
	EnvVarPulumiBackendUrl  = "PULUMI_BACKEND_URL"
)

var ErrAccessTokenNotFound = fmt.Errorf("pulumi access token not found")

// GetConfig accesses the config associated with the current request.
func GetConfig(ctx p.Context) Config { return infer.GetConfig[Config](ctx) }

type Config struct {
	AccessToken string `pulumi:"accessToken,optional"`
	ServiceURL  string `pulumi:"serviceURL,optional"`
	Client      pulumiapi.TeamClient
}

var (
	_ infer.Annotated       = (*Config)(nil)
	_ infer.CustomConfigure = (*Config)(nil)
)

func (c *Config) Annotate(a infer.Annotator) {
	a.Describe(&c.AccessToken, "Access Token to authenticate with Pulumi Cloud.")
	a.SetDefault(nil, EnvVarPulumiAccessToken)

	a.Describe(&c.ServiceURL, "The service URL used to reach Pulumi Cloud.")
	a.SetDefault("https://api.pulumi.com", EnvVarPulumiBackendUrl)
}

func (c *Config) Configure(p.Context) error {
	// Ensure that we have an access token
	if len(c.AccessToken) == 0 {
		creds, err := workspace.GetStoredCredentials()
		if err != nil {
			return ErrAccessTokenNotFound
		}
		if token, ok := creds.AccessTokens[creds.Current]; ok {
			c.AccessToken = token
		} else {
			return ErrAccessTokenNotFound
		}
	}

	// Construct the PulumiService client
	client, err := pulumiapi.NewClient(&http.Client{
		Timeout: 60 * time.Second,
	}, c.AccessToken, c.ServiceURL)

	c.Client = client
	return err
}
