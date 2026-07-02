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

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/config"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
)

type OidcIssuer struct{}

var (
	_ infer.CustomCreate[OidcIssuerInput, OidcIssuerState] = &OidcIssuer{}
	_ infer.CustomUpdate[OidcIssuerInput, OidcIssuerState] = &OidcIssuer{}
	_ infer.CustomDelete[OidcIssuerState]                  = &OidcIssuer{}
	_ infer.CustomRead[OidcIssuerInput, OidcIssuerState]   = &OidcIssuer{}
)

func (*OidcIssuer) Annotate(a infer.Annotator) {
	a.Describe(&OidcIssuer{}, "Register an OIDC Provider to establish a trust relationship between third-party "+
		"systems like GitHub Actions and Pulumi Cloud, obviating the need to store a hard-coded Pulumi Cloud token "+
		"in systems that need to run Pulumi commands or consume Pulumi Cloud APIs. Instead of a hard-coded, static "+
		"token that must be manually rotated, trusted systems are granted temporary Pulumi Cloud tokens on an "+
		"as-needed basis, which is more secure than static tokens.")
	a.SetToken("index", "OidcIssuer")
}

// AuthPolicyDecision is the rule type of a policy definition.
type AuthPolicyDecision string

const (
	AuthPolicyDecisionDeny  AuthPolicyDecision = "deny"
	AuthPolicyDecisionAllow AuthPolicyDecision = "allow"
)

func (AuthPolicyDecision) Values() []infer.EnumValue[AuthPolicyDecision] {
	return []infer.EnumValue[AuthPolicyDecision]{
		{Value: AuthPolicyDecisionDeny, Description: "A deny rule for Oidc Issuer Policy."},
		{Value: AuthPolicyDecisionAllow, Description: "An allow rule for Oidc Issuer Policy."},
	}
}

// AuthPolicyTokenType is the token type for a policy definition.
type AuthPolicyTokenType string

const (
	AuthPolicyTokenTypePersonal         AuthPolicyTokenType = "personal"
	AuthPolicyTokenTypeTeam             AuthPolicyTokenType = gcTeam
	AuthPolicyTokenTypeOrganization     AuthPolicyTokenType = gcOrganization
	AuthPolicyTokenTypeDeploymentRunner AuthPolicyTokenType = "runner"
)

func (AuthPolicyTokenType) Values() []infer.EnumValue[AuthPolicyTokenType] {
	return []infer.EnumValue[AuthPolicyTokenType]{
		{Value: AuthPolicyTokenTypePersonal, Description: "Personal Pulumi token. Requires userLogin field to be filled."},
		{Value: AuthPolicyTokenTypeTeam, Description: "Team Pulumi token. Requires teamName field to be filled."},
		{
			Value:       AuthPolicyTokenTypeOrganization,
			Description: "Organization Pulumi token. Requires authorizedPermissions field to be filled.",
		},
		{
			Value:       AuthPolicyTokenTypeDeploymentRunner,
			Description: "Deployment Runner Pulumi token. Requires runnerID field to be filled.",
		},
	}
}

// AuthPolicyPermissionLevel is the permission level for an organization token.
type AuthPolicyPermissionLevel string

const (
	AuthPolicyPermissionLevelStandard AuthPolicyPermissionLevel = "standard"
	AuthPolicyPermissionLevelAdmin    AuthPolicyPermissionLevel = gcAdmin
)

func (AuthPolicyPermissionLevel) Values() []infer.EnumValue[AuthPolicyPermissionLevel] {
	return []infer.EnumValue[AuthPolicyPermissionLevel]{
		{Value: AuthPolicyPermissionLevelStandard, Description: "Standard level of permissions."},
		{Value: AuthPolicyPermissionLevelAdmin, Description: "Admin level of permissions."},
	}
}

// AuthPolicyDefinition is a single auth policy entry attached to an OIDC issuer.
type AuthPolicyDefinition struct {
	Decision              AuthPolicyDecision          `pulumi:"decision"`
	TokenType             AuthPolicyTokenType         `pulumi:"tokenType"`
	TeamName              *string                     `pulumi:"teamName,optional"`
	UserLogin             *string                     `pulumi:"userLogin,optional"`
	RunnerID              *string                     `pulumi:"runnerID,optional"`
	AuthorizedPermissions []AuthPolicyPermissionLevel `pulumi:"authorizedPermissions,optional"`
	Rules                 map[string]string           `pulumi:"rules"`
}

