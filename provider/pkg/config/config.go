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

func GetClient(ctx context.Context) *pulumiapi.Client { return infer.GetConfig[*Config](ctx).client }

var ErrAccessTokenNotFound = fmt.Errorf("pulumi access token not found")

var (
	_ infer.CustomConfigure = &Config{}
)

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

	httpClient := http.Client{
		Timeout: 60 * time.Second,
	}
	var err error
	c.client, err = pulumiapi.NewClient(&httpClient, c.AccessToken, c.ApiURL)
	return err
}
