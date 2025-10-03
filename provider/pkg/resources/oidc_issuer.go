package resources

import (
	"context"
	"fmt"
	"strings"

	pbempty "google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/util"
)

// PulumiServiceOidcIssuerResource manages OIDC issuer resources in Pulumi Service.
type PulumiServiceOidcIssuerResource struct {
	Client pulumiapi.OidcClient
}

// AuthPolicyDecision represents the decision type for an authentication policy.
type AuthPolicyDecision string

const (
	// AuthPolicyDecisionAllow indicates the policy allows access.
	AuthPolicyDecisionAllow = AuthPolicyDecision("allow")
	// AuthPolicyDecisionDeny indicates the policy denies access.
	AuthPolicyDecisionDeny = AuthPolicyDecision("deny")
)

type AuthPolicyTokenType string

const (
	AuthPolicyTokenTypeOrganization     = AuthPolicyTokenType("organization")
	AuthPolicyTokenTypePersonal         = AuthPolicyTokenType("personal")
	AuthPolicyTokenTypeTeam             = AuthPolicyTokenType("team")
	AuthPolicyTokenTypeDeploymentRunner = AuthPolicyTokenType("runner")
)

type AuthPolicyPermissionLevel string

const (
	AuthPolicyPermissionLevelStandard = AuthPolicyPermissionLevel("standard")
	AuthPolicyPermissionLevelAdmin    = AuthPolicyPermissionLevel("admin")
)

type PulumiServiceAuthPolicyDefinition struct {
	Decision              AuthPolicyDecision
	TokenType             AuthPolicyTokenType
	TeamName              *string
	UserLogin             *string
	RunnerID              *string
	AuthorizedPermissions []AuthPolicyPermissionLevel
	Rules                 map[string]string
}

type PulumiServiceOidcIssuerInput struct {
	Organization         string
	Name                 string
	URL                  string
	MaxExpirationSeconds *int64
	Thumbprints          []string
	Policies             []PulumiServiceAuthPolicyDefinition
}

type PulumiServiceOidcIssuerProperties struct {
	PulumiServiceOidcIssuerInput
}

func GenerateOidcIssuerProperties(
	input PulumiServiceOidcIssuerInput,
	issuer *pulumiapi.OidcIssuerRegistrationResponse,
	authPolicy *pulumiapi.AuthPolicy,
) (outputs *structpb.Struct, inputs *structpb.Struct, err error) {
	inputMap := input.toPropertyMap()

	outputMap := inputMap
	outputMap["thumbprints"] = resource.NewPropertyValue(issuer.Thumbprints)
	outputMap["policies"] = apiPoliciesToPropertyValue(authPolicy.Definition)

	inputs, err = plugin.MarshalProperties(inputMap, util.StandardMarshal)
	if err != nil {
		return nil, nil, err
	}

	outputs, err = plugin.MarshalProperties(outputMap, util.StandardMarshal)
	if err != nil {
		return nil, nil, err
	}

	return outputs, inputs, err
}

func (oir *PulumiServiceOidcIssuerResource) ToPulumiServiceOidcIssuerInput(
	inputMap resource.PropertyMap,
) PulumiServiceOidcIssuerInput {
	input := PulumiServiceOidcIssuerInput{}

	input.Organization = inputMap["organization"].StringValue()
	input.Name = inputMap["name"].StringValue()
	input.URL = inputMap["url"].StringValue()
	if inputMap["maxExpirationSeconds"].HasValue() {
		seconds := int64(inputMap["maxExpirationSeconds"].NumberValue())
		input.MaxExpirationSeconds = &seconds
	}
	if inputMap["thumbprints"].HasValue() {
		thumbprints := []string{}
		for _, thumb := range inputMap["thumbprints"].ArrayValue() {
			thumbprints = append(thumbprints, thumb.StringValue())
		}
		input.Thumbprints = thumbprints
	}
	if inputMap["policies"].HasValue() {
		policies := []PulumiServiceAuthPolicyDefinition{}
		for _, policy := range inputMap["policies"].ArrayValue() {
			policies = append(policies, propertyMapToPolicy(policy.ObjectValue()))
		}
		input.Policies = policies
	}

	return input
}

func (oir *PulumiServiceOidcIssuerResource) Name() string {
	return "pulumiservice:index:OidcIssuer"
}

func (oir *PulumiServiceOidcIssuerResource) Diff(req *pulumirpc.DiffRequest) (*pulumirpc.DiffResponse, error) {
	replaceProperties := map[string]bool{
		"organization": true,
		"url":          true,
	}

	return util.StandardDiff(req, replaceProperties)
}