func (p *AuthPolicyDefinition) Annotate(a infer.Annotator) {
	a.Describe(&p.Decision, "The rule type of this policy definition.")
	a.Describe(&p.TokenType, "The token type for this policy definition.")
	a.Describe(&p.TeamName, "The team name for team tokens.")
	a.Describe(&p.UserLogin, "The user login for personal tokens.")
	a.Describe(&p.RunnerID, "The runner ID for deployment runner tokens.")
	a.Describe(&p.AuthorizedPermissions, "The permission level for organization tokens.")
	a.Describe(&p.Rules, "OIDC rules to set for this policy.")
}

type OidcIssuerInput struct {
	Organization         string                 `pulumi:"organization" provider:"replaceOnChanges"`
	Name                 string                 `pulumi:"name"`
	URL                  string                 `pulumi:"url"          provider:"replaceOnChanges"`
	MaxExpirationSeconds *int64                 `pulumi:"maxExpirationSeconds,optional"`
	Thumbprints          []string               `pulumi:"thumbprints,optional"`
	Policies             []AuthPolicyDefinition `pulumi:"policies,optional"`
}

const (
	thumbprintsDescription = "The thumbprints of issuer's TLS certificates. By default, Pulumi will store the " +
		"thumbprint of the certificate used to serve the OpenID configuration. If the provider uses multiple " +
		"certificates to serve content, it is required to manually configure these."
	policiesDescription = "The auth policies for this Oidc Issuer."
)

func (i *OidcIssuerInput) Annotate(a infer.Annotator) {
	a.Describe(&i.Organization, "Organization name.")
	a.Describe(&i.Name, "Issuer name.")
	a.Describe(&i.URL, "The OIDC issuer URL.")
	a.Describe(
		&i.MaxExpirationSeconds,
		"The maximum duration of the Pulumi access token working after an exchange, specified in seconds.",
	)
	a.Describe(&i.Thumbprints, thumbprintsDescription)
	a.Describe(&i.Policies, policiesDescription)
}

type OidcIssuerState struct {
	OidcIssuerInput
	// thumbprints and policies are optional to supply but always present in
	// the API response, so they are required outputs. Shadowing the embedded
	// optional input fields with required ones keeps the inputs permissive
	// while restoring the non-optional output types.
	Thumbprints []string               `pulumi:"thumbprints"`
	Policies    []AuthPolicyDefinition `pulumi:"policies"`
}

// Annotate describes the shadowing output fields. It overrides the Annotate
// method promoted from the embedded OidcIssuerInput, which would otherwise run
// against OidcIssuerState and reference the shadowed embedded fields that are
// no longer visible, panicking the annotator.
func (s *OidcIssuerState) Annotate(a infer.Annotator) {
	a.Describe(&s.Thumbprints, thumbprintsDescription)
	a.Describe(&s.Policies, policiesDescription)
}

// newOidcIssuerState builds the state from a set of inputs, populating the
// required output fields that shadow their optional input counterparts.
func newOidcIssuerState(input OidcIssuerInput) OidcIssuerState {
	return OidcIssuerState{
		OidcIssuerInput: input,
		Thumbprints:     input.Thumbprints,
		Policies:        input.Policies,
	}
}

