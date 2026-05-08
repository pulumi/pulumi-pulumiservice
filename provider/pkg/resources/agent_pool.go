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
	"fmt"
	"strings"

	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/pulumi/pulumi/sdk/v3/go/property"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/config"
)

type AgentPool struct{}

var (
	_ infer.CustomCreate[AgentPoolInput, AgentPoolState] = &AgentPool{}
	_ infer.CustomUpdate[AgentPoolInput, AgentPoolState] = &AgentPool{}
	_ infer.CustomDelete[AgentPoolState]                 = &AgentPool{}
	_ infer.CustomRead[AgentPoolInput, AgentPoolState]   = &AgentPool{}
	_ infer.CustomStateMigrations[AgentPoolState]        = &AgentPool{}
)

func (*AgentPool) Annotate(a infer.Annotator) {
	a.Describe(&AgentPool{}, "Agent Pool for customer managed deployments.")
	a.SetToken("index", "AgentPool")
}

type AgentPoolInput struct {
	OrganizationName string `pulumi:"organizationName"     provider:"replaceOnChanges"`
	Name             string `pulumi:"name"`
	Description      string `pulumi:"description,optional"`
	ForceDestroy     bool   `pulumi:"forceDestroy,optional"`
}

func (i *AgentPoolInput) Annotate(a infer.Annotator) {
	a.Describe(&i.OrganizationName, "The organization's name.")
	a.Describe(&i.Name, "Name of the agent pool.")
	a.Describe(&i.Description, "Description of the agent pool.")
	a.Describe(
		&i.ForceDestroy,
		"Optional. Flag indicating whether to delete the agent pool even if stacks are configured to use it.",
	)
}

type AgentPoolState struct {
	AgentPoolInput
	AgentPoolID string `pulumi:"agentPoolId"`
	TokenValue  string `pulumi:"tokenValue" provider:"secret"`
}

func (s *AgentPoolState) Annotate(a infer.Annotator) {
	a.Describe(&s.AgentPoolID, "The agent pool identifier.")
	a.Describe(&s.TokenValue, "The agent pool's token's value.")
}

func (*AgentPool) Create(
	ctx context.Context,
	req infer.CreateRequest[AgentPoolInput],
) (infer.CreateResponse[AgentPoolState], error) {
	if req.DryRun {
		return infer.CreateResponse[AgentPoolState]{
			Output: AgentPoolState{AgentPoolInput: req.Inputs},
		}, nil
	}
	pool, err := config.GetClient(ctx).CreateAgentPool(
		ctx, req.Inputs.OrganizationName, req.Inputs.Name, req.Inputs.Description,
	)
	if err != nil {
		return infer.CreateResponse[AgentPoolState]{}, fmt.Errorf(
			"error creating agent pool %q: %w", req.Inputs.Name, err,
		)
	}
	return infer.CreateResponse[AgentPoolState]{
		ID: agentPoolResourceID(req.Inputs.OrganizationName, req.Inputs.Name, pool.ID),
		Output: AgentPoolState{
			AgentPoolInput: req.Inputs,
			AgentPoolID:    pool.ID,
			TokenValue:     pool.TokenValue,
		},
	}, nil
}

func (*AgentPool) Update(
	ctx context.Context,
	req infer.UpdateRequest[AgentPoolInput, AgentPoolState],
) (infer.UpdateResponse[AgentPoolState], error) {
	if req.DryRun {
		return infer.UpdateResponse[AgentPoolState]{
			Output: AgentPoolState{
				AgentPoolInput: req.Inputs,
				AgentPoolID:    req.State.AgentPoolID,
				TokenValue:     req.State.TokenValue,
			},
		}, nil
	}
	err := config.GetClient(ctx).UpdateAgentPool(
		ctx, req.State.AgentPoolID, req.Inputs.OrganizationName, req.Inputs.Name, req.Inputs.Description,
	)
	if err != nil {
		return infer.UpdateResponse[AgentPoolState]{}, fmt.Errorf(
			"error updating agent pool %q: %w", req.Inputs.Name, err,
		)
	}
	return infer.UpdateResponse[AgentPoolState]{
		Output: AgentPoolState{
			AgentPoolInput: req.Inputs,
			AgentPoolID:    req.State.AgentPoolID,
			TokenValue:     req.State.TokenValue,
		},
	}, nil
}