func (oir *PulumiServiceOidcIssuerResource) Delete(req *pulumirpc.DeleteRequest) (*pbempty.Empty, error) {
	ctx := context.Background()
	organization, issuerID, err := splitIssuerID(req.Id)
	if err != nil {
		return nil, err
	}

	err = oir.Client.DeleteOidcIssuer(ctx, *organization, *issuerID)
	if err != nil {
		return nil, err
	}

	return &pbempty.Empty{}, nil
}

func (oir *PulumiServiceOidcIssuerResource) Create(req *pulumirpc.CreateRequest) (*pulumirpc.CreateResponse, error) {
	ctx := context.Background()
	inputMap, err := plugin.UnmarshalProperties(req.GetProperties(), util.StandardUnmarshal)
	if err != nil {
		return nil, err
	}

	// Create OIDC Issuer itself
	input := oir.ToPulumiServiceOidcIssuerInput(inputMap)
	registerResponse, err := oir.Client.RegisterOidcIssuer(ctx, input.Organization, input.toCreateRequest())
	if err != nil {
		return nil, fmt.Errorf("error creating oidc issuer '%s': %s", input.Name, err.Error())
	}

	// Retrieve policy ID
	authPolicy, err := oir.Client.GetAuthPolicies(ctx, input.Organization, registerResponse.ID)
	if err != nil {
		return nil, fmt.Errorf(
			"error retrieving auth policies for oidc issuer '%s': %s",
			registerResponse.ID,
			err.Error(),
		)
	}

	// If user has provided policies, update with those, otherwise use the default one
	if len(input.Policies) > 0 {
		request := policiesToAPIRequest(input.Policies)
		authPolicy, err = oir.Client.UpdateAuthPolicies(ctx, input.Organization, authPolicy.ID, request)
		if err != nil {
			// To prevent resource being stuck in limbo, best-effort delete the issuer if policies were invalid
			_ = oir.Client.DeleteOidcIssuer(ctx, input.Organization, registerResponse.ID)

			return nil, fmt.Errorf(
				"error updating auth policies for oidc issuer '%s': %s",
				registerResponse.ID,
				err.Error(),
			)
		}
	}

	outputProperties, _, err := GenerateOidcIssuerProperties(input, registerResponse, authPolicy)
	if err != nil {
		return nil, err
	}

	return &pulumirpc.CreateResponse{
		Id:         fmt.Sprintf("%s/%s", input.Organization, registerResponse.ID),
		Properties: outputProperties,
	}, nil
}

func (oir *PulumiServiceOidcIssuerResource) Check(req *pulumirpc.CheckRequest) (*pulumirpc.CheckResponse, error) {
	// This conversion to and from is needed to sort policies on input
	inputMap, err := plugin.UnmarshalProperties(req.GetNews(), util.StandardUnmarshal)
	if err != nil {
		return nil, err
	}
	input := oir.ToPulumiServiceOidcIssuerInput(inputMap)
	inputs, err := plugin.MarshalProperties(input.toPropertyMap(), util.StandardMarshal)
	if err != nil {
		return nil, err
	}
	return &pulumirpc.CheckResponse{Inputs: inputs, Failures: nil}, nil
}

func (oir *PulumiServiceOidcIssuerResource) Update(req *pulumirpc.UpdateRequest) (*pulumirpc.UpdateResponse, error) {
	ctx := context.Background()
	inputMap, err := plugin.UnmarshalProperties(req.GetNews(), util.StandardUnmarshal)
	if err != nil {
		return nil, err
	}

	_, issuerID, err := splitIssuerID(req.Id)
	if err != nil {
		return nil, err
	}

	// Update OIDC Issuer itself
	input := oir.ToPulumiServiceOidcIssuerInput(inputMap)
	updateResponse, err := oir.Client.UpdateOidcIssuer(ctx, input.Organization, *issuerID, input.toUpdateRequest())
	if err != nil {
		return nil, fmt.Errorf("error creating oidc issuer '%s': %s", input.Name, err.Error())
	}

	// Retrieve policy ID
	authPolicy, err := oir.Client.GetAuthPolicies(ctx, input.Organization, *issuerID)
	if err != nil {
		return nil, fmt.Errorf("error retrieving auth policies for oidc issuer '%s': %s", *issuerID, err.Error())
	}

	// If user has provided policies, update with those, otherwise use the default one
	if len(input.Policies) > 0 {
		request := policiesToAPIRequest(input.Policies)
		authPolicy, err = oir.Client.UpdateAuthPolicies(ctx, input.Organization, authPolicy.ID, request)
		if err != nil {
			return nil, fmt.Errorf("error updating auth policies for oidc issuer '%s': %s", *issuerID, err.Error())
		}
	}

	outputProperties, _, err := GenerateOidcIssuerProperties(input, updateResponse, authPolicy)
	if err != nil {
		return nil, err
	}

	return &pulumirpc.UpdateResponse{
		Properties: outputProperties,
	}, nil
}