func (*OidcIssuer) Create(
	ctx context.Context,
	req infer.CreateRequest[OidcIssuerInput],
) (infer.CreateResponse[OidcIssuerState], error) {
	if req.DryRun {
		return infer.CreateResponse[OidcIssuerState]{
			Output: newOidcIssuerState(req.Inputs),
		}, nil
	}

	client := config.GetClient(ctx)
	registerResponse, err := client.RegisterOidcIssuer(ctx, req.Inputs.Organization, req.Inputs.toCreateRequest())
	if err != nil {
		return infer.CreateResponse[OidcIssuerState]{}, fmt.Errorf(
			"error creating oidc issuer %q: %w", req.Inputs.Name, err,
		)
	}

	authPolicy, err := client.GetAuthPolicies(ctx, req.Inputs.Organization, registerResponse.ID)
	if err != nil {
		return infer.CreateResponse[OidcIssuerState]{}, fmt.Errorf(
			"error retrieving auth policies for oidc issuer %q: %w", registerResponse.ID, err,
		)
	}

	if len(req.Inputs.Policies) > 0 {
		request := policiesToAPIRequest(req.Inputs.Policies)
		authPolicy, err = client.UpdateAuthPolicies(ctx, req.Inputs.Organization, authPolicy.ID, request)
		if err != nil {
			// Best-effort cleanup to avoid leaving the issuer in limbo when
			// the user-supplied policies are rejected by the API.
			_ = client.DeleteOidcIssuer(ctx, req.Inputs.Organization, registerResponse.ID)
			return infer.CreateResponse[OidcIssuerState]{}, fmt.Errorf(
				"error updating auth policies for oidc issuer %q: %w", registerResponse.ID, err,
			)
		}
	}

	state := oidcIssuerStateFromAPI(req.Inputs.Organization, *registerResponse, authPolicy)
	return infer.CreateResponse[OidcIssuerState]{
		ID:     oidcIssuerID(req.Inputs.Organization, registerResponse.ID),
		Output: state,
	}, nil
}

func (*OidcIssuer) Update(
	ctx context.Context,
	req infer.UpdateRequest[OidcIssuerInput, OidcIssuerState],
) (infer.UpdateResponse[OidcIssuerState], error) {
	if req.DryRun {
		return infer.UpdateResponse[OidcIssuerState]{
			Output: newOidcIssuerState(req.Inputs),
		}, nil
	}

	_, issuerID, err := splitOidcIssuerID(req.ID)
	if err != nil {
		return infer.UpdateResponse[OidcIssuerState]{}, err
	}

	client := config.GetClient(ctx)
	updateResponse, err := client.UpdateOidcIssuer(
		ctx, req.Inputs.Organization, issuerID, req.Inputs.toUpdateRequest(),
	)
	if err != nil {
		return infer.UpdateResponse[OidcIssuerState]{}, fmt.Errorf(
			"error updating oidc issuer %q: %w", req.Inputs.Name, err,
		)
	}

	authPolicy, err := client.GetAuthPolicies(ctx, req.Inputs.Organization, issuerID)
	if err != nil {
		return infer.UpdateResponse[OidcIssuerState]{}, fmt.Errorf(
			"error retrieving auth policies for oidc issuer %q: %w", issuerID, err,
		)
	}

	if len(req.Inputs.Policies) > 0 {
		request := policiesToAPIRequest(req.Inputs.Policies)
		authPolicy, err = client.UpdateAuthPolicies(ctx, req.Inputs.Organization, authPolicy.ID, request)
		if err != nil {
			return infer.UpdateResponse[OidcIssuerState]{}, fmt.Errorf(
				"error updating auth policies for oidc issuer %q: %w", issuerID, err,
			)
		}
	}

	state := oidcIssuerStateFromAPI(req.Inputs.Organization, *updateResponse, authPolicy)
	return infer.UpdateResponse[OidcIssuerState]{Output: state}, nil
}

func (*OidcIssuer) Delete(
	ctx context.Context,
	req infer.DeleteRequest[OidcIssuerState],
) (infer.DeleteResponse, error) {
	orgName, issuerID, err := splitOidcIssuerID(req.ID)
	if err != nil {
		return infer.DeleteResponse{}, err
	}
	return infer.DeleteResponse{}, config.GetClient(ctx).DeleteOidcIssuer(ctx, orgName, issuerID)
}

func (*OidcIssuer) Read(
	ctx context.Context,
	req infer.ReadRequest[OidcIssuerInput, OidcIssuerState],
) (infer.ReadResponse[OidcIssuerInput, OidcIssuerState], error) {
	orgName, issuerID, err := splitOidcIssuerID(req.ID)
	if err != nil {
		return infer.ReadResponse[OidcIssuerInput, OidcIssuerState]{}, err
	}

	client := config.GetClient(ctx)
	readResponse, err := client.GetOidcIssuer(ctx, orgName, issuerID)
	if err != nil {
		return infer.ReadResponse[OidcIssuerInput, OidcIssuerState]{}, fmt.Errorf(
			"failed to read oidc issuer (%q): %w", req.ID, err,
		)
	}
	if readResponse == nil {
		return infer.ReadResponse[OidcIssuerInput, OidcIssuerState]{}, nil
	}

	authPolicy, err := client.GetAuthPolicies(ctx, orgName, issuerID)
	if err != nil {
		return infer.ReadResponse[OidcIssuerInput, OidcIssuerState]{}, fmt.Errorf(
			"error retrieving auth policies for oidc issuer %q: %w", issuerID, err,
		)
	}

	state := oidcIssuerStateFromAPI(orgName, *readResponse, authPolicy)
	return infer.ReadResponse[OidcIssuerInput, OidcIssuerState]{
		ID:     req.ID,
		Inputs: state.OidcIssuerInput,
		State:  state,
	}, nil
}

