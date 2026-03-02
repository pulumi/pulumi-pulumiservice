// Copyright 2016-2026, Pulumi Corporation.
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
	"fmt"
	"strings"

	pbempty "google.golang.org/protobuf/types/known/emptypb"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/util"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
)

type PulumiServiceAdoIntegrationInput struct {
	Organization        string
	AdoOrganizationName string
	ProjectID           string
	DisablePRComments   bool
	DisableNeoSummaries bool
	DisableDetailedDiff bool
}

type PulumiServiceAdoIntegrationProperties struct {
	PulumiServiceAdoIntegrationInput
	IntegrationID      string
	Valid              bool
	AdoOrganizationID  string
	AdoOrganizationURL string
	ProjectName        string
}

func (p *PulumiServiceAdoIntegrationProperties) ToPropertyMap() resource.PropertyMap {
	pm := resource.PropertyMap{}
	pm["organization"] = resource.NewPropertyValue(p.Organization)
	pm["adoOrganizationName"] = resource.NewPropertyValue(p.AdoOrganizationName)
	pm["projectId"] = resource.NewPropertyValue(p.ProjectID)
	pm["disablePRComments"] = resource.NewPropertyValue(p.DisablePRComments)
	pm["disableNeoSummaries"] = resource.NewPropertyValue(p.DisableNeoSummaries)
	pm["disableDetailedDiff"] = resource.NewPropertyValue(p.DisableDetailedDiff)
	pm["integrationId"] = resource.NewPropertyValue(p.IntegrationID)
	pm["valid"] = resource.NewPropertyValue(p.Valid)
	pm["adoOrganizationId"] = resource.NewPropertyValue(p.AdoOrganizationID)
	pm["adoOrganizationUrl"] = resource.NewPropertyValue(p.AdoOrganizationURL)
	pm["projectName"] = resource.NewPropertyValue(p.ProjectName)
	return pm
}

func ToPulumiServiceAdoIntegrationInput(propMap resource.PropertyMap) PulumiServiceAdoIntegrationInput {
	input := PulumiServiceAdoIntegrationInput{}
	input.Organization = util.GetSecretOrStringValue(propMap["organization"])
	input.AdoOrganizationName = util.GetSecretOrStringValue(propMap["adoOrganizationName"])
	input.ProjectID = util.GetSecretOrStringValue(propMap["projectId"])
	input.DisablePRComments = util.GetSecretOrBoolValue(propMap["disablePRComments"])
	input.DisableNeoSummaries = util.GetSecretOrBoolValue(propMap["disableNeoSummaries"])
	input.DisableDetailedDiff = util.GetSecretOrBoolValue(propMap["disableDetailedDiff"])
	return input
}

type PulumiServiceAdoIntegrationResource struct {
	Client pulumiapi.AzureDevOpsIntegrationClient
}

func (r *PulumiServiceAdoIntegrationResource) Name() string {
	return "pulumiservice:index:AzureDevOpsIntegration"
}

func (r *PulumiServiceAdoIntegrationResource) Check(req *pulumirpc.CheckRequest) (*pulumirpc.CheckResponse, error) {
	news, err := plugin.UnmarshalProperties(
		req.GetNews(),
		plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true, KeepSecrets: true},
	)
	if err != nil {
		return nil, err
	}

	var failures []*pulumirpc.CheckFailure
	for _, p := range []resource.PropertyKey{"organization", "adoOrganizationName", "projectId"} {
		if !news[p].HasValue() {
			failures = append(failures, &pulumirpc.CheckFailure{
				Reason:   fmt.Sprintf("missing required property '%s'", p),
				Property: string(p),
			})
		}
	}

	inputNews, err := plugin.MarshalProperties(
		news,
		plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true, KeepSecrets: true},
	)
	if err != nil {
		return nil, err
	}

	return &pulumirpc.CheckResponse{Inputs: inputNews, Failures: failures}, nil
}

