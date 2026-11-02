// Copyright 2026, Pulumi Corporation.
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

package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/pulumi/pulumi-go-provider/infer"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/config"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/util"
)

type AgentDefinition struct{}

var (
	_ infer.CustomCreate[AgentDefinitionInput, AgentDefinitionState] = &AgentDefinition{}
	_ infer.CustomUpdate[AgentDefinitionInput, AgentDefinitionState] = &AgentDefinition{}
	_ infer.CustomRead[AgentDefinitionInput, AgentDefinitionState]   = &AgentDefinition{}
	_ infer.CustomDelete[AgentDefinitionState]                       = &AgentDefinition{}
)

func (*AgentDefinition) Annotate(a infer.Annotator) {
	a.SetToken("index", "AgentDefinition")
	a.Describe(&AgentDefinition{}, "A saved, team-owned Custom Agent in Pulumi Cloud. Its role, instructions, "+
		"and integrations apply to tasks run as the agent. The resource ID is the durable definition UUID; "+
		"renaming or updating the agent preserves it. Use this ID in `delivery.Agent.agent`."+
		util.ImportDocs("pulumiservice:index:AgentDefinition", "{organization}/{definitionId}",
			"my-org/4f1d6f44-bf29-435b-af3c-c6309c0eacb4",
			"Import resolves the active definition and stores its UUID as the resource ID. "+
				"The provider's caller must be able to see the agent's role."))
}

type AgentApprovalMode string

func (AgentApprovalMode) Values() []infer.EnumValue[AgentApprovalMode] {
	return []infer.EnumValue[AgentApprovalMode]{
		{Name: "Auto", Value: "auto", Description: "Approve agent runs automatically."},
		{Name: "Manual", Value: "manual", Description: "Pause each agent run until a person approves it."},
	}
}

type AgentDefinitionIntegration struct {
	ID string `pulumi:"id"`
}

func (i *AgentDefinitionIntegration) Annotate(a infer.Annotator) {
	a.Describe(i, "An integration enabled for the agent's tasks.")
	a.Describe(&i.ID, "The catalog integration ID, such as `honeycomb` or `datadog`, not a GitHub installation ID.")
}

type AgentDefinitionInput struct {
	OrganizationName   string  `pulumi:"organizationName" provider:"replaceOnChanges"`
	Name               string  `pulumi:"name"`
	DisplayName        string  `pulumi:"displayName,optional"`
	RoleID             *string `pulumi:"roleId,optional"`
	CustomInstructions string  `pulumi:"customInstructions,optional"`
	// Pointer indirection preserves an explicitly empty list through infer decoding.
	EnabledIntegrations *[]AgentDefinitionIntegration `pulumi:"enabledIntegrations,optional"`
	Purpose             string                        `pulumi:"purpose,optional"`
	Avatar              *string                       `pulumi:"avatar,optional"`
	ApprovalMode        *AgentApprovalMode            `pulumi:"approvalMode,optional"`
}

func (i *AgentDefinitionInput) Annotate(a infer.Annotator) {
	a.Describe(&i.OrganizationName, "The Pulumi Cloud organization that owns the Custom Agent.")
	a.Describe(&i.Name, "The agent's name, unique within the organization: lowercase letters, digits, and hyphens, "+
		"up to 64 characters. Changing the name preserves the definition ID.")
	a.Describe(&i.DisplayName, "The human-readable label. Defaults to the agent's name.")
	a.Describe(&i.RoleID, "The organization role ID assumed by the agent's tasks. On creation, omitted or empty "+
		"uses the read-only default. On update, omitted preserves the role and empty selects the default. "+
		"Use `getOrganizationRole` to resolve an existing role, for example `Member`.")
	a.Describe(&i.CustomInstructions, "Instructions that shape how the agent behaves in the team's environment.")
	a.Describe(&i.EnabledIntegrations, "Integrations enabled for the agent's tasks. Omitted or null inherits all "+
		"organization integrations; an empty list enables none.")
	a.Describe(&i.Purpose, "A one-line description of the agent's purpose.")
	a.Describe(&i.Avatar, "A presentation token the console maps to an icon, not an uploaded asset.")
	a.Describe(&i.ApprovalMode, "The approval posture inherited by agent runs. Defaults to `auto`. `manual` pauses "+
		"every run for a person's approval, including scheduled and triggered runs.")
	a.SetDefault(&i.ApprovalMode, "auto")
}

type AgentDefinitionState struct {
	AgentDefinitionInput
}

func (*AgentDefinition) Create(
	ctx context.Context, req infer.CreateRequest[AgentDefinitionInput],
) (infer.CreateResponse[AgentDefinitionState], error) {
	if req.DryRun {
		return infer.CreateResponse[AgentDefinitionState]{Output: AgentDefinitionState{req.Inputs}}, nil
	}
	definition, err := config.GetClient(ctx).CreateAgentDefinition(
		ctx, req.Inputs.OrganizationName, req.Inputs.apiRequest(),
	)
	if err != nil {
		return infer.CreateResponse[AgentDefinitionState]{}, err
	}
	return infer.CreateResponse[AgentDefinitionState]{
		ID: definition.ID, Output: agentDefinitionState(req.Inputs.OrganizationName, definition),
	}, nil
}