func (oir *PulumiServiceOidcIssuerResource) Read(req *pulumirpc.ReadRequest) (*pulumirpc.ReadResponse, error) {
	ctx := context.Background()

	organization, issuerID, err := splitIssuerID(req.Id)
	if err != nil {
		return nil, err
	}

	readResponse, err := oir.Client.GetOidcIssuer(ctx, *organization, *issuerID)
	if err != nil {
		return nil, err
	}
	if readResponse == nil {
		return &pulumirpc.ReadResponse{}, nil
	}

	authPolicy, err := oir.Client.GetAuthPolicies(ctx, *organization, *issuerID)
	if err != nil {
		return nil, fmt.Errorf("error retrieving auth policies for oidc issuer '%s': %s", *issuerID, err.Error())
	}

	input := PulumiServiceOidcIssuerInput{
		Organization:         *organization,
		Name:                 readResponse.Name,
		URL:                  readResponse.URL,
		MaxExpirationSeconds: readResponse.MaxExpiration,
		Thumbprints:          readResponse.Thumbprints,
		Policies:             apiPoliciesToInputs(authPolicy.Definition),
	}

	outputProperties, inputs, err := GenerateOidcIssuerProperties(input, readResponse, authPolicy)
	if err != nil {
		return nil, err
	}

	return &pulumirpc.ReadResponse{
		Id:         req.GetId(),
		Properties: outputProperties,
		Inputs:     inputs,
	}, nil
}

func (input PulumiServiceOidcIssuerInput) toPropertyMap() resource.PropertyMap {
	inputMap := resource.PropertyMap{}
	inputMap["organization"] = resource.NewPropertyValue(input.Organization)
	inputMap["name"] = resource.NewPropertyValue(input.Name)
	inputMap["url"] = resource.NewPropertyValue(input.URL)
	if input.MaxExpirationSeconds != nil {
		inputMap["maxExpirationSeconds"] = resource.NewPropertyValue(input.MaxExpirationSeconds)
	}
	if len(input.Thumbprints) > 0 {
		inputMap["thumbprints"] = resource.NewPropertyValue(input.Thumbprints)
	}
	if len(input.Policies) > 0 {
		policyValues := []resource.PropertyValue{}
		for _, policy := range input.Policies {
			propertyMap := policy.toPropertyMap()
			policyValues = append(policyValues, resource.NewObjectProperty(propertyMap))
		}
		inputMap["policies"] = resource.NewArrayProperty(policyValues)
	}
	return inputMap
}

func (policy PulumiServiceAuthPolicyDefinition) toPropertyMap() resource.PropertyMap {
	propertyMap := resource.PropertyMap{}
	propertyMap["decision"] = resource.NewPropertyValue(policy.Decision)
	propertyMap["tokenType"] = resource.NewPropertyValue(policy.TokenType)
	if policy.TeamName != nil {
		propertyMap["teamName"] = resource.NewPropertyValue(policy.TeamName)
	}
	if policy.UserLogin != nil {
		propertyMap["userLogin"] = resource.NewPropertyValue(policy.UserLogin)
	}
	if policy.RunnerID != nil {
		propertyMap["runnerID"] = resource.NewPropertyValue(policy.RunnerID)
	}
	if len(policy.AuthorizedPermissions) > 0 {
		propertyMap["authorizedPermissions"] = resource.NewPropertyValue(policy.AuthorizedPermissions)
	}
	propertyMap["rules"] = resource.NewPropertyValue(policy.Rules)

	return propertyMap
}

func apiPoliciesToPropertyValue(apiPolicies []*pulumiapi.AuthPolicyDefinition) resource.PropertyValue {
	policyValues := []resource.PropertyValue{}
	for _, policy := range apiPoliciesToInputs(apiPolicies) {
		propertyMap := policy.toPropertyMap()
		policyValues = append(policyValues, resource.NewObjectProperty(propertyMap))
	}
	return resource.NewArrayProperty(policyValues)
}

func (input *PulumiServiceOidcIssuerInput) toCreateRequest() pulumiapi.OidcIssuerRegistrationRequest {
	return pulumiapi.OidcIssuerRegistrationRequest{
		Name:          input.Name,
		URL:           input.URL,
		Thumbprints:   input.Thumbprints,
		MaxExpiration: input.MaxExpirationSeconds,
	}
}

