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
	"runtime"
	"strings"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	pbempty "google.golang.org/protobuf/types/known/emptypb"

	esc_client "github.com/pulumi/esc/cmd/esc/cli/client"
	"github.com/pulumi/esc/cmd/esc/cli/version"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/resources"
	"github.com/pulumi/pulumi/pkg/v3/codegen/schema"
	"github.com/pulumi/pulumi/pkg/v3/resource/provider"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
)

type PulumiServiceResource interface {
	Diff(ctx context.Context, req *pulumirpc.DiffRequest) (*pulumirpc.DiffResponse, error)
	Create(ctx context.Context, req *pulumirpc.CreateRequest) (*pulumirpc.CreateResponse, error)
	Delete(ctx context.Context, req *pulumirpc.DeleteRequest) (*pbempty.Empty, error)
	Check(ctx context.Context, req *pulumirpc.CheckRequest) (*pulumirpc.CheckResponse, error)
	Update(ctx context.Context, req *pulumirpc.UpdateRequest) (*pulumirpc.UpdateResponse, error)
	Read(ctx context.Context, req *pulumirpc.ReadRequest) (*pulumirpc.ReadResponse, error)
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
	client          *pulumiapi.Client
	escClient       esc_client.Client
}

func MakeProvider(host *provider.HostClient, name, version, schema string) (pulumirpc.ResourceProviderServer, error) {
	// inject version into schema
	versionedSchema := mustSetSchemaVersion(schema, version)
	// Return the new provider
	return &pulumiserviceProvider{
		host:        host,
		name:        name,
		schema:      versionedSchema,
		version:     version,
		AccessToken: "",
	}, nil
}

