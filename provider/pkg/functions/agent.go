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

package functions

import (
	"context"

	_ "embed" // for the generated example below

	"github.com/pulumi/pulumi-go-provider/infer"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/config"
)

// Generated from examples/get_agent_examples.yaml by make generate_examples.
//
//go:embed get_agent_examples.md
var getAgentExamples string

type GetAgentFunction struct{}

type GetAgentInput struct {
	Organization string `pulumi:"organization"`
	AgentID      string `pulumi:"agentId"`
}

type GetAgentOutput struct {
	Organization string `pulumi:"organization"`
	AgentID      string `pulumi:"agentId"`
	Name         string `pulumi:"name"`
	DisplayName  string `pulumi:"displayName"`
}

func (GetAgentFunction) Annotate(a infer.Annotator) {
	a.SetToken("index", "getAgent")
	a.Describe(&GetAgentFunction{}, "Looks up an active saved Custom Agent in Pulumi Cloud by its durable definition ID. "+
		"Pass the returned `agentId` to `delivery.Agent.agent`. Renaming the agent does not change this ID. "+
		"The provider's caller must be allowed to use the agent's role. "+
		"Deleted, missing, or invisible agents fail the lookup. "+
		"This lookup returns identity metadata only; it does not create a task or grant permission to run one.\n\n"+
		getAgentExamples)
}

func (i *GetAgentInput) Annotate(a infer.Annotator) {
	a.Describe(&i.Organization, "The Pulumi Cloud organization that owns the Custom Agent.")
	a.Describe(&i.AgentID, "The durable Custom Agent definition ID from Pulumi Cloud, not its name or an agent task ID.")
}

func (o *GetAgentOutput) Annotate(a infer.Annotator) {
	a.Describe(&o.Organization, "The Pulumi Cloud organization that owns the Custom Agent.")
	a.Describe(&o.AgentID, "The durable Custom Agent definition ID. Use this value in `delivery.Agent.agent`.")
	a.Describe(&o.Name, "The agent's current name. This is descriptive metadata, not its identity.")
	a.Describe(&o.DisplayName, "The agent's current display name.")
}

func (GetAgentFunction) Invoke(
	ctx context.Context,
	req infer.FunctionRequest[GetAgentInput],
) (infer.FunctionResponse[GetAgentOutput], error) {
	agent, err := config.GetClient(ctx).GetAgentDefinition(ctx, req.Input.Organization, req.Input.AgentID)
	if err != nil {
		return infer.FunctionResponse[GetAgentOutput]{}, err
	}
	return infer.FunctionResponse[GetAgentOutput]{Output: GetAgentOutput{
		Organization: req.Input.Organization,
		AgentID:      agent.ID,
		Name:         agent.Name,
		DisplayName:  agent.DisplayName,
	}}, nil
}