func (*AgentPool) Delete(
	ctx context.Context,
	req infer.DeleteRequest[AgentPoolState],
) (infer.DeleteResponse, error) {
	return infer.DeleteResponse{}, config.GetClient(ctx).DeleteAgentPool(
		ctx, req.State.AgentPoolID, req.State.OrganizationName, req.State.ForceDestroy,
	)
}

func (*AgentPool) Read(
	ctx context.Context,
	req infer.ReadRequest[AgentPoolInput, AgentPoolState],
) (infer.ReadResponse[AgentPoolInput, AgentPoolState], error) {
	orgName, _, agentPoolID, err := splitAgentPoolID(req.ID)
	if err != nil {
		return infer.ReadResponse[AgentPoolInput, AgentPoolState]{}, err
	}
	pool, err := config.GetClient(ctx).GetAgentPool(ctx, agentPoolID, orgName)
	if err != nil {
		return infer.ReadResponse[AgentPoolInput, AgentPoolState]{}, err
	}
	if pool == nil {
		return infer.ReadResponse[AgentPoolInput, AgentPoolState]{}, nil
	}
	inputs := AgentPoolInput{
		OrganizationName: orgName,
		Name:             pool.Name,
		Description:      pool.Description,
		ForceDestroy:     req.State.ForceDestroy,
	}
	tokenValue := pool.TokenValue
	if tokenValue == "" {
		// The Get-agent-pool API does not return the token value; carry the
		// existing secret from state so refresh does not erase it.
		tokenValue = req.State.TokenValue
	}
	return infer.ReadResponse[AgentPoolInput, AgentPoolState]{
		ID:     req.ID,
		Inputs: inputs,
		State: AgentPoolState{
			AgentPoolInput: inputs,
			AgentPoolID:    agentPoolID,
			TokenValue:     tokenValue,
		},
	}, nil
}

// StateMigrations adapts state written by the pre-infer AgentPool resource:
//   - the legacy framework embedded a "__inputs" outputs property which infer
//     rejects as an unrecognized field;
//   - the legacy code wrote the agent pool ID under the key "agentPoolID"
//     (capital D) even though the schema declared it as "agentPoolId".
func (*AgentPool) StateMigrations(context.Context) []infer.StateMigrationFunc[AgentPoolState] {
	return []infer.StateMigrationFunc[AgentPoolState]{
		infer.StateMigration(migrateAgentPoolLegacyState),
	}
}

func migrateAgentPoolLegacyState(
	_ context.Context, old property.Map,
) (infer.MigrationResult[AgentPoolState], error) {
	_, hasLegacyInputs := old.GetOk("__inputs")
	_, hasMisnamedID := old.GetOk("agentPoolID")
	if !hasLegacyInputs && !hasMisnamedID {
		return infer.MigrationResult[AgentPoolState]{}, nil
	}
	state := AgentPoolState{}
	if v, ok := old.GetOk("organizationName"); ok && v.IsString() {
		state.OrganizationName = v.AsString()
	}
	if v, ok := old.GetOk("name"); ok && v.IsString() {
		state.Name = v.AsString()
	}
	if v, ok := old.GetOk("description"); ok && v.IsString() {
		state.Description = v.AsString()
	}
	if v, ok := old.GetOk("forceDestroy"); ok && v.IsBool() {
		state.ForceDestroy = v.AsBool()
	}
	if v, ok := old.GetOk("agentPoolId"); ok && v.IsString() {
		state.AgentPoolID = v.AsString()
	} else if v, ok := old.GetOk("agentPoolID"); ok && v.IsString() {
		state.AgentPoolID = v.AsString()
	}
	if v, ok := old.GetOk("tokenValue"); ok && v.IsString() {
		state.TokenValue = v.AsString()
	}
	return infer.MigrationResult[AgentPoolState]{Result: &state}, nil
}

func agentPoolResourceID(orgName, name, agentPoolID string) string {
	return fmt.Sprintf("%s/%s/%s", orgName, name, agentPoolID)
}

func splitAgentPoolID(id string) (orgName, name, agentPoolID string, err error) {
	s := strings.Split(id, "/")
	if len(s) != 3 {
		return "", "", "", fmt.Errorf(
			"%q is invalid, must be in the format: organization/name/agentPoolID", id,
		)
	}
	return s[0], s[1], s[2], nil
}