func (r *PulumiServiceAdoIntegrationResource) Diff(req *pulumirpc.DiffRequest) (*pulumirpc.DiffResponse, error) {
	olds, err := plugin.UnmarshalProperties(
		req.GetOldInputs(),
		plugin.MarshalOptions{KeepUnknowns: false, SkipNulls: true},
	)
	if err != nil {
		return nil, err
	}

	news, err := plugin.UnmarshalProperties(
		req.GetNews(),
		plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true},
	)
	if err != nil {
		return nil, err
	}

	diffs := olds.Diff(news)
	if diffs == nil {
		return &pulumirpc.DiffResponse{
			Changes: pulumirpc.DiffResponse_DIFF_NONE,
		}, nil
	}

	dd := plugin.NewDetailedDiffFromObjectDiff(diffs, false)

	detailedDiffs := map[string]*pulumirpc.PropertyDiff{}
	replaceProperties := map[string]bool{
		"organization":        true,
		"adoOrganizationName": true,
		"projectId":           true,
	}
	for k, v := range dd {
		if _, ok := replaceProperties[k]; ok {
			v.Kind = v.Kind.AsReplace()
		}
		detailedDiffs[k] = &pulumirpc.PropertyDiff{
			Kind:      pulumirpc.PropertyDiff_Kind(v.Kind), //nolint:gosec // safe conversion from plugin.DiffKind
			InputDiff: v.InputDiff,
		}
	}

	changes := pulumirpc.DiffResponse_DIFF_NONE
	if len(detailedDiffs) > 0 {
		changes = pulumirpc.DiffResponse_DIFF_SOME
	}
	return &pulumirpc.DiffResponse{
		Changes:         changes,
		DetailedDiff:    detailedDiffs,
		HasDetailedDiff: true,
	}, nil
}

func (r *PulumiServiceAdoIntegrationResource) Create(req *pulumirpc.CreateRequest) (*pulumirpc.CreateResponse, error) {
	ctx := context.Background()
	inputMap, err := plugin.UnmarshalProperties(
		req.GetProperties(),
		plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true, KeepSecrets: true},
	)
	if err != nil {
		return nil, err
	}

	input := ToPulumiServiceAdoIntegrationInput(inputMap)

	// ADO integrations are created via the Pulumi Cloud UI (OAuth flow).
	// This resource adopts an existing integration by looking it up.
	integrations, err := r.Client.ListAzureDevOpsIntegrations(ctx, input.Organization)
	if err != nil {
		return nil, fmt.Errorf("failed to list azure devops integrations: %w", err)
	}

	var integration *pulumiapi.AzureDevOpsIntegration
	for _, i := range integrations {
		if i.Organization.Name == input.AdoOrganizationName && i.Project.ID == input.ProjectID {
			integration = &i
			break
		}
	}

	if integration == nil {
		return nil, fmt.Errorf(
			"no azure devops integration found for ADO organization %q and project %q in Pulumi organization %q; "+
				"integrations must be created via the Pulumi Cloud console first",
			input.AdoOrganizationName, input.ProjectID, input.Organization,
		)
	}

	// Apply the desired settings
	updateReq := pulumiapi.UpdateAzureDevOpsIntegrationRequest{
		DisablePRComments:   input.DisablePRComments,
		DisableNeoSummaries: input.DisableNeoSummaries,
		DisableDetailedDiff: input.DisableDetailedDiff,
	}
	err = r.Client.UpdateAzureDevOpsIntegration(
		ctx, input.Organization, integration.ID, updateReq,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update azure devops integration settings: %w", err)
	}

	props := PulumiServiceAdoIntegrationProperties{
		PulumiServiceAdoIntegrationInput: input,
		IntegrationID:                    integration.ID,
		Valid:                            integration.Valid,
		AdoOrganizationID:                integration.Organization.ID,
		AdoOrganizationURL:               integration.Organization.AccountURL,
		ProjectName:                      integration.Project.Name,
	}

	properties, err := plugin.MarshalProperties(
		props.ToPropertyMap(),
		plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true},
	)
	if err != nil {
		return nil, err
	}

	return &pulumirpc.CreateResponse{
		Id:         fmt.Sprintf("%s/%s", input.Organization, integration.ID),
		Properties: properties,
	}, nil
}

