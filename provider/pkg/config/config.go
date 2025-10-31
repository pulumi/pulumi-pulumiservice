package config

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
	"github.com/pulumi/pulumi/sdk/v3/go/common/workspace"
)

const (
	EnvVarPulumiAccessToken = "PULUMI_ACCESS_TOKEN"
	EnvVarPulumiBackendUrl  = "PULUMI_BACKEND_URL"
)

func GetClient(ctx context.Context) Client {
	if v := ctx.Value(mockClientKey{}); v != nil {
		return v.(Client)
	}
	return infer.GetConfig[Config](ctx).client
}

type mockClientKey struct{}

// WithMockClient injects a client into the context for testing. This should only be used
// for testing.
func WithMockClient(ctx context.Context, client Client) context.Context {
	return context.WithValue(ctx, mockClientKey{}, client)
}

var ErrAccessTokenNotFound = fmt.Errorf("pulumi access token not found")

var (
	_ infer.CustomConfigure = &Config{}
)

// An interface to represent [*pulumiapi.Client] that remains mock-able.
//
// All client interfaces from [pulumiapi] should be added to this interface.
type Client interface {
	pulumiapi.AgentPoolClient
	pulumiapi.ApprovalRuleClient
	pulumiapi.DeploymentSettingsClient
	pulumiapi.EnvironmentScheduleClient
	pulumiapi.OidcClient
	pulumiapi.TeamClient
	pulumiapi.WebhookClient
}

type Config struct {
	AccessToken string `pulumi:"accessToken,optional" provider:"secret"`
	ApiURL      string `pulumi:"apiUrl,optional"`

	client *pulumiapi.Client
}

func (c *Config) Annotate(a infer.Annotator) {
	a.Describe(&c.AccessToken, "Access Token to authenticate with Pulumi Cloud.")
	a.Describe(&c.ApiURL, "Optional override of Pulumi Cloud API endpoint.")
	a.SetDefault(&c.ApiURL, "https://api.pulumi.com", EnvVarPulumiBackendUrl)
}

func (c *Config) Configure(context.Context) error {
	if c.AccessToken == "" {
		// If the env var is set, use that.
		c.AccessToken = os.Getenv(EnvVarPulumiAccessToken)
	}
	if c.AccessToken == "" {
		// attempt to grab credentials directly from the pulumi configuration on the machine
		creds, err := workspace.GetStoredCredentials()
		if err != nil {
			return ErrAccessTokenNotFound
		}
		c.AccessToken = creds.Current
	}

	if c.AccessToken == "" {
		return ErrAccessTokenNotFound
	}

	var err error
	c.client, err = pulumiapi.NewClient(&http.Client{
		Timeout: 60 * time.Second,
	}, c.AccessToken, c.ApiURL)
	return err
}
