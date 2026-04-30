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
	pulumiapi.AgentPoolClient
	pulumiapi.ApprovalRuleClient
	pulumiapi.DeploymentSettingsClient
	pulumiapi.EnvironmentMetadataClient
	pulumiapi.EnvironmentScheduleClient
	pulumiapi.InsightsAccountClient
	pulumiapi.MemberClient
	pulumiapi.OidcClient
	pulumiapi.RoleClient
	pulumiapi.StackTagClient
	pulumiapi.TeamClient
	pulumiapi.TeamRoleClient
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
	a.SetDefault(&c.APIURL, "https://api.pulumi.com", EnvVarPulumiBackendURL)
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

	// Fall back to PULUMI_BACKEND_URL when the engine hands us an empty
	// apiUrl. `pulumi import` persists the default-provider's config to
	// state with empty `inputs: {}` — accessToken survives the round trip
	// because of the manual env-var fallback above, but apiUrl had no
	// symmetric fallback. On destroy of an imported resource against a
	// non-prod Pulumi Cloud (review stack, self-hosted), c.APIURL would
	// reach this point empty, NewClient would default to
	// https://api.pulumi.com, the DELETE would dial production, and a
	// non-prod token would get rejected with
	//   401 API error: Unauthorized: No credentials provided or are invalid
	// even though up/import worked moments earlier in the same shell.
	//
	// `pulumi up` is unaffected — it captures the resolved provider
	// config (including SetDefault-applied apiUrl) and persists that.
	// Only the import path's empty-inputs state needs this fallback.
	if c.APIURL == "" {
		c.APIURL = os.Getenv(EnvVarPulumiBackendURL)
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
