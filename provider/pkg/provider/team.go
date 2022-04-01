package provider

import (
	"errors"
	"fmt"
	"strings"

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
		input.DisplayName = inputMap["displayName"].StringValue()
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

	if inputMap["organisationName"].HasValue() && inputMap["organisationName"].IsString() {
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
	return &pulumirpc.CheckResponse{Inputs: req.News, Failures: nil}, nil
}

func (tr *PulumiServiceTeamResource) Delete(req *pulumirpc.DeleteRequest) (*pbempty.Empty, error) {
	err := tr.deleteTeam(req.Id)
	if err != nil {
		return &pbempty.Empty{}, err
	}

	return &pbempty.Empty{}, nil
}

func (tr *PulumiServiceTeamResource) Diff(req *pulumirpc.DiffRequest) (*pulumirpc.DiffResponse, error) {
	olds, err := plugin.UnmarshalProperties(req.GetOlds(), plugin.MarshalOptions{KeepUnknowns: false, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	news, err := plugin.UnmarshalProperties(req.GetNews(), plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: false})
	if err != nil {
		return nil, err
	}

	diffs := olds["__inputs"].ObjectValue().Diff(news)
	if diffs == nil {
		return &pulumirpc.DiffResponse{
			Changes:             pulumirpc.DiffResponse_DIFF_NONE,
			Replaces:            []string{},
			Stables:             []string{},
			DeleteBeforeReplace: false,
		}, nil
	}

	changes := pulumirpc.DiffResponse_DIFF_NONE
	if diffs.Changed("type") ||
		diffs.Changed("name") ||
		diffs.Changed("displayName") ||
		diffs.Changed("description") ||
		diffs.Changed("members") ||
		diffs.Changed("organisationName") {
		changes = pulumirpc.DiffResponse_DIFF_SOME
	}

	return &pulumirpc.DiffResponse{
		Changes:             changes,
		Replaces:            []string{},
		Stables:             []string{},
		DeleteBeforeReplace: false,
	}, nil
}

func (tr *PulumiServiceTeamResource) Read(req *pulumirpc.ReadRequest) (*pulumirpc.ReadResponse, error) {
	return nil, errors.New("Read construct is not yet implemented")
}

func (tr *PulumiServiceTeamResource) Update(req *pulumirpc.UpdateRequest) (*pulumirpc.UpdateResponse, error) {
	inputsOld, err := plugin.UnmarshalProperties(req.GetOlds(), plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
	if err != nil {
		return nil, err
	}
	inputsNew, err := plugin.UnmarshalProperties(req.GetNews(), plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	teamOld := tr.ToPulumiServiceTeamInput(inputsOld["__inputs"].ObjectValue())
	teamNew := tr.ToPulumiServiceTeamInput(inputsNew)

	inputsChanged := teamOld
	if teamOld.Description != teamNew.Description || teamOld.DisplayName != teamNew.Description {

		inputsChanged.Description = teamNew.Description
		inputsChanged.DisplayName = teamNew.DisplayName

		tr.updateTeam(inputsChanged)
	}

	if !Equal(teamOld.Members, teamNew.Members) {
		inputsChanged.Members = teamNew.Members
		for _, usernameToDelete := range teamOld.Members {
			if !InSlice(usernameToDelete, teamNew.Members) {
				tr.deleteFromTeam(teamOld.OrganisationName, teamOld.Name, usernameToDelete)
			}
		}

		for _, usernameToAdd := range teamNew.Members {
			if !InSlice(usernameToAdd, teamOld.Members) {
				tr.addToTeam(teamOld.OrganisationName, teamOld.Name, usernameToAdd)
			}
		}
	}

	outputStore := resource.PropertyMap{}
	outputStore["__inputs"] = resource.NewObjectProperty(inputsChanged.ToPropertyMap())

	outputProperties, err := plugin.MarshalProperties(
		outputStore,
		plugin.MarshalOptions{},
	)
	if err != nil {
		return nil, err
	}
	return &pulumirpc.UpdateResponse{
		Properties: outputProperties,
	}, nil
}

func (tr *PulumiServiceTeamResource) Create(req *pulumirpc.CreateRequest) (*pulumirpc.CreateResponse, error) {
	inputs, err := plugin.UnmarshalProperties(req.GetProperties(), plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	inputsTeam := tr.ToPulumiServiceTeamInput(inputs)
	team, err := tr.createTeam(inputsTeam)
	if err != nil {
		return nil, fmt.Errorf("error creating team '%s': %s", inputsTeam.Name, err.Error())
	}

	for _, memberToAdd := range inputsTeam.Members {
		err = tr.addToTeam(inputsTeam.OrganisationName, inputsTeam.Name, memberToAdd)
		if err != nil {
			return nil, err
		}
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
		Id:         *team,
		Properties: outputProperties,
	}, nil
}

func (t *PulumiServiceTeamResource) updateTeam(input PulumiServiceTeamInput) error {
	token, err := t.config.getPulumiAccessToken()
	if err != nil {
		return err
	}

	url, err := t.config.getPulumiServiceUrl()
	if err != nil {
		return err
	}

	c := pulumiapi.NewClient(*token, *url)
	_, err = c.UpdateTeam(input.OrganisationName, input.Name, input.DisplayName, input.Description)
	if err != nil {
		return err
	}
	return nil
}

func (t *PulumiServiceTeamResource) createTeam(input PulumiServiceTeamInput) (*string, error) {
	token, err := t.config.getPulumiAccessToken()
	if err != nil {
		return nil, err
	}

	url, err := t.config.getPulumiServiceUrl()
	if err != nil {
		return nil, err
	}

	c := pulumiapi.NewClient(*token, *url)
	_, err = c.CreateTeam(input.OrganisationName, input.Name, input.Type, input.DisplayName, input.Description)
	if err != nil {
		return nil, err
	}

	teamUrn := fmt.Sprintf("%s/%s", input.OrganisationName, input.Name)
	return &teamUrn, nil
}

func (t *PulumiServiceTeamResource) deleteFromTeam(orgName string, teamName string, userName string) error {

	if len(orgName) == 0 {
		return errors.New("orgname must not be empty")
	}

	if len(teamName) == 0 {
		return errors.New("teamname must not be empty")
	}

	if len(userName) == 0 {
		return errors.New("username must not be empty")
	}

	token, err := t.config.getPulumiAccessToken()
	if err != nil {
		return err
	}

	url, err := t.config.getPulumiServiceUrl()
	if err != nil {
		return err
	}

	c := pulumiapi.NewClient(*token, *url)

	err = c.DeleteMemberFromTeam(orgName, teamName, userName)
	if err != nil {
		return err
	}
	return nil
}

func (t *PulumiServiceTeamResource) addToTeam(orgName string, teamName string, userName string) error {
	if len(orgName) == 0 {
		return errors.New("orgname must not be empty")
	}

	if len(teamName) == 0 {
		return errors.New("teamname must not be empty")
	}

	if len(userName) == 0 {
		return errors.New("username must not be empty")
	}
	token, err := t.config.getPulumiAccessToken()
	if err != nil {
		return err
	}

	url, err := t.config.getPulumiServiceUrl()
	if err != nil {
		return err
	}

	c := pulumiapi.NewClient(*token, *url)

	err = c.AddMemberToTeam(orgName, teamName, userName)
	if err != nil {
		return err
	}
	return nil
}

func (t *PulumiServiceTeamResource) deleteTeam(id string) error {
	token, err := t.config.getPulumiAccessToken()
	if err != nil {
		return err
	}

	url, err := t.config.getPulumiServiceUrl()
	if err != nil {
		return err
	}

	s := strings.Split(id, "/")

	c := pulumiapi.NewClient(*token, *url)
	err = c.DeleteTeam(s[0], s[1])
	if err != nil {
		return err
	}
	return nil
}
