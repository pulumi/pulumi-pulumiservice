package provider

import (
	"fmt"
	"strings"

	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/pulumi/pulumi/sdk/v3/go/common/util/contract"
)

type AgentPool struct{}

// We list out the expected capabilities of AgentPool, since infer is not capable of
// determining which capabilities we intend to implement and which capabilities we
// intentionally ignored.
var (
	// Life-cycle participation
	_ infer.CustomResource[AgentPoolInput, AgentPoolState] = (*AgentPool)(nil)
	_ infer.CustomUpdate[AgentPoolInput, AgentPoolState]   = (*AgentPool)(nil)
	_ infer.CustomDelete[AgentPoolState]                   = (*AgentPool)(nil)
	_ infer.CustomRead[AgentPoolInput, AgentPoolState]     = (*AgentPool)(nil)

	// Schema documentation
	_ infer.Annotated = (*AgentPool)(nil)
	_ infer.Annotated = (*AgentPoolInput)(nil)
	_ infer.Annotated = (*AgentPoolState)(nil)

	// Secret values
	_ infer.ExplicitDependencies[AgentPoolInput, AgentPoolState] = (*AgentPool)(nil)
)

func (p *AgentPool) Annotate(a infer.Annotator) {
	a.Describe(p, "Agent Pool for customer manager deployments")
}

type AgentPoolInput struct {
	Name             string `pulumi:"name"`
	Description      string `pulumi:"description,optional"`
	OrganizationName string `pulumi:"organizationName" provider:"replaceOnChanges"`
}

func (p *AgentPoolInput) Annotate(a infer.Annotator) {
	a.Describe(&p.Name, "Name of the agent pool.")
	a.Describe(&p.Description, "Description of the agent pool.")
	a.Describe(&p.OrganizationName, "The organization's name.")
}

type AgentPoolState struct {
	AgentPoolInput
	AgentPoolID string `pulumi:"agentPoolId"`
	TokenValue  string `pulumi:"tokenValue"`
}

func (p *AgentPoolState) Annotate(a infer.Annotator) {
	a.Describe(&p.AgentPoolID, "The agent pool identifier.")
	a.Describe(&p.TokenValue, "The agent pool's token's value.")
}

func (*AgentPool) WireDependencies(f infer.FieldSelector, args *AgentPoolInput, state *AgentPoolState) {
	f.OutputField(&state.TokenValue).AlwaysSecret()
}

func (*AgentPool) Delete(ctx p.Context, id string, props AgentPoolState) error {
	return GetConfig(ctx).Client.DeleteAgentPool(ctx, props.AgentPoolID, props.OrganizationName)
}

func (*AgentPool) Create(
	ctx p.Context, name string, input AgentPoolInput, preview bool,
) (string, AgentPoolState, error) {
	if preview {
		return "", AgentPoolState{}, nil
	}

	agentPool, err := GetConfig(ctx).Client.CreateAgentPool(ctx, input.OrganizationName, input.Name, input.Description)
	if err != nil {
		return "", AgentPoolState{}, fmt.Errorf("error creating agent pool '%s': %w", input.Name, err)
	}

	id := fmt.Sprintf(input.OrganizationName + "/" + input.Name + "/" + agentPool.ID)

	return id, AgentPoolState{
		AgentPoolInput: input,
		AgentPoolID:    agentPool.ID,
		TokenValue:     agentPool.TokenValue,
	}, nil
}

func (*AgentPool) Update(
	ctx p.Context, id string, olds AgentPoolState, news AgentPoolInput, preview bool,
) (AgentPoolState, error) {
	if preview {
		return AgentPoolState{}, nil
	}

	contract.Assertf(olds.OrganizationName == news.OrganizationName,
		"Changing the org name should be a replace")

	// The only thing that can actually change here is name and description
	err := GetConfig(ctx).Client.UpdateAgentPool(ctx,
		olds.AgentPoolID, olds.OrganizationName,
		news.Name, news.Description)
	if err != nil {
		return AgentPoolState{}, err
	}

	olds.Name = news.Name
	olds.Description = news.Description
	return olds, nil
}

func (*AgentPool) Read(
	ctx p.Context, id string, _ AgentPoolInput, _ AgentPoolState,
) (string, AgentPoolInput, AgentPoolState, error) {
	orgName, _, agentPoolId, err := splitAgentPoolId(id)
	if err != nil {
		return "", AgentPoolInput{}, AgentPoolState{}, err
	}

	agentPool, err := GetConfig(ctx).Client.GetAgentPool(ctx, agentPoolId, orgName)
	if err != nil {
		return "", AgentPoolInput{}, AgentPoolState{}, err
	}

	inputs := AgentPoolInput{
		Name:             agentPool.Name,
		Description:      agentPool.Description,
		OrganizationName: orgName,
	}

	outputs := AgentPoolState{
		AgentPoolInput: inputs,
		AgentPoolID:    agentPool.ID,
		TokenValue:     agentPool.TokenValue,
	}

	return id, inputs, outputs, nil
}

func splitAgentPoolId(id string) (string, string, string, error) {
	// format: organization/name/agentPoolId
	s := strings.Split(id, "/")
	if len(s) != 3 {
		return "", "", "", fmt.Errorf("%q is invalid, must be in the format: organization/name/agentPoolId", id)
	}
	return s[0], s[1], s[2], nil
}
