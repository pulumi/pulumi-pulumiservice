package provider

import (
	"errors"
	"fmt"

	pbempty "github.com/golang/protobuf/ptypes/empty"
	pulumiapi "github.com/pierskarsenbarg/pulumi-apiclient"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
)

type PulumiServiceTeamResource struct {
	config PulumiServiceConfig
}

type PulumiServiceTeamInput struct {
	Type             string
	Name             string
	DisplayName      string
	Description      string
	OrganisationName string
	Members          []string
}

func (i *PulumiServiceTeamInput) ToPropertyMap() resource.PropertyMap {
	pm := resource.PropertyMap{}
	pm["type"] = resource.NewPropertyValue(i.Type)
	pm["name"] = resource.NewPropertyValue(i.Name)
	pm["displayName"] = resource.NewPropertyValue(i.DisplayName)
	pm["description"] = resource.NewPropertyValue(i.Description)
	pm["members"] = resource.NewPropertyValue(i.Members)
	pm["organisationName"] = resource.NewPropertyValue(i.OrganisationName)
	return pm
}

func (t *PulumiServiceTeamResource) ToPulumiServiceTeamInput(inputMap resource.PropertyMap) PulumiServiceTeamInput {
	input := PulumiServiceTeamInput{}

	if inputMap["name"].HasValue() && inputMap["name"].IsString() {
		input.Name = inputMap["name"].StringValue()
	}

	if inputMap["type"].HasValue() && inputMap["type"].IsString() {
		input.Type = inputMap["type"].StringValue()
	}

	if inputMap["displayName"].HasValue() && inputMap["displayName"].IsString() {
		input.DisplayName = inputMap["displayname"].StringValue()
	}

	if inputMap["description"].HasValue() && inputMap["description"].IsString() {
		input.Description = inputMap["description"].StringValue()
	}

	if inputMap["members"].HasValue() && inputMap["members"].IsArray() {
		for _, m := range inputMap["members"].ArrayValue() {
			if m.HasValue() && m.IsString() {
				input.Members = append(input.Members, m.StringValue())
			}
		}
	}

	if inputMap["organisationName"].HasValue() && inputMap["organisationName"].IsArray() {
		input.OrganisationName = inputMap["organisationName"].StringValue()
	}

	return input
}

func (t *PulumiServiceTeamResource) Name() string {
	return "pulumiservice:index:Team"
}

func (t *PulumiServiceTeamResource) Configure(config PulumiServiceConfig) {
	t.config = config
}

func (tr *PulumiServiceTeamResource) Check(req *pulumirpc.CheckRequest) (*pulumirpc.CheckResponse, error) {
	return nil, errors.New("Construct is not yet implemented")
}

func (tr *PulumiServiceTeamResource) Delete(req *pulumirpc.DeleteRequest) (*pbempty.Empty, error) {
	return nil, errors.New("Construct is not yet implemented")
}

func (tr *PulumiServiceTeamResource) Diff(req *pulumirpc.DiffRequest) (*pulumirpc.DiffResponse, error) {
	return nil, errors.New("Construct is not yet implemented")
}

func (tr *PulumiServiceTeamResource) Read(req *pulumirpc.ReadRequest) (*pulumirpc.ReadResponse, error) {
	return nil, errors.New("Construct is not yet implemented")
}

func (tr *PulumiServiceTeamResource) Update(req *pulumirpc.UpdateRequest) (*pulumirpc.UpdateResponse, error) {
	return nil, errors.New("Construct is not yet implemented")
}

func (tr *PulumiServiceTeamResource) Create(req *pulumirpc.CreateRequest) (*pulumirpc.CreateResponse, error) {
	inputs, err := plugin.UnmarshalProperties(req.GetProperties(), plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	inputsTeam := tr.ToPulumiServiceTeamInput(inputs)
	channelId, err := tr.createTeam(inputsTeam)
	if err != nil {
		return nil, fmt.Errorf("error creating team '%s': %s", inputsTeam.Name, err.Error())
	}

	outputStore := resource.PropertyMap{}
	outputStore["__inputs"] = resource.NewObjectProperty(inputs)

	outputProperties, err := plugin.MarshalProperties(
		outputStore,
		plugin.MarshalOptions{},
	)
	if err != nil {
		return nil, err
	}

	return &pulumirpc.CreateResponse{
		Id:         *channelId,
		Properties: outputProperties,
	}, nil
}

func (t *PulumiServiceTeamResource) createTeam(input PulumiServiceTeamInput) (*string, error) {
	token, err := t.config.getPulumiAccessToken()
	if err != nil {
		return nil, err
	}

	c := pulumiapi.NewClient(*token)
	team, err := c.CreateTeam(input.OrganisationName, input.Name, input.Type, input.DisplayName, input.Description)
	if err != nil {
		return nil, err
	}

	return &team.Name, nil
}