func (input *PulumiServiceOidcIssuerInput) toUpdateRequest() pulumiapi.OidcIssuerUpdateRequest {
	return pulumiapi.OidcIssuerUpdateRequest{
		Name:          &input.Name,
		Thumbprints:   &input.Thumbprints,
		MaxExpiration: input.MaxExpirationSeconds,
	}
}

func propertyMapToPolicy(policyObject resource.PropertyMap) PulumiServiceAuthPolicyDefinition {
	var teamName *string
	if policyObject["teamName"].HasValue() {
		value := policyObject["teamName"].StringValue()
		teamName = &value
	}

	var userLogin *string
	if policyObject["userLogin"].HasValue() {
		value := policyObject["userLogin"].StringValue()
		userLogin = &value
	}

	var runnerID *string
	if policyObject["runnerID"].HasValue() {
		value := policyObject["runnerID"].StringValue()
		runnerID = &value
	}

	authorizedPermissions := []AuthPolicyPermissionLevel{}
	if policyObject["authorizedPermissions"].HasValue() {
		for _, permission := range policyObject["authorizedPermissions"].ArrayValue() {
			authorizedPermissions = append(authorizedPermissions, AuthPolicyPermissionLevel(permission.StringValue()))
		}
	}

	rules := map[string]string{}
	if policyObject["rules"].HasValue() {
		for k, v := range policyObject["rules"].ObjectValue() {
			key := string(k)
			rules[key] = v.StringValue()
		}
	}

	return PulumiServiceAuthPolicyDefinition{
		Decision:              AuthPolicyDecision(policyObject["decision"].StringValue()),
		TokenType:             AuthPolicyTokenType(policyObject["tokenType"].StringValue()),
		TeamName:              teamName,
		UserLogin:             userLogin,
		RunnerID:              runnerID,
		AuthorizedPermissions: authorizedPermissions,
		Rules:                 rules,
	}
}

func policiesToAPIRequest(policies []PulumiServiceAuthPolicyDefinition) pulumiapi.AuthPolicyUpdateRequest {
	apiPolicies := []pulumiapi.AuthPolicyDefinition{}

	for _, policy := range policies {
		apiPolicies = append(apiPolicies, pulumiapi.AuthPolicyDefinition{
			Decision:              string(policy.Decision),
			TokenType:             string(policy.TokenType),
			TeamName:              policy.TeamName,
			UserLogin:             policy.UserLogin,
			RunnerID:              policy.RunnerID,
			AuthorizedPermissions: permissionsToAPI(policy.AuthorizedPermissions),
			Rules:                 policy.Rules,
		})
	}

	return pulumiapi.AuthPolicyUpdateRequest{
		Definition: apiPolicies,
	}
}

func apiPoliciesToInputs(policies []*pulumiapi.AuthPolicyDefinition) []PulumiServiceAuthPolicyDefinition {
	inputPolicies := []PulumiServiceAuthPolicyDefinition{}
	for _, policy := range policies {
		inputPolicies = append(inputPolicies, apiPolicyToInput(*policy))
	}
	return inputPolicies
}

func apiPolicyToInput(policy pulumiapi.AuthPolicyDefinition) PulumiServiceAuthPolicyDefinition {
	return PulumiServiceAuthPolicyDefinition{
		Decision:              AuthPolicyDecision(policy.Decision),
		TokenType:             AuthPolicyTokenType(policy.TokenType),
		TeamName:              policy.TeamName,
		UserLogin:             policy.UserLogin,
		RunnerID:              policy.RunnerID,
		AuthorizedPermissions: permissionsToInput(policy.AuthorizedPermissions),
		Rules:                 policy.Rules,
	}
}

func permissionsToAPI(permissions []AuthPolicyPermissionLevel) []string {
	apiPermissions := []string{}
	for _, permission := range permissions {
		apiPermissions = append(apiPermissions, string(permission))
	}
	return apiPermissions
}

func permissionsToInput(permissions []string) []AuthPolicyPermissionLevel {
	inputPermissions := []AuthPolicyPermissionLevel{}
	for _, permission := range permissions {
		inputPermissions = append(inputPermissions, AuthPolicyPermissionLevel(permission))
	}
	return inputPermissions
}

func splitIssuerID(id string) (organization *string, issuerID *string, err error) {
	splitID := strings.Split(id, "/")
	if len(splitID) != 2 {
		return nil, nil, fmt.Errorf("error splitting resource id '%s'", id)
	}
	return &splitID[0], &splitID[1], nil
}