// oidcIssuerStateFromAPI builds an OidcIssuerState from a successful API
// response. The ordering of policies returned by the API is preserved as-is.
func oidcIssuerStateFromAPI(
	organization string,
	issuer pulumiapi.OidcIssuerRegistrationResponse,
	authPolicy *pulumiapi.AuthPolicy,
) OidcIssuerState {
	input := OidcIssuerInput{
		Organization:         organization,
		Name:                 issuer.Name,
		URL:                  issuer.URL,
		MaxExpirationSeconds: issuer.MaxExpiration,
		Thumbprints:          issuer.Thumbprints,
	}
	if authPolicy != nil {
		input.Policies = apiPoliciesToInputs(authPolicy.Definition)
	}
	return newOidcIssuerState(input)
}

func (i *OidcIssuerInput) toCreateRequest() pulumiapi.OidcIssuerRegistrationRequest {
	return pulumiapi.OidcIssuerRegistrationRequest{
		Name:          i.Name,
		URL:           i.URL,
		Thumbprints:   i.Thumbprints,
		MaxExpiration: i.MaxExpirationSeconds,
	}
}

func (i *OidcIssuerInput) toUpdateRequest() pulumiapi.OidcIssuerUpdateRequest {
	return pulumiapi.OidcIssuerUpdateRequest{
		Name:          &i.Name,
		Thumbprints:   &i.Thumbprints,
		MaxExpiration: i.MaxExpirationSeconds,
	}
}

func policiesToAPIRequest(policies []AuthPolicyDefinition) pulumiapi.AuthPolicyUpdateRequest {
	apiPolicies := make([]pulumiapi.AuthPolicyDefinition, 0, len(policies))
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
	return pulumiapi.AuthPolicyUpdateRequest{Definition: apiPolicies}
}

func apiPoliciesToInputs(policies []*pulumiapi.AuthPolicyDefinition) []AuthPolicyDefinition {
	if len(policies) == 0 {
		return nil
	}
	result := make([]AuthPolicyDefinition, 0, len(policies))
	for _, policy := range policies {
		if policy == nil {
			continue
		}
		result = append(result, AuthPolicyDefinition{
			Decision:              AuthPolicyDecision(policy.Decision),
			TokenType:             AuthPolicyTokenType(policy.TokenType),
			TeamName:              policy.TeamName,
			UserLogin:             policy.UserLogin,
			RunnerID:              policy.RunnerID,
			AuthorizedPermissions: permissionsToInput(policy.AuthorizedPermissions),
			Rules:                 policy.Rules,
		})
	}
	return result
}

func permissionsToAPI(permissions []AuthPolicyPermissionLevel) []string {
	out := make([]string, 0, len(permissions))
	for _, permission := range permissions {
		out = append(out, string(permission))
	}
	return out
}

func permissionsToInput(permissions []string) []AuthPolicyPermissionLevel {
	if len(permissions) == 0 {
		return nil
	}
	out := make([]AuthPolicyPermissionLevel, 0, len(permissions))
	for _, permission := range permissions {
		out = append(out, AuthPolicyPermissionLevel(permission))
	}
	return out
}

func oidcIssuerID(organization, issuerID string) string {
	return fmt.Sprintf("%s/%s", organization, issuerID)
}

func splitOidcIssuerID(id string) (organization, issuerID string, err error) {
	splitID := strings.Split(id, "/")
	if len(splitID) != 2 {
		return "", "", fmt.Errorf("error splitting resource id %q", id)
	}
	return splitID[0], splitID[1], nil
}
