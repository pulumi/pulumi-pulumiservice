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
	pbempty "google.golang.org/protobuf/types/known/emptypb"

	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/pulumi/pulumi-go-provider/middleware/schema"
	"github.com/pulumi/pulumi/pkg/v3/resource/provider"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
)

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
}

func makeProvider(host *provider.HostClient, name, version string) (pulumirpc.ResourceProviderServer, error) {
	return p.RawServer(name, version, infer.Provider(infer.Options{
		Metadata: schema.Metadata{
			LanguageMap: map[string]any{
				"csharp": map[string]any{
					"namespaces": map[string]string{
						"pulumiservice": "PulumiService",
					},
					"packageReferences": map[string]string{
						"Pulumi": "3.*",
					},
				},
				"go": map[string]any{
					"generateResourceContainerTypes": true,
					"importBasePath":                 "github.com/pulumi/pulumi-pulumiservice/sdk/go/pulumiservice",
				},
				"nodejs": map[string]any{
					"packageName": "@pulumi/pulumiservice",
					"dependencies": map[string]string{
						"@pulumi/pulumi": "^3.0.0",
					},
				},
				"python": map[string]any{
					"packageName": "pulumi_pulumiservice",
					"requires": map[string]string{
						"pulumi": ">=3.0.0,<4.0.0",
					},
				},
			},
			Description: "A native Pulumi package for creating and managing Pulumi Cloud constructs",
			DisplayName: "Pulumi Cloud",
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
		Config: infer.Config[Config](),
		Resources: []infer.InferredResource{
			infer.Resource[*AgentPool](),
		},
	}))(nil)
}