func (*AgentDefinition) Update(
	ctx context.Context, req infer.UpdateRequest[AgentDefinitionInput, AgentDefinitionState],
) (infer.UpdateResponse[AgentDefinitionState], error) {
	if req.DryRun {
		return infer.UpdateResponse[AgentDefinitionState]{Output: AgentDefinitionState{req.Inputs}}, nil
	}
	definition, err := config.GetClient(ctx).UpdateAgentDefinition(
		ctx, req.Inputs.OrganizationName, req.ID, req.Inputs.apiRequest(),
	)
	if err != nil {
		return infer.UpdateResponse[AgentDefinitionState]{}, err
	}
	return infer.UpdateResponse[AgentDefinitionState]{
		Output: agentDefinitionState(req.Inputs.OrganizationName, definition),
	}, nil
}

func (*AgentDefinition) Read(
	ctx context.Context, req infer.ReadRequest[AgentDefinitionInput, AgentDefinitionState],
) (infer.ReadResponse[AgentDefinitionInput, AgentDefinitionState], error) {
	organization := req.State.OrganizationName
	if organization == "" {
		organization = req.Inputs.OrganizationName
	}
	organization, id, err := agentDefinitionIdentity(organization, req.ID)
	if err != nil {
		return infer.ReadResponse[AgentDefinitionInput, AgentDefinitionState]{}, err
	}
	definition, err := config.GetClient(ctx).GetActiveAgentDefinition(ctx, organization, id)
	if err != nil {
		return infer.ReadResponse[AgentDefinitionInput, AgentDefinitionState]{}, err
	}
	if definition == nil {
		return infer.ReadResponse[AgentDefinitionInput, AgentDefinitionState]{}, nil
	}
	state := agentDefinitionState(organization, definition)
	return infer.ReadResponse[AgentDefinitionInput, AgentDefinitionState]{
		ID: id, Inputs: state.AgentDefinitionInput, State: state,
	}, nil
}

func (*AgentDefinition) Delete(
	ctx context.Context, req infer.DeleteRequest[AgentDefinitionState],
) (infer.DeleteResponse, error) {
	organization, id, err := agentDefinitionIdentity(req.State.OrganizationName, req.ID)
	if err != nil {
		return infer.DeleteResponse{}, err
	}
	return infer.DeleteResponse{}, config.GetClient(ctx).DeleteAgentDefinition(ctx, organization, id)
}

// Imports include the organization so the UUID can be resolved without prior
// state. After import, the resource ID is the same durable UUID as on create.
func agentDefinitionIdentity(organization, id string) (string, string, error) {
	if strings.Contains(id, "/") {
		parts := strings.Split(id, "/")
		if len(parts) != 2 {
			return "", "", fmt.Errorf("%q is invalid, must be of the form organization/definitionId", id)
		}
		organization, id = parts[0], parts[1]
	}
	if organization == "" || id == "" {
		return "", "", fmt.Errorf("organization and definition ID are required; import with organization/definitionId")
	}
	return organization, id, nil
}

func (i AgentDefinitionInput) apiRequest() pulumiapi.AgentDefinitionRequest {
	request := pulumiapi.AgentDefinitionRequest{
		Name: i.Name, DisplayName: i.DisplayName, RoleID: i.RoleID, CustomInstructions: i.CustomInstructions,
		Purpose: i.Purpose, Avatar: i.Avatar,
	}
	if i.ApprovalMode != nil {
		request.ApprovalMode = string(*i.ApprovalMode)
	}
	if i.EnabledIntegrations != nil {
		request.EnabledIntegrations = make([]pulumiapi.AgentTaskIntegrationRef, len(*i.EnabledIntegrations))
		for n, integration := range *i.EnabledIntegrations {
			request.EnabledIntegrations[n] = pulumiapi.AgentTaskIntegrationRef{ID: integration.ID}
		}
	}
	return request
}

func agentDefinitionState(organization string, definition *pulumiapi.AgentDefinition) AgentDefinitionState {
	state := AgentDefinitionState{AgentDefinitionInput: AgentDefinitionInput{
		OrganizationName: organization, Name: definition.Name, DisplayName: definition.DisplayName,
		RoleID: definition.RoleID, CustomInstructions: definition.CustomInstructions,
		Purpose: definition.Purpose, Avatar: definition.Avatar,
	}}
	mode := AgentApprovalMode(definition.ApprovalMode)
	if mode == "" {
		mode = "auto"
	}
	state.ApprovalMode = &mode
	if definition.EnabledIntegrations != nil {
		integrations := make([]AgentDefinitionIntegration, len(definition.EnabledIntegrations))
		for n, integration := range definition.EnabledIntegrations {
			integrations[n] = AgentDefinitionIntegration{ID: integration.ID}
		}
		state.EnabledIntegrations = &integrations
	}
	return state
}