func (r *PulumiServiceAdoIntegrationResource) Read(req *pulumirpc.ReadRequest) (*pulumirpc.ReadResponse, error) {
	ctx := context.Background()

	orgName, integrationID, err := splitAdoIntegrationID(req.GetId())
	if err != nil {
		return nil, err
	}

	integration, err := r.Client.GetAzureDevOpsIntegration(ctx, orgName, integrationID)
	if err != nil {
		return nil, err
	}

	if integration == nil {
		return &pulumirpc.ReadResponse{}, nil
	}

	props := PulumiServiceAdoIntegrationProperties{
		PulumiServiceAdoIntegrationInput: PulumiServiceAdoIntegrationInput{
			Organization:        orgName,
			AdoOrganizationName: integration.Organization.Name,
			ProjectID:           integration.Project.ID,
			DisablePRComments:   integration.DisablePRComments,
			DisableNeoSummaries: integration.DisableNeoSummaries,
			DisableDetailedDiff: integration.DisableDetailedDiff,
		},
		IntegrationID:      integration.ID,
		Valid:              integration.Valid,
		AdoOrganizationID:  integration.Organization.ID,
		AdoOrganizationURL: integration.Organization.AccountURL,
		ProjectName:        integration.Project.Name,
	}

	properties, err := plugin.MarshalProperties(
		props.ToPropertyMap(),
		plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true, KeepSecrets: true},
	)
	if err != nil {
		return nil, err
	}

	// Build inputs (only the input properties)
	inputMap := resource.PropertyMap{}
	inputMap["organization"] = resource.NewPropertyValue(props.Organization)
	inputMap["adoOrganizationName"] = resource.NewPropertyValue(props.AdoOrganizationName)
	inputMap["projectId"] = resource.NewPropertyValue(props.ProjectID)
	inputMap["disablePRComments"] = resource.NewPropertyValue(props.DisablePRComments)
	inputMap["disableNeoSummaries"] = resource.NewPropertyValue(props.DisableNeoSummaries)
	inputMap["disableDetailedDiff"] = resource.NewPropertyValue(props.DisableDetailedDiff)

	inputs, err := plugin.MarshalProperties(
		inputMap,
		plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true, KeepSecrets: true},
	)
	if err != nil {
		return nil, err
	}

	return &pulumirpc.ReadResponse{
		Id:         req.GetId(),
		Properties: properties,
		Inputs:     inputs,
	}, nil
}

func (r *PulumiServiceAdoIntegrationResource) Update(req *pulumirpc.UpdateRequest) (*pulumirpc.UpdateResponse, error) {
	ctx := context.Background()

	orgName, integrationID, err := splitAdoIntegrationID(req.GetId())
	if err != nil {
		return nil, fmt.Errorf("invalid resource id: %v", err)
	}

	inputMap, err := plugin.UnmarshalProperties(
		req.GetNews(),
		plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true, KeepSecrets: true},
	)
	if err != nil {
		return nil, err
	}

	input := ToPulumiServiceAdoIntegrationInput(inputMap)

	err = r.Client.UpdateAzureDevOpsIntegration(ctx, orgName, integrationID, pulumiapi.UpdateAzureDevOpsIntegrationRequest{
		DisablePRComments:   input.DisablePRComments,
		DisableNeoSummaries: input.DisableNeoSummaries,
		DisableDetailedDiff: input.DisableDetailedDiff,
	})
	if err != nil {
		return nil, err
	}

	// Re-read to get the latest state
	integration, err := r.Client.GetAzureDevOpsIntegration(ctx, orgName, integrationID)
	if err != nil {
		return nil, err
	}

	if integration == nil {
		return nil, fmt.Errorf("azure devops integration %q not found after update", integrationID)
	}

	props := PulumiServiceAdoIntegrationProperties{
		PulumiServiceAdoIntegrationInput: input,
		IntegrationID:                    integration.ID,
		Valid:                            integration.Valid,
		AdoOrganizationID:                integration.Organization.ID,
		AdoOrganizationURL:               integration.Organization.AccountURL,
		ProjectName:                      integration.Project.Name,
	}

	properties, err := plugin.MarshalProperties(
		props.ToPropertyMap(),
		plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true},
	)
	if err != nil {
		return nil, err
	}

	return &pulumirpc.UpdateResponse{
		Properties: properties,
	}, nil
}

func (r *PulumiServiceAdoIntegrationResource) Delete(req *pulumirpc.DeleteRequest) (*pbempty.Empty, error) {
	ctx := context.Background()

	orgName, integrationID, err := splitAdoIntegrationID(req.GetId())
	if err != nil {
		return nil, err
	}

	err = r.Client.DeleteAzureDevOpsIntegration(ctx, orgName, integrationID)
	return &pbempty.Empty{}, err
}

func splitAdoIntegrationID(id string) (string, string, error) {
	// format: orgName/integrationID
	s := strings.SplitN(id, "/", 2)
	if len(s) != 2 {
		return "", "", fmt.Errorf("%q is not a valid Azure DevOps integration ID, expected format: orgName/integrationID", id)
	}
	return s[0], s[1], nil
}