// Call dynamically executes a method in the provider associated with a component resource.
func (k *pulumiserviceProvider) Call(ctx context.Context, req *pulumirpc.CallRequest) (*pulumirpc.CallResponse, error) {
	return nil, status.Error(codes.Unimplemented, "Call is not yet implemented")
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
func (k *pulumiserviceProvider) Construct(ctx context.Context, req *pulumirpc.ConstructRequest) (*pulumirpc.ConstructResponse, error) {
	return nil, status.Error(codes.Unimplemented, "Construct is not yet implemented")
}

// CheckConfig validates the configuration for this provider.
func (k *pulumiserviceProvider) CheckConfig(ctx context.Context, req *pulumirpc.CheckRequest) (*pulumirpc.CheckResponse, error) {
	return &pulumirpc.CheckResponse{Inputs: req.GetNews()}, nil
}

// DiffConfig diffs the configuration for this provider.
func (k *pulumiserviceProvider) DiffConfig(ctx context.Context, req *pulumirpc.DiffRequest) (*pulumirpc.DiffResponse, error) {
	return &pulumirpc.DiffResponse{}, nil
}

// Configure configures the resource provider with "globals" that control its behavior.
func (k *pulumiserviceProvider) Configure(_ context.Context, req *pulumirpc.ConfigureRequest) (*pulumirpc.ConfigureResponse, error) {

	sc := PulumiServiceConfig{}
	sc.Config = make(map[string]string)
	for key, val := range req.GetVariables() {
		sc.Config[strings.TrimPrefix(key, "pulumiservice:config:")] = val
	}

	httpClient := http.Client{
		Timeout: 60 * time.Second,
	}
	token, err := sc.getPulumiAccessToken()
	if err != nil {
		return nil, err
	}
	url, err := sc.getPulumiServiceUrl()
	if err != nil {
		return nil, err
	}
	client, err := pulumiapi.NewClient(&httpClient, *token, *url)

	escClient := esc_client.New(fmt.Sprintf("provider-pulumiservice/1 (%s; %s)", version.Version, runtime.GOOS), *url, *token, false)

	if err != nil {
		return nil, err
	}

	// Store the clients for use in Invoke functions and context middleware
	k.client = client
	k.escClient = escClient

	// Resources are instantiated as empty structs following the context-based client injection pattern.
	// Clients are obtained from context via config.GetClient[T](ctx) in each resource method.
	// This pattern matches Ian's PR #258 implementation.
	k.pulumiResources = []PulumiServiceResource{
		&resources.PulumiServiceTeamResource{},
		&resources.PulumiServiceAccessTokenResource{},
		&resources.PulumiServiceWebhookResource{},
		// StackTag has been migrated to infer - see provider/pkg/infer/stack_tag.go
		// and is registered in hybrid.go buildInferOptions()
		&resources.TeamStackPermissionResource{},
		&resources.PulumiServiceTeamAccessTokenResource{},
		&resources.PulumiServiceOrgAccessTokenResource{},
		&resources.PulumiServiceDeploymentSettingsResource{},
		&resources.PulumiServiceAgentPoolResource{},
		&resources.PulumiServiceDeploymentScheduleResource{},
		&resources.PulumiServiceDriftScheduleResource{},
		&resources.PulumiServiceTtlScheduleResource{},
		&resources.PulumiServiceEnvironmentResource{},
		&resources.PulumiServiceTeamEnvironmentPermissionResource{},
		&resources.PulumiServiceEnvironmentVersionTagResource{},
		&resources.PulumiServiceStackResource{},
		&resources.PulumiServiceTemplateSourceResource{},
		&resources.PulumiServiceOidcIssuerResource{},
		&resources.PulumiServiceEnvironmentRotationScheduleResource{},
		&resources.PulumiServiceApprovalRuleResource{},
		&resources.PulumiServicePolicyGroupResource{},
	}

	return &pulumirpc.ConfigureResponse{
		AcceptSecrets: true,
	}, nil
}

// Invoke dynamically executes a built-in function in the provider.
func (k *pulumiserviceProvider) Invoke(ctx context.Context, req *pulumirpc.InvokeRequest) (*pulumirpc.InvokeResponse, error) {
	tok := req.GetTok()

	switch tok {
	case "pulumiservice:index:getPolicyPacks":
		return k.invokeFunctionGetPolicyPacks(ctx, req)
	case "pulumiservice:index:getPolicyPack":
		return k.invokeFunctionGetPolicyPack(ctx, req)
	default:
		return nil, fmt.Errorf("unknown Invoke token '%s'", tok)
	}
}

// Check validates that the given property bag is valid for a resource of the given type and returns
// the inputs that should be passed to successive calls to Diff, Create, or Update for this
// resource. As a rule, the provider inputs returned by a call to Check should preserve the original
// representation of the properties as present in the program inputs. Though this rule is not
// required for correctness, violations thereof can negatively impact the end-user experience, as
// the provider inputs are using for detecting and rendering diffs.
func (k *pulumiserviceProvider) Check(ctx context.Context, req *pulumirpc.CheckRequest) (*pulumirpc.CheckResponse, error) {
	rn := getResourceNameFromRequest(req)
	res := k.getPulumiServiceResource(rn)
	return res.Check(ctx, req)
}

// Diff checks what impacts a hypothetical update will have on the resource's properties.
func (k *pulumiserviceProvider) Diff(ctx context.Context, req *pulumirpc.DiffRequest) (*pulumirpc.DiffResponse, error) {
	rn := getResourceNameFromRequest(req)
	res := k.getPulumiServiceResource(rn)
	return res.Diff(ctx, req)
}

// Create allocates a new instance of the provided resource and returns its unique ID afterwards.
func (k *pulumiserviceProvider) Create(ctx context.Context, req *pulumirpc.CreateRequest) (*pulumirpc.CreateResponse, error) {
	rn := getResourceNameFromRequest(req)
	res := k.getPulumiServiceResource(rn)
	return res.Create(ctx, req)
}

// Read the current live state associated with a resource.
func (k *pulumiserviceProvider) Read(ctx context.Context, req *pulumirpc.ReadRequest) (*pulumirpc.ReadResponse, error) {
	rn := getResourceNameFromRequest(req)
	res := k.getPulumiServiceResource(rn)
	return res.Read(ctx, req)
}

// Update updates an existing resource with new values.
func (k *pulumiserviceProvider) Update(ctx context.Context, req *pulumirpc.UpdateRequest) (*pulumirpc.UpdateResponse, error) {
	rn := getResourceNameFromRequest(req)
	res := k.getPulumiServiceResource(rn)
	return res.Update(ctx, req)
}

// Delete tears down an existing resource with the given ID.  If it fails, the resource is assumed
// to still exist.
func (k *pulumiserviceProvider) Delete(ctx context.Context, req *pulumirpc.DeleteRequest) (*pbempty.Empty, error) {
	rn := getResourceNameFromRequest(req)
	res := k.getPulumiServiceResource(rn)
	return res.Delete(ctx, req)
}

// GetPluginInfo returns generic information about this plugin, like its version.
func (k *pulumiserviceProvider) GetPluginInfo(context.Context, *pbempty.Empty) (*pulumirpc.PluginInfo, error) {
	return &pulumirpc.PluginInfo{
		Version: k.version,
	}, nil
}

// GetSchema returns the JSON-serialized schema for the provider.
func (k *pulumiserviceProvider) GetSchema(ctx context.Context, req *pulumirpc.GetSchemaRequest) (*pulumirpc.GetSchemaResponse, error) {
	return &pulumirpc.GetSchemaResponse{
		Schema: k.schema,
	}, nil
}

// Cancel signals the provider to gracefully shut down and abort any ongoing resource operations.
// Operations aborted in this way will return an error (e.g., `Update` and `Create` will either a
// creation error or an initialization error). Since Cancel is advisory and non-blocking, it is up
// to the host to decide how long to wait after Cancel is called before (e.g.)
// hard-closing any gRPC connection.
func (k *pulumiserviceProvider) Cancel(context.Context, *pbempty.Empty) (*pbempty.Empty, error) {
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

// invokeFunctionGetPolicyPacks implements the getPolicyPacks function
func (k *pulumiserviceProvider) invokeFunctionGetPolicyPacks(ctx context.Context, req *pulumirpc.InvokeRequest) (*pulumirpc.InvokeResponse, error) {
	if k.client == nil {
		return nil, fmt.Errorf("provider not configured")
	}

	// Parse inputs
	inputs, err := plugin.UnmarshalProperties(req.GetArgs(), plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	orgName := inputs["organizationName"].StringValue()
	if orgName == "" {
		return nil, fmt.Errorf("organizationName is required")
	}

	// Call the API
	policyPacks, err := k.client.ListPolicyPacks(ctx, orgName)
	if err != nil {
		return nil, fmt.Errorf("failed to list policy packs: %w", err)
	}

	// Build output
	outputProps := resource.PropertyMap{
		"policyPacks": resource.NewPropertyValue(convertPolicyPacksToProperties(policyPacks)),
	}

	outputProperties, err := plugin.MarshalProperties(outputProps, plugin.MarshalOptions{})
	if err != nil {
		return nil, err
	}

	return &pulumirpc.InvokeResponse{
		Return: outputProperties,
	}, nil
}

// invokeFunctionGetPolicyPack implements the getPolicyPack function
func (k *pulumiserviceProvider) invokeFunctionGetPolicyPack(ctx context.Context, req *pulumirpc.InvokeRequest) (*pulumirpc.InvokeResponse, error) {
	if k.client == nil {
		return nil, fmt.Errorf("provider not configured")
	}

	// Parse inputs
	inputs, err := plugin.UnmarshalProperties(req.GetArgs(), plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	orgName := inputs["organizationName"].StringValue()
	if orgName == "" {
		return nil, fmt.Errorf("organizationName is required")
	}

	policyPackName := inputs["policyPackName"].StringValue()
	if policyPackName == "" {
		return nil, fmt.Errorf("policyPackName is required")
	}

	// Call the API - either specific version or latest
	var policyPack *pulumiapi.PolicyPackDetail
	if inputs["version"].HasValue() && inputs["version"].IsNumber() {
		version := int(inputs["version"].NumberValue())
		policyPack, err = k.client.GetPolicyPack(ctx, orgName, policyPackName, version)
	} else {
		policyPack, err = k.client.GetLatestPolicyPack(ctx, orgName, policyPackName)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get policy pack: %w", err)
	}

	if policyPack == nil {
		return nil, fmt.Errorf("policy pack not found")
	}

	// Build output
	outputProps := convertPolicyPackDetailToProperties(policyPack)

	outputProperties, err := plugin.MarshalProperties(outputProps, plugin.MarshalOptions{})
	if err != nil {
		return nil, err
	}

	return &pulumirpc.InvokeResponse{
		Return: outputProperties,
	}, nil
}

// Helper functions to convert API types to property values
func convertPolicyPacksToProperties(packs []pulumiapi.PolicyPackWithVersions) []resource.PropertyValue {
	result := make([]resource.PropertyValue, len(packs))
	for i, pack := range packs {
		versions := make([]resource.PropertyValue, len(pack.Versions))
		for j, v := range pack.Versions {
			versions[j] = resource.NewNumberProperty(float64(v))
		}

		versionTags := make([]resource.PropertyValue, len(pack.VersionTags))
		for j, vt := range pack.VersionTags {
			versionTags[j] = resource.NewStringProperty(vt)
		}

		result[i] = resource.NewObjectProperty(resource.PropertyMap{
			"name":        resource.NewStringProperty(pack.Name),
			"displayName": resource.NewStringProperty(pack.DisplayName),
			"versions":    resource.NewArrayProperty(versions),
			"versionTags": resource.NewArrayProperty(versionTags),
		})
	}
	return result
}

func convertPolicyPackDetailToProperties(pack *pulumiapi.PolicyPackDetail) resource.PropertyMap {
	props := resource.PropertyMap{
		"name":        resource.NewStringProperty(pack.Name),
		"displayName": resource.NewStringProperty(pack.DisplayName),
		"version":     resource.NewNumberProperty(float64(pack.Version)),
	}

	if pack.VersionTag != "" {
		props["versionTag"] = resource.NewStringProperty(pack.VersionTag)
	}

	if pack.Config != nil {
		props["config"] = resource.NewObjectProperty(convertMapToPropertyMap(pack.Config))
	}

	if len(pack.Policies) > 0 {
		policies := make([]resource.PropertyValue, len(pack.Policies))
		for i, policy := range pack.Policies {
			policyProps := resource.PropertyMap{
				"name": resource.NewStringProperty(policy.Name),
			}
			if policy.DisplayName != "" {
				policyProps["displayName"] = resource.NewStringProperty(policy.DisplayName)
			}
			if policy.Description != "" {
				policyProps["description"] = resource.NewStringProperty(policy.Description)
			}
			if policy.EnforcementLevel != "" {
				policyProps["enforcementLevel"] = resource.NewStringProperty(policy.EnforcementLevel)
			}
			if policy.Message != "" {
				policyProps["message"] = resource.NewStringProperty(policy.Message)
			}
			if policy.ConfigSchema != nil {
				policyProps["configSchema"] = resource.NewObjectProperty(convertMapToPropertyMap(policy.ConfigSchema))
			}
			policies[i] = resource.NewObjectProperty(policyProps)
		}
		props["policies"] = resource.NewArrayProperty(policies)
	}

	return props
}

func convertMapToPropertyMap(m map[string]interface{}) resource.PropertyMap {
	props := resource.PropertyMap{}
	for k, v := range m {
		props[resource.PropertyKey(k)] = convertInterfaceToPropertyValue(v)
	}
	return props
}

func convertInterfaceToPropertyValue(v interface{}) resource.PropertyValue {
	if v == nil {
		return resource.NewNullProperty()
	}

	switch val := v.(type) {
	case string:
		return resource.NewStringProperty(val)
	case float64:
		return resource.NewNumberProperty(val)
	case int:
		return resource.NewNumberProperty(float64(val))
	case bool:
		return resource.NewBoolProperty(val)
	case []interface{}:
		arr := make([]resource.PropertyValue, len(val))
		for i, item := range val {
			arr[i] = convertInterfaceToPropertyValue(item)
		}
		return resource.NewArrayProperty(arr)
	case map[string]interface{}:
		return resource.NewObjectProperty(convertMapToPropertyMap(val))
	default:
		return resource.NewNullProperty()
	}
}
