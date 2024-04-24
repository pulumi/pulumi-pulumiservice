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

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	pbempty "google.golang.org/protobuf/types/known/emptypb"

	"github.com/pulumi/pulumi-go-provider/infer"
	inferSchema "github.com/pulumi/pulumi-go-provider/middleware/schema"
	"github.com/pulumi/pulumi/pkg/v3/codegen/schema"
	"github.com/pulumi/pulumi/pkg/v3/resource/provider"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/tokens"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
)

func inferProvider() infer.Options {
	return infer.Options{
		Metadata: inferSchema.Metadata{
			DisplayName: "Pulumi Cloud",
			Description: "A native Pulumi package for creating and managing Pulumi Cloud constructs",
			LanguageMap: map[string]any{
				"csharp": map[string]any{
					"namespaces": map[string]any{
						"pulumiservice": "PulumiService",
					},
					"packageReferences": map[string]any{
						"Pulumi": "3.*",
					},
				},
				"go": map[string]any{
					"generateResourceContainerTypes": true,
					"importBasePath":                 "github.com/pulumi/pulumi-pulumiservice/sdk/go/pulumiservice",
				},
				"nodejs": map[string]any{
					"packageName": "@pulumi/pulumiservice",
					"dependencies": map[string]any{
						"@pulumi/pulumi": "^3.0.0",
					},
				},
				"python": map[string]any{
					"packageName": "pulumi_pulumiservice",
					"requires": map[string]any{
						"pulumi": ">=3.0.0,<4.0.0",
					},
				},
			},
			Keywords: []string{
				"pulumi",
				"kind/native",
				"category/infrastructure",
			},
			Homepage:   "https://pulumi.com",
			Repository: "https://github.com/pulumi/pulumi-pulumiservice",
			Publisher:  "Pulumi",
			License:    "Apache-2.0",
		},
		Resources:  []infer.InferredResource{},
		Components: []infer.InferredComponent{},
		Functions:  []infer.InferredFunction{},
		Config:     infer.Config[*Config](),
		ModuleMap: map[tokens.ModuleName]tokens.ModuleName{
			"provider": "index",
		},
	}
}

type PulumiServiceResource interface {
	Diff(context.Context, *pulumirpc.DiffRequest) (*pulumirpc.DiffResponse, error)
	Create(context.Context, *pulumirpc.CreateRequest) (*pulumirpc.CreateResponse, error)
	Delete(context.Context, *pulumirpc.DeleteRequest) (*pbempty.Empty, error)
	Check(context.Context, *pulumirpc.CheckRequest) (*pulumirpc.CheckResponse, error)
	Update(context.Context, *pulumirpc.UpdateRequest) (*pulumirpc.UpdateResponse, error)
	Read(context.Context, *pulumirpc.ReadRequest) (*pulumirpc.ReadResponse, error)
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
func (k *pulumiserviceProvider) Configure(context.Context, *pulumirpc.ConfigureRequest) (*pulumirpc.ConfigureResponse, error) {
	return &pulumirpc.ConfigureResponse{AcceptSecrets: true}, nil
}

// Invoke dynamically executes a built-in function in the provider.
func (k *pulumiserviceProvider) Invoke(_ context.Context, req *pulumirpc.InvokeRequest) (*pulumirpc.InvokeResponse, error) {
	tok := req.GetTok()
	return nil, fmt.Errorf("unknown Invoke token '%s'", tok)
}

// StreamInvoke dynamically executes a built-in function in the provider. The result is streamed
// back as a series of messages.
func (k *pulumiserviceProvider) StreamInvoke(req *pulumirpc.InvokeRequest, server pulumirpc.ResourceProvider_StreamInvokeServer) error {
	tok := req.GetTok()
	return fmt.Errorf("unknown StreamInvoke token '%s'", tok)
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
	return &pulumirpc.GetSchemaResponse{Schema: k.schema}, nil
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
