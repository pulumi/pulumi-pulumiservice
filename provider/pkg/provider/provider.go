// Copyright 2016-2020, Pulumi Corporation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync/atomic"
	"time"

	_ "embed" // For manualSchema.

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	pbempty "google.golang.org/protobuf/types/known/emptypb"

	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
	mw "github.com/pulumi/pulumi-go-provider/middleware"
	ctxmw "github.com/pulumi/pulumi-go-provider/middleware/context"
	"github.com/pulumi/pulumi-go-provider/middleware/dispatch"
	"github.com/pulumi/pulumi-go-provider/middleware/rpc"
	"github.com/pulumi/pulumi/pkg/v3/codegen/schema"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	"github.com/pulumi/pulumi/sdk/v3/go/common/tokens"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/provider"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/cloud"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/config"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/functions"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/resources"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/rest"
)

const (
	respectSchemaVersionKey = "respectSchemaVersion"
	readmeKey               = "readme"
	versionKey              = "version"
)

//go:embed README.md
var readme string

type PulumiServiceResource interface {
	Diff(req *pulumirpc.DiffRequest) (*pulumirpc.DiffResponse, error)
	Create(req *pulumirpc.CreateRequest) (*pulumirpc.CreateResponse, error)
	Delete(req *pulumirpc.DeleteRequest) (*pbempty.Empty, error)
	Check(req *pulumirpc.CheckRequest) (*pulumirpc.CheckResponse, error)
	Update(req *pulumirpc.UpdateRequest) (*pulumirpc.UpdateResponse, error)
	Read(req *pulumirpc.ReadRequest) (*pulumirpc.ReadResponse, error)
	Name() string
}

type pulumiserviceProvider struct {
	pulumirpc.UnimplementedResourceProviderServer

	host            *provider.HostClient
	name            string
	version         string
	schema          string
	pulumiResources []PulumiServiceResource
	AccessToken     string
	// transportRef is the per-provider api transport, populated during
	// Configure and consumed by the ctxmw.Wrap layer to attach this
	// provider's transport to every CRUD context — keeps multi-provider
	// instances from racing on a package-global.
	transportRef *atomic.Value
}

// embed manual-schema.json directly into resource binary so that we can properly serve the schema
// directly from the resource provider.
//
//go:embed manual-schema.json
var manualSchema string

