package config

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/pulumi/pulumi/sdk/v3/go/common/workspace"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
)

const (
	EnvVarPulumiAccessToken = "PULUMI_ACCESS_TOKEN"
	EnvVarPulumiBackendURL  = "PULUMI_BACKEND_URL"
	// EnvVarPulumiAPI is an internal alias the Pulumi CLI sets on certain login
	// flows (e.g. `pl login devstack`). We honor it as a fallback so a token
	// scoped to a non-prod backend doesn't silently route to api.pulumi.com.
	EnvVarPulumiAPI = "PULUMI_API"
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
	_ infer.CustomConfigure              = &Config{}
	_ infer.CustomDiff[*Config, *Config] = &Config{}
)

// An interface to represent [*pulumiapi.Client] that remains mock-able.
//
// All client interfaces from [pulumiapi] should be added to this interface.
type Client interface {
	pulumiapi.AccessTokenClient
	pulumiapi.AgentPoolClient
	pulumiapi.ApprovalRuleClient
	pulumiapi.DeploymentSettingsClient
	pulumiapi.EnvironmentMetadataClient
	pulumiapi.EnvironmentScheduleClient
	pulumiapi.InsightsAccountClient
	pulumiapi.MemberClient
	pulumiapi.OidcClient
	pulumiapi.OrgAccessTokenClient
	pulumiapi.RoleClient
	pulumiapi.StackScheduleClient
	pulumiapi.StackTagClient
	pulumiapi.TeamAccessTokenClient
	pulumiapi.TeamClient
	pulumiapi.TeamRoleClient
	pulumiapi.TemplateSourceClient
	pulumiapi.UserClient
	pulumiapi.WebhookClient
}

type Config struct {
	AccessToken string `pulumi:"accessToken,optional" provider:"secret"`
	APIURL      string `pulumi:"apiUrl,optional"`

	client *pulumiapi.Client
}

func (c *Config) Annotate(a infer.Annotator) {
	a.Describe(&c.AccessToken, "Access Token to authenticate with Pulumi Cloud.")
	a.Describe(&c.APIURL, "Optional override of Pulumi Cloud API endpoint.")
	a.SetDefault(&c.APIURL, "https://api.pulumi.com", EnvVarPulumiBackendURL, EnvVarPulumiAPI)
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

		c.AccessToken = creds.AccessTokens[creds.Current]
	}

	if c.AccessToken == "" {
		return ErrAccessTokenNotFound
	}

	// SetDefault on APIURL applies during CheckConfig, not Configure;
	// `pulumi import` persists default-provider state with empty inputs,
	// and on subsequent destroy the engine calls Configure with apiUrl=""
	// without re-running Check. Without this fallback NewClient would
	// default to api.pulumi.com and 401 against a non-prod token
	// (regression caught by TestYamlRbacComposeImport).
	if c.APIURL == "" {
		c.APIURL = os.Getenv(EnvVarPulumiBackendURL)
	}
	if c.APIURL == "" {
		c.APIURL = os.Getenv(EnvVarPulumiAPI)
	}

	var err error
	c.client, err = pulumiapi.NewClient(&http.Client{
		Timeout: 60 * time.Second,
	}, c.AccessToken, c.APIURL)
	return err
}

func (*Config) Diff(_ context.Context, req infer.DiffRequest[*Config, *Config]) (infer.DiffResponse, error) {
	kind := func(input, state string) p.DiffKind {
		switch {
		case input == "" && state != "":
			return p.Delete
		case input != "" && state == "":
			return p.Add
		case input != state:
			return p.Update
		default:
			return p.Stable
		}
	}
	var hasChanges bool
	detailedDiff := map[string]p.PropertyDiff{}

	if req.Inputs.AccessToken != req.State.AccessToken {
		hasChanges = true
		detailedDiff["accessToken"] = p.PropertyDiff{
			Kind:      kind(req.Inputs.AccessToken, req.State.AccessToken),
			InputDiff: true,
		}
	}

	if req.Inputs.APIURL != req.State.APIURL {
		hasChanges = true
		detailedDiff["apiUrl"] = p.PropertyDiff{
			Kind:      kind(req.Inputs.APIURL, req.State.APIURL),
			InputDiff: true,
		}
	}
	return infer.DiffResponse{
		HasChanges:   hasChanges,
		DetailedDiff: detailedDiff,
	}, nil
}
