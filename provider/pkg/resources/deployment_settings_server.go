// Copyright 2026, Pulumi Corporation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package resources

import (
	"context"
	"errors"
	"strings"

	_ "embed" // for the //go:embed example file below

	pbempty "google.golang.org/protobuf/types/known/emptypb"

	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
	mw "github.com/pulumi/pulumi-go-provider/middleware"
	"github.com/pulumi/pulumi-go-provider/middleware/rpc"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/config"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/util"
)

// DeploymentSettingsToken is the type token of DeploymentSettings.
const DeploymentSettingsToken = "pulumiservice:index:DeploymentSettings"

// DeploymentSettings defines the schema of pulumiservice:index:DeploymentSettings.
//
// infer only generates its schema. Its requests are served by
// [NewDeploymentSettingsHandler], which runs the property-map implementation in
// deployment_settings.go: infer's typed inputs can't carry which environment
// variables are secret, and its Diff can't see the old inputs.
type DeploymentSettings struct{}

var _ infer.CustomResource[DeploymentSettingsInput, DeploymentSettingsState] = &DeploymentSettings{}

// The example file is generated from the Pulumi YAML in examples/; edit
// that and run `make generate_examples`.
//
//go:embed deployment_settings_examples.md
var deploymentSettingsExamples string

func (r *DeploymentSettings) Annotate(a infer.Annotator) {
	a.Describe(r, "Deployment settings configure Pulumi Deployments for a stack.\n\n"+
		escRecommendation+" See "+cloudCredentialsGuide+" to choose between ESC and Deployments OIDC "+
		"(`operationContext.oidc`).\n\n"+
		strings.TrimSpace(deploymentSettingsExamples)+
		util.ImportDocs(DeploymentSettingsToken, "{organization}/{project}/{stack}", "my-org/my-project/my-stack",
			"Secret values can't be read back from Pulumi Cloud, so import sets each one to the placeholder "+
				"`<REPLACE WITH ACTUAL SECRET VALUE>`. Replace them with the real values in your program."))
	a.SetToken("index", "DeploymentSettings")
}

// Create is never called: [NewDeploymentSettingsHandler] serves every request.
func (*DeploymentSettings) Create(
	context.Context, infer.CreateRequest[DeploymentSettingsInput],
) (infer.CreateResponse[DeploymentSettingsState], error) {
	return infer.CreateResponse[DeploymentSettingsState]{},
		errors.New("DeploymentSettings is served by NewDeploymentSettingsHandler")
}

// NewDeploymentSettingsHandler serves DeploymentSettings requests with
// [PulumiServiceDeploymentSettingsResource].
//
// It goes through [rpc.Provider], as the resource did when the gRPC provider
// served it, so requests are converted the same way. That provider is
// configured here, once, with the response the gRPC provider gave: secrets are
// passed through, and preview doesn't call Create or Update.
func NewDeploymentSettingsHandler() (mw.CustomResource, error) {
	prov := rpc.Provider(deploymentSettingsServer{})
	if err := prov.Configure(context.Background(), p.ConfigureRequest{}); err != nil {
		return nil, err
	}
	return deploymentSettingsHandler{prov}, nil
}

type deploymentSettingsHandler struct{ p.Provider }

func (h deploymentSettingsHandler) Check(ctx context.Context, req p.CheckRequest) (p.CheckResponse, error) {
	return h.Provider.Check(ctx, req)
}

func (h deploymentSettingsHandler) Diff(ctx context.Context, req p.DiffRequest) (p.DiffResponse, error) {
	return h.Provider.Diff(ctx, req)
}

func (h deploymentSettingsHandler) Create(ctx context.Context, req p.CreateRequest) (p.CreateResponse, error) {
	return h.Provider.Create(ctx, req)
}

func (h deploymentSettingsHandler) Read(ctx context.Context, req p.ReadRequest) (p.ReadResponse, error) {
	return h.Provider.Read(ctx, req)
}

func (h deploymentSettingsHandler) Update(ctx context.Context, req p.UpdateRequest) (p.UpdateResponse, error) {
	return h.Provider.Update(ctx, req)
}

func (h deploymentSettingsHandler) Delete(ctx context.Context, req p.DeleteRequest) error {
	return h.Provider.Delete(ctx, req)
}

// deploymentSettingsServer adapts [PulumiServiceDeploymentSettingsResource] to
// a gRPC server, with the client infer configured for the request.
type deploymentSettingsServer struct {
	pulumirpc.UnimplementedResourceProviderServer
}

func deploymentSettingsResource(ctx context.Context) *PulumiServiceDeploymentSettingsResource {
	return &PulumiServiceDeploymentSettingsResource{Client: config.GetClient(ctx)}
}

func (deploymentSettingsServer) Configure(
	context.Context, *pulumirpc.ConfigureRequest,
) (*pulumirpc.ConfigureResponse, error) {
	return &pulumirpc.ConfigureResponse{AcceptSecrets: true}, nil
}

func (deploymentSettingsServer) Check(
	ctx context.Context, req *pulumirpc.CheckRequest,
) (*pulumirpc.CheckResponse, error) {
	return deploymentSettingsResource(ctx).Check(req)
}

func (deploymentSettingsServer) Diff(ctx context.Context, req *pulumirpc.DiffRequest) (*pulumirpc.DiffResponse, error) {
	return deploymentSettingsResource(ctx).Diff(req)
}

func (deploymentSettingsServer) Create(
	ctx context.Context, req *pulumirpc.CreateRequest,
) (*pulumirpc.CreateResponse, error) {
	return deploymentSettingsResource(ctx).Create(req)
}

func (deploymentSettingsServer) Read(ctx context.Context, req *pulumirpc.ReadRequest) (*pulumirpc.ReadResponse, error) {
	return deploymentSettingsResource(ctx).Read(req)
}

func (deploymentSettingsServer) Update(
	ctx context.Context, req *pulumirpc.UpdateRequest,
) (*pulumirpc.UpdateResponse, error) {
	return deploymentSettingsResource(ctx).Update(req)
}

func (deploymentSettingsServer) Delete(ctx context.Context, req *pulumirpc.DeleteRequest) (*pbempty.Empty, error) {
	return deploymentSettingsResource(ctx).Delete(req)
}