// MakeProvider builds the unified Pulumi Cloud Provider. Three layers:
//
//  1. legacyRaw — the existing custom gRPC server (pulumiserviceProvider)
//     handling resources defined in manual-schema.json.
//  2. dispatch.Wrap — overlays metadata-driven api resources from
//     provider/pkg/cloud/metadata.json. Schema for these is spliced into
//     GetSchema responses by withCloudApiSchema.
//  3. infer.NewProviderBuilder — adds modern infer-style resources
//     (Team, OrganizationRole, etc.) at pulumiservice:index:* and stamps
//     in package-level metadata (display name, language map, config schema).
//
// Existing user code keeps working unchanged: pulumiservice:index:* tokens
// resolve through layers 1 and 3; the new api resources at
// pulumiservice:api:* resolve through layer 2.
func MakeProvider(host *provider.HostClient, name, version string) (pulumirpc.ResourceProviderServer, error) {
	transportRef := &atomic.Value{}
	legacyRaw := rpc.Provider(&pulumiserviceProvider{
		host:         host,
		name:         name,
		schema:       mustSetSchemaVersion(manualSchema, version),
		version:      version,
		transportRef: transportRef,
	})

	customs := map[tokens.Type]mw.CustomResource{}
	for tok, h := range rest.Resources(cloud.Spec(), cloud.Metadata()) {
		customs[tokens.Type(tok)] = h
	}
	composed := dispatch.Wrap(legacyRaw, dispatch.Options{Customs: customs})
	composed = withCloudApiSchema(composed, cloud.Spec(), cloud.Metadata(), name)
	// Attach this provider's transport to every CRUD context. Pairs with
	// pulumiserviceProvider.Configure storing into transportRef.
	composed = ctxmw.Wrap(composed, func(ctx context.Context) context.Context {
		if v := transportRef.Load(); v != nil {
			return rest.WithTransport(ctx, v.(rest.Transport))
		}
		return ctx
	})

	provider, err := infer.NewProviderBuilder().
		WithDisplayName("Pulumi Cloud").
		WithNamespace("pulumi").
		WithWrapped(composed).
		WithResources(
			infer.Resource(&resources.AccessToken{}),
			infer.Resource(&resources.AgentPool{}),
			infer.Resource(&resources.ApprovalRule{}),
			infer.Resource(&resources.DeploymentSchedule{}),
			infer.Resource(&resources.DriftSchedule{}),
			infer.Resource(&resources.Environment{}),
			infer.Resource(&resources.EnvironmentRotationSchedule{}),
			infer.Resource(&resources.EnvironmentVersionTag{}),
			infer.Resource(&resources.InsightsAccount{}),
			infer.Resource(&resources.OidcIssuer{}),
			infer.Resource(&resources.OrgAccessToken{}),
			infer.Resource(&resources.OrganizationMember{}),
			infer.Resource(&resources.OrganizationRole{}),
			infer.Resource(&resources.PolicyGroup{}),
			infer.Resource(&resources.PolicyPack{}),
			infer.Resource(&resources.Stack{}),
			infer.Resource(&resources.StackTag{}),
			infer.Resource(&resources.StackTags{}),
			infer.Resource(&resources.TTLSchedule{}),
			infer.Resource(&resources.Team{}),
			infer.Resource(&resources.TeamAccessToken{}),
			infer.Resource(&resources.TeamEnvironmentPermission{}),
			infer.Resource(&resources.TeamRoleAssignment{}),
			infer.Resource(&resources.TeamStackPermission{}),
			infer.Resource(&resources.TemplateSource{}),
			infer.Resource(&resources.Webhook{}),
		).
		WithFunctions(
			infer.Function(&functions.BuildAccountMatchRuleFunction{}),
			infer.Function(&functions.BuildAllowPermissionsFunction{}),
			infer.Function(&functions.BuildEnvironmentScopedPermissionsFunction{}),
			infer.Function(&functions.BuildInsightsAccountScopedPermissionsFunction{}),
			infer.Function(&functions.BuildRolePermissionsFunction{}),
			infer.Function(&functions.BuildStackMatchRuleFunction{}),
			infer.Function(&functions.BuildStackScopedPermissionsFunction{}),
			infer.Function(&functions.GetCurrentUserFunction{}),
			infer.Function(&functions.GetEnvironmentFunction{}),
			infer.Function(&functions.GetInsightsAccountFunction{}),
			infer.Function(&functions.GetInsightsAccountsFunction{}),
			infer.Function(&functions.GetOrganizationMemberFunction{}),
			infer.Function(&functions.GetOrganizationMembersFunction{}),
			infer.Function(&functions.GetOrganizationRoleFunction{}),
			infer.Function(&functions.GetOrganizationRoleScopesFunction{}),
			infer.Function(&functions.GetOrganizationRolesFunction{}),
			infer.Function(&functions.GetPolicyPackFunction{}),
			infer.Function(&functions.GetPolicyPacksFunction{}),
			infer.Function(&functions.GetStackFunction{}),
		).
		WithModuleMap(map[tokens.ModuleName]tokens.ModuleName{
			"resources": "index",
			"functions": "index",
		}).
		WithConfig(infer.Config(&config.Config{})).
		WithLanguageMap(map[string]any{
			"csharp": map[string]any{
				"namespaces": map[string]any{
					"pulumiservice": "PulumiService",
					"api":           "Api",
				},
				"packageReferences": map[string]any{
					"Pulumi": "3.*",
				},
				respectSchemaVersionKey: true,
			},
			"go": map[string]any{
				"generateResourceContainerTypes": true,
				"importBasePath":                 "github.com/pulumi/pulumi-pulumiservice/sdk/go/pulumiservice",
				respectSchemaVersionKey:          true,
			},
			"java": map[string]any{
				"basePackage": "",
				"buildFiles":  "gradle",
				"dependencies": map[string]any{
					"com.google.code.findbugs:jsr305": "3.0.2",
					"com.google.code.gson:gson":       "2.8.9",
					"com.pulumi:pulumi":               "1.16.2",
				},
				"gradleNexusPublishPluginVersion": "2.0.0",
				"gradleTest":                      "",
				readmeKey:                         readme,
			},
			"nodejs": map[string]any{
				"packageName": "@pulumi/pulumiservice",
				"dependencies": map[string]any{
					"@pulumi/pulumi": "^3.0.0",
				},
				respectSchemaVersionKey: true,
				readmeKey:               readme,
			},
			"python": map[string]any{
				"packageName": "pulumi_pulumiservice",
				"requires": map[string]any{
					// 3.235.0 is the first runtime that preserves
					// `__`-prefixed input keys (pulumi/pulumi#22834),
					// which OrganizationRole.permissions relies on for
					// the `__type` discriminator at every level of the
					// descriptor tree.
					"pulumi": ">=3.235.0,<4.0.0",
				},
				"pyproject": map[string]any{
					"enabled": true,
				},
				respectSchemaVersionKey: true,
				readmeKey:               readme,
			},
		}).Build()
	if err != nil {
		return nil, err
	}
	provider = withFunctionDeprecations(provider, functions.DeprecatedFunctions)
	return p.RawServer(name, version, resources.WithEnvironmentCompat(provider))(host)
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

func withCloudApiSchema(prov p.Provider, spec *rest.Spec, metadata *rest.Metadata, pkg string) p.Provider {
	inner := prov.GetSchema
	prov.GetSchema = func(ctx context.Context, req p.GetSchemaRequest) (p.GetSchemaResponse, error) {
		resp, err := inner(ctx, req)
		if err != nil {
			return resp, err
		}
		var base schema.PackageSpec
		if err := json.Unmarshal([]byte(resp.Schema), &base); err != nil {
			return resp, fmt.Errorf("withCloudApiSchema: parse base schema: %w", err)
		}
		fragment, err := rest.BuildSchema(spec, metadata, pkg)
		if err != nil {
			return resp, fmt.Errorf("withCloudApiSchema: %w", err)
		}
		mergeSpec(&base, fragment)
		out, err := json.Marshal(base)
		if err != nil {
			return resp, fmt.Errorf("withCloudApiSchema: re-encode schema: %w", err)
		}
		resp.Schema = string(out)
		return resp, nil
	}
	return prov
}

// withFunctionDeprecations sets each function's deprecation message, which
// pulumi-go-provider doesn't emit for functions yet
// (pulumi/pulumi-go-provider#636). It wraps the built provider because infer
// adds its functions to the schema outside every inner layer.
func withFunctionDeprecations(prov p.Provider, messages map[string]string) p.Provider {
	inner := prov.GetSchema
	prov.GetSchema = func(ctx context.Context, req p.GetSchemaRequest) (p.GetSchemaResponse, error) {
		resp, err := inner(ctx, req)
		if err != nil {
			return resp, err
		}
		var spec schema.PackageSpec
		if err := json.Unmarshal([]byte(resp.Schema), &spec); err != nil {
			return resp, fmt.Errorf("withFunctionDeprecations: parse schema: %w", err)
		}
		for token, message := range messages {
			fn, ok := spec.Functions[token]
			if !ok {
				return resp, fmt.Errorf("withFunctionDeprecations: no function %q in the schema", token)
			}
			fn.DeprecationMessage = message
			spec.Functions[token] = fn
		}
		out, err := json.Marshal(spec)
		if err != nil {
			return resp, fmt.Errorf("withFunctionDeprecations: re-encode schema: %w", err)
		}
		resp.Schema = string(out)
		return resp, nil
	}
	return prov
}

func mergeSpec(dst, src *schema.PackageSpec) {
	if dst.Resources == nil {
		dst.Resources = map[string]schema.ResourceSpec{}
	}
	for k, v := range src.Resources {
		if _, ok := dst.Resources[k]; !ok {
			dst.Resources[k] = v
		}
	}
	if dst.Types == nil {
		dst.Types = map[string]schema.ComplexTypeSpec{}
	}
	for k, v := range src.Types {
		if _, ok := dst.Types[k]; !ok {
			dst.Types[k] = v
		}
	}
	if dst.Functions == nil {
		dst.Functions = map[string]schema.FunctionSpec{}
	}
	for k, v := range src.Functions {
		if _, ok := dst.Functions[k]; !ok {
			dst.Functions[k] = v
		}
	}
}

// Attach implements pulumirpc.ResourceProviderServer
func (k *pulumiserviceProvider) Attach(_ context.Context, req *pulumirpc.PluginAttach) (*pbempty.Empty, error) {
	host, err := provider.NewHostClient(req.Address)
	if err != nil {
		return nil, err
	}
	k.host = host
	return &pbempty.Empty{}, nil
}

// Construct creates a new component resource.
func (k *pulumiserviceProvider) Construct(
	_ context.Context,
	_ *pulumirpc.ConstructRequest,
) (*pulumirpc.ConstructResponse, error) {
	return nil, status.Error(codes.Unimplemented, "Construct is not yet implemented")
}

// CheckConfig validates the configuration for this provider.
func (k *pulumiserviceProvider) CheckConfig(
	_ context.Context,
	req *pulumirpc.CheckRequest,
) (*pulumirpc.CheckResponse, error) {
	return &pulumirpc.CheckResponse{Inputs: req.GetNews()}, nil
}

// DiffConfig diffs the configuration for this provider.
func (k *pulumiserviceProvider) DiffConfig(
	_ context.Context,
	_ *pulumirpc.DiffRequest,
) (*pulumirpc.DiffResponse, error) {
	return &pulumirpc.DiffResponse{}, nil
}

// Configure configures the resource provider with "globals" that control its behavior.
func (k *pulumiserviceProvider) Configure(
	_ context.Context,
	req *pulumirpc.ConfigureRequest,
) (*pulumirpc.ConfigureResponse, error) {

	sc := PulumiServiceConfig{}
	sc.Config = make(map[string]string)
	args, err := plugin.UnmarshalProperties(req.GetArgs(), plugin.MarshalOptions{
		KeepResources: true,
		SkipNulls:     true,
	})
	if err != nil {
		return nil, fmt.Errorf("unmarshaling configure args: %w", err)
	}
	for key, val := range args {
		// The engine sends the provider "version" alongside the config
		// properties; only string config values are meaningful here.
		if key == versionKey || !val.IsString() {
			continue
		}
		sc.Config[string(key)] = val.StringValue()
	}

	httpClient := http.Client{
		Timeout: 60 * time.Second,
	}
	token, err := sc.getPulumiAccessToken()
	if err != nil {
		return nil, err
	}
	url, err := sc.getPulumiServiceURL()
	if err != nil {
		return nil, err
	}
	client, err := pulumiapi.NewClient(&httpClient, *token, *url)
	if err != nil {
		return nil, err
	}

	transport := rest.Transport(&authedTransport{
		baseURL: *url,
		token:   *token,
		client:  &httpClient,
	})
	// Two paths: ctxmw.Wrap (set in MakeProvider) picks this up and attaches
	// it to every CRUD context; SetTransportResolver remains as the legacy
	// global so any code path that bypasses the middleware still resolves a
	// transport.
	if k.transportRef != nil {
		k.transportRef.Store(transport)
	}
	rest.SetTransportResolver(func(_ context.Context) (rest.Transport, error) {
		return transport, nil
	})

	k.pulumiResources = []PulumiServiceResource{
		&resources.PulumiServiceDeploymentSettingsResource{
			Client: client,
		},
	}

	return &pulumirpc.ConfigureResponse{
		AcceptSecrets: true,
	}, nil
}

// Invoke rejects every token: all functions are served by infer, which handles
// them before the request reaches this server.
func (k *pulumiserviceProvider) Invoke(
	_ context.Context,
	req *pulumirpc.InvokeRequest,
) (*pulumirpc.InvokeResponse, error) {
	return nil, fmt.Errorf("unknown Invoke token '%s'", req.GetTok())
}

// Check validates that the given property bag is valid for a resource of the given type and returns
// the inputs that should be passed to successive calls to Diff, Create, or Update for this
// resource. As a rule, the provider inputs returned by a call to Check should preserve the original
// representation of the properties as present in the program inputs. Though this rule is not
// required for correctness, violations thereof can negatively impact the end-user experience, as
// the provider inputs are using for detecting and rendering diffs.
func (k *pulumiserviceProvider) Check(
	_ context.Context,
	req *pulumirpc.CheckRequest,
) (*pulumirpc.CheckResponse, error) {
	rn := getResourceNameFromRequest(req)
	res := k.getPulumiServiceResource(rn)
	return res.Check(req)
}

// Diff checks what impacts a hypothetical update will have on the resource's properties.
func (k *pulumiserviceProvider) Diff(_ context.Context, req *pulumirpc.DiffRequest) (*pulumirpc.DiffResponse, error) {
	rn := getResourceNameFromRequest(req)
	res := k.getPulumiServiceResource(rn)
	return res.Diff(req)
}

// Create allocates a new instance of the provided resource and returns its unique ID afterwards.
func (k *pulumiserviceProvider) Create(
	_ context.Context,
	req *pulumirpc.CreateRequest,
) (*pulumirpc.CreateResponse, error) {
	rn := getResourceNameFromRequest(req)
	res := k.getPulumiServiceResource(rn)
	return res.Create(req)
}

// Read the current live state associated with a resource.
func (k *pulumiserviceProvider) Read(_ context.Context, req *pulumirpc.ReadRequest) (*pulumirpc.ReadResponse, error) {
	rn := getResourceNameFromRequest(req)
	res := k.getPulumiServiceResource(rn)
	return res.Read(req)
}

// Update updates an existing resource with new values.
func (k *pulumiserviceProvider) Update(
	_ context.Context,
	req *pulumirpc.UpdateRequest,
) (*pulumirpc.UpdateResponse, error) {
	rn := getResourceNameFromRequest(req)
	res := k.getPulumiServiceResource(rn)
	return res.Update(req)
}

// Delete tears down an existing resource with the given ID.  If it fails, the resource is assumed
// to still exist.
func (k *pulumiserviceProvider) Delete(_ context.Context, req *pulumirpc.DeleteRequest) (*pbempty.Empty, error) {
	rn := getResourceNameFromRequest(req)
	res := k.getPulumiServiceResource(rn)
	return res.Delete(req)
}

// GetPluginInfo returns generic information about this plugin, like its version.
func (k *pulumiserviceProvider) GetPluginInfo(context.Context, *pbempty.Empty) (*pulumirpc.PluginInfo, error) {
	return &pulumirpc.PluginInfo{
		Version: k.version,
	}, nil
}

// GetSchema returns the JSON-serialized schema for the provider.
func (k *pulumiserviceProvider) GetSchema(
	_ context.Context,
	_ *pulumirpc.GetSchemaRequest,
) (*pulumirpc.GetSchemaResponse, error) {
	return &pulumirpc.GetSchemaResponse{
		Schema: k.schema,
	}, nil
}

// Cancel signals the provider to gracefully shut down and abort any ongoing resource operations.
// Operations aborted in this way will return an error (e.g., `Update` and `Create` will either a
// creation error or an initialization error). Since Cancel is advisory and non-blocking, it is up
// to the host to decide how long to wait after Cancel is called before (e.g.)
// hard-closing any gRPC connection.
func (k *pulumiserviceProvider) Cancel(_ context.Context, _ *pbempty.Empty) (*pbempty.Empty, error) {
	// TODO
	return &pbempty.Empty{}, nil
}

func (k *pulumiserviceProvider) getPulumiServiceResource(name string) PulumiServiceResource {
	for _, r := range k.pulumiResources {
		if r.Name() == name {
			return r
		}
	}

	return &PulumiServiceUnknownResource{}
}

func getResourceNameFromRequest(req ResourceBase) string {
	urn := resource.URN(req.GetUrn())
	return urn.Type().String()
}

// mustSetSchemaVersion deserializes schemaStr from json, sets Version field
// then serializes back to json string
func mustSetSchemaVersion(schemaStr string, version string) string {
	var spec schema.PackageSpec
	if err := json.Unmarshal([]byte(schemaStr), &spec); err != nil {
		panic(fmt.Errorf("failed to parse schema: %w", err))
	}
	spec.Version = version
	bytes, err := json.Marshal(spec)
	if err != nil {
		panic(fmt.Errorf("failed to serialize versioned schema: %w", err))
	}
	return string(bytes)
}

type ResourceBase interface {
	GetUrn() string
}
