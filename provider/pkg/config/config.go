package config

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"time"

	esc_client "github.com/pulumi/esc/cmd/esc/cli/client"
	"github.com/pulumi/esc/cmd/esc/cli/version"
	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/pulumi/pulumi/sdk/v3/go/common/workspace"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/rest"
)

const (
	EnvVarPulumiAccessToken = "PULUMI_ACCESS_TOKEN"
	EnvVarPulumiBackendURL  = "PULUMI_BACKEND_URL"
	// EnvVarPulumiAPI is an internal alias the Pulumi CLI sets on certain login
	// flows (e.g. `pl login devstack`). We honor it as a fallback so a token
	// scoped to a non-prod backend doesn't silently route to api.pulumi.com.
	EnvVarPulumiAPI = "PULUMI_API"

	defaultAPIURL = "https://api.pulumi.com"
)

func GetClient(ctx context.Context) Client {
	if v := ctx.Value(mockClientKey{}); v != nil {
		return v.(Client)
	}
	return infer.GetConfig[Config](ctx).client
}

// GetEscClient returns the upstream ESC client. Resources that operate on
// ESC-served endpoints (e.g. environment revision tags) use this in lieu of
// the pulumiapi.Client surface.
func GetEscClient(ctx context.Context) esc_client.Client {
	if v := ctx.Value(mockEscClientKey{}); v != nil {
		return v.(esc_client.Client)
	}
	return infer.GetConfig[Config](ctx).escClient
}

// GetTransport returns the transport the api resources send requests through,
// or nil before Configure has run.
func GetTransport(ctx context.Context) rest.Transport {
	return infer.GetConfig[Config](ctx).transport
}

type (
	mockClientKey    struct{}
	mockEscClientKey struct{}
)

// WithMockClient injects a client into the context for testing. This should only be used
// for testing.
func WithMockClient(ctx context.Context, client Client) context.Context {
	return context.WithValue(ctx, mockClientKey{}, client)
}

// WithMockEscClient injects an ESC client into the context for testing. This
// should only be used for testing.
func WithMockEscClient(ctx context.Context, client esc_client.Client) context.Context {
	return context.WithValue(ctx, mockEscClientKey{}, client)
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
	pulumiapi.EntityAssignmentClient
	pulumiapi.EnvironmentMetadataClient
	pulumiapi.EnvironmentScheduleClient
	pulumiapi.InsightsAccountClient
	pulumiapi.MemberClient
	pulumiapi.OidcClient
	pulumiapi.OrgAccessTokenClient
	pulumiapi.PolicyGroupClient
	pulumiapi.PolicyPackClient
	pulumiapi.RegistryPolicyPackClient
	pulumiapi.RoleClient
	pulumiapi.StackClient
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

	client    *pulumiapi.Client
	escClient esc_client.Client
	transport rest.Transport
}

func (c *Config) Annotate(a infer.Annotator) {
	a.Describe(&c.AccessToken, "Access token to authenticate with Pulumi Cloud. Optional: if unset, the provider "+
		"uses the `PULUMI_ACCESS_TOKEN` environment variable, then the access token of the account you're "+
		"logged in to with `pulumi login`.")
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

		c.AccessToken = creds.Accounts[creds.Current].AccessToken
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

	httpClient := &http.Client{
		Timeout: 60 * time.Second,
	}
	var err error
	c.client, err = pulumiapi.NewClient(httpClient, c.AccessToken, c.APIURL)
	if err != nil {
		return err
	}
	// pulumiapi.NewClient and the ESC client default an empty URL on their
	// own; the transport needs the default spelled out.
	baseURL := c.APIURL
	if baseURL == "" {
		baseURL = defaultAPIURL
	}
	c.transport = &authedTransport{
		baseURL: baseURL,
		token:   c.AccessToken,
		client:  httpClient,
	}
	c.escClient = esc_client.New(
		fmt.Sprintf("provider-pulumiservice/1 (%s; %s)", version.Version, runtime.GOOS),
		c.APIURL,
		c.AccessToken,
		false,
	)
	return nil
}

type authedTransport struct {
	baseURL string
	token   string
	client  *http.Client
}

func (t *authedTransport) Do(_ context.Context, req *http.Request) (*http.Response, error) {
	base, err := url.Parse(t.baseURL)
	if err != nil {
		return nil, fmt.Errorf("authedTransport: parse base URL %q: %w", t.baseURL, err)
	}

	req.URL.Scheme = base.Scheme
	req.URL.Host = base.Host
	req.Host = base.Host

	req.Header.Set("Authorization", "token "+t.token)
	// v0 (legacy) callers leave Accept unset and want the Pulumi-versioned
	// media type; api (rest.Resource) sets Accept: application/json explicitly
	// because its OpenAPI-described endpoints return standard JSON.
	if req.Header.Get("Accept") == "" {
		req.Header.Set("Accept", pulumiapi.AcceptMediaType)
	}
	req.Header.Set("X-Pulumi-Source", "provider")

	// Request host is pinned to the operator-configured Pulumi API base URL
	// (authedTransport.baseURL), not attacker-controlled per-request input.
	return t.client.Do(req) //nolint:gosec // G704: host pinned to configured API base URL, not request input
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
