package provider

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/pulumi/pulumi/sdk/v3/go/common/workspace"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/internal/pulumiapi"
)

const (
	EnvVarPulumiAccessToken = "PULUMI_ACCESS_TOKEN"
	EnvVarPulumiBackendUrl  = "PULUMI_BACKEND_URL"
)

var ErrAccessTokenNotFound = fmt.Errorf("pulumi access token not found")

type testClientKey struct{}

// TestClientKey is the key used during client lookups to override the normal key.
var TestClientKey = testClientKey{}

// GetClient gets the client for a particular service, or a test client if it is
// available.
func GetClient[T any](ctx context.Context) T {
	if v := ctx.Value(TestClientKey); v != nil {
		return v.(T)
	}

	return any(GetConfig(ctx).client).(T)
}

// GetConfig accesses the config associated with the current request.
func GetConfig(ctx context.Context) Config { return *infer.GetConfig[*Config](ctx) }

type Config struct {
	AccessToken string `pulumi:"accessToken,optional" provider:"secret"`
	ServiceURL  string `pulumi:"serviceURL,optional"`
	client      *pulumiapi.Client
}

var (
	_ infer.Annotated       = (*Config)(nil)
	_ infer.CustomConfigure = (*Config)(nil)
)

func (c *Config) Annotate(a infer.Annotator) {
	a.Describe(&c.AccessToken, "Access Token to authenticate with Pulumi Cloud.")
	a.SetDefault(&c.AccessToken, nil, EnvVarPulumiAccessToken)

	a.Describe(&c.ServiceURL, "The service URL used to reach Pulumi Cloud.")
	a.SetDefault(&c.ServiceURL, "https://api.pulumi.com", EnvVarPulumiBackendUrl)
}

func (c *Config) Configure(context.Context) error {
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

	c.client = client
	return err
}
