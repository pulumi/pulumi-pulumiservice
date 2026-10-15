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
	"time"

	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/pulumi/pulumi/sdk/v3/go/property"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/config"
)

// The examples use the docs `{{% examples %}}` shortcodes, which SDK
// generation and the registry render per language.
const orgAccessTokenExamples = "{{% examples %}}\n## Example Usage\n\n" +
	"{{% example %}}\n### Token with an expiry and a custom role\n\n" +
	"```typescript\n" +
	`import * as pulumiservice from "@pulumi/pulumiservice";

const ciDeployerRole = pulumiservice.getOrganizationRoleOutput({
    organizationName: "my-org",
    name: "CI Deployer",
});

const ciToken = new pulumiservice.OrgAccessToken("ciToken", {
    organizationName: "my-org",
    name: "ci-token",
    description: "Token for CI pipelines",
    expires: "2027-06-30T17:00:00Z",
    roleId: ciDeployerRole.roleId,
});

export const ciTokenValue = ciToken.value;
` + "```\n" +
	"```python\n" +
	`import pulumi
import pulumi_pulumiservice as pulumiservice

ci_deployer_role = pulumiservice.get_organization_role_output(
    organization_name="my-org",
    name="CI Deployer",
)

ci_token = pulumiservice.OrgAccessToken("ciToken",
    organization_name="my-org",
    name="ci-token",
    description="Token for CI pipelines",
    expires="2027-06-30T17:00:00Z",
    role_id=ci_deployer_role.role_id,
)

pulumi.export("ciTokenValue", ci_token.value)
` + "```\n" +
	"```go\n" +
	`package main

import (
	"github.com/pulumi/pulumi-pulumiservice/sdk/go/pulumiservice"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		ciDeployerRole := pulumiservice.LookupOrganizationRoleOutput(ctx, pulumiservice.LookupOrganizationRoleOutputArgs{
			OrganizationName: pulumi.String("my-org"),
			Name:             pulumi.String("CI Deployer"),
		})

		ciToken, err := pulumiservice.NewOrgAccessToken(ctx, "ciToken", &pulumiservice.OrgAccessTokenArgs{
			OrganizationName: pulumi.String("my-org"),
			Name:             pulumi.String("ci-token"),
			Description:      pulumi.String("Token for CI pipelines"),
			Expires:          pulumi.String("2027-06-30T17:00:00Z"),
			RoleId:           ciDeployerRole.RoleId(),
		})
		if err != nil {
			return err
		}

		ctx.Export("ciTokenValue", ciToken.Value)
		return nil
	})
}
` + "```\n" +
	"```csharp\n" +
	`using System.Collections.Generic;
using Pulumi;
using PulumiService = Pulumi.PulumiService;

return await Deployment.RunAsync(() =>
{
    var ciDeployerRole = PulumiService.GetOrganizationRole.Invoke(new()
    {
        OrganizationName = "my-org",
        Name = "CI Deployer",
    });

    var ciToken = new PulumiService.OrgAccessToken("ciToken", new()
    {
        OrganizationName = "my-org",
        Name = "ci-token",
        Description = "Token for CI pipelines",
        Expires = "2027-06-30T17:00:00Z",
        RoleId = ciDeployerRole.Apply(role => role.RoleId),
    });

    return new Dictionary<string, object?>
    {
        ["ciTokenValue"] = ciToken.Value,
    };
});
` + "```\n" +
	"```java\n" +
	`package generated_program;

import com.pulumi.Context;
import com.pulumi.Pulumi;
import com.pulumi.pulumiservice.OrgAccessToken;
import com.pulumi.pulumiservice.OrgAccessTokenArgs;
import com.pulumi.pulumiservice.PulumiserviceFunctions;
import com.pulumi.pulumiservice.inputs.GetOrganizationRoleArgs;

public class App {
    public static void main(String[] args) {
        Pulumi.run(App::stack);
    }

    public static void stack(Context ctx) {
        final var ciDeployerRole = PulumiserviceFunctions.getOrganizationRole(GetOrganizationRoleArgs.builder()
            .organizationName("my-org")
            .name("CI Deployer")
            .build());

        var ciToken = new OrgAccessToken("ciToken", OrgAccessTokenArgs.builder()
            .organizationName("my-org")
            .name("ci-token")
            .description("Token for CI pipelines")
            .expires("2027-06-30T17:00:00Z")
            .roleId(ciDeployerRole.applyValue(role -> role.roleId()))
            .build());

        ctx.export("ciTokenValue", ciToken.value());
    }
}
` + "```\n" +
	"```yaml\n" +
	`variables:
  ciDeployerRole:
    fn::invoke:
      function: pulumiservice:getOrganizationRole
      arguments:
        organizationName: my-org
        name: CI Deployer

resources:
  ciToken:
    type: pulumiservice:OrgAccessToken
    properties:
      organizationName: my-org
      name: ci-token
      description: Token for CI pipelines
      expires: "2027-06-30T17:00:00Z"
      roleId: ${ciDeployerRole.roleId}

outputs:
  ciTokenValue: ${ciToken.value}
` + "```\n{{% /example %}}\n{{% /examples %}}"

type OrgAccessToken struct{}

var (
	_ infer.CustomCheck[OrgAccessTokenInput]                       = &OrgAccessToken{}
	_ infer.CustomCreate[OrgAccessTokenInput, OrgAccessTokenState] = &OrgAccessToken{}
	_ infer.CustomDelete[OrgAccessTokenState]                      = &OrgAccessToken{}
	_ infer.CustomRead[OrgAccessTokenInput, OrgAccessTokenState]   = &OrgAccessToken{}
	_ infer.CustomStateMigrations[OrgAccessTokenState]             = &OrgAccessToken{}
)

func (*OrgAccessToken) Annotate(a infer.Annotator) {
	a.Describe(
		&OrgAccessToken{},
		"Creates an organization access token in Pulumi Cloud. Organization tokens aren't tied to a user, "+
			"so they suit CI/CD and other automation.\n\n"+
			"The provider must authenticate as a user, for example with a personal access token, to create "+
			"or delete this resource. Pulumi Cloud rejects organization and team access tokens here, even "+
			"Admin ones, with `403 Machine tokens are not allowed to perform this operation`, so that one "+
			"machine token can't mint another. For automation, use the personal access token of a dedicated "+
			"service-account user, or use `TeamAccessToken`, which has no such restriction.\n\n"+
			orgAccessTokenExamples,
	)
}

type OrgAccessTokenInput struct {
	Name             string  `pulumi:"name"                 provider:"replaceOnChanges"`
	OrganizationName string  `pulumi:"organizationName"     provider:"replaceOnChanges"`
	Description      *string `pulumi:"description,optional" provider:"replaceOnChanges"`
	Admin            *bool   `pulumi:"admin,optional"       provider:"replaceOnChanges"`
	Expires          *string `pulumi:"expires,optional"     provider:"replaceOnChanges"`
	RoleID           *string `pulumi:"roleId,optional"      provider:"replaceOnChanges"`
}

func (i *OrgAccessTokenInput) Annotate(a infer.Annotator) {
	a.Describe(&i.Name, "The name for the token.")
	a.Describe(&i.OrganizationName, "The organization's name.")
	a.Describe(&i.Description, "A description of the token.")
	a.Describe(
		&i.Admin,
		"Whether the token has the Admin role. Defaults to `false`. Can't be `true` when `roleId` is set.",
	)
	a.SetDefault(&i.Admin, false)
	a.Describe(
		&i.Expires,
		"When the token expires, as an RFC 3339 date-time such as `2027-06-30T17:00:00Z`, or a date such as "+
			"`2027-06-30`, which means midnight UTC at the start of that day. Must be in the future and no more "+
			"than two years away. Omit it for a token that never expires.",
	)
	a.Describe(
		&i.RoleID,
		"The ID of the role to assign to the token. Use `getOrganizationRole` to look up a role's ID by its "+
			"name. Can't be combined with `admin: true`. When neither is set, the token gets the Member role.",
	)
}

type OrgAccessTokenState struct {
	OrgAccessTokenInput
	Value string `pulumi:"value" provider:"secret"`
}

func (s *OrgAccessTokenState) Annotate(a infer.Annotator) {
	a.Describe(&s.Value, "The token's value.")
}

func (*OrgAccessToken) Create(
	ctx context.Context,
	req infer.CreateRequest[OrgAccessTokenInput],
) (infer.CreateResponse[OrgAccessTokenState], error) {
	if req.DryRun {
		return infer.CreateResponse[OrgAccessTokenState]{
			Output: OrgAccessTokenState{OrgAccessTokenInput: req.Inputs},
		}, nil
	}
	desc := ""
	if req.Inputs.Description != nil {
		desc = *req.Inputs.Description
	}
	admin := false
	if req.Inputs.Admin != nil {
		admin = *req.Inputs.Admin
	}
	var expires int64
	if req.Inputs.Expires != nil {
		t, err := parseTokenExpiry(*req.Inputs.Expires)
		if err != nil {
			return infer.CreateResponse[OrgAccessTokenState]{}, err
		}
		expires = t.Unix()
	}
	roleID := ""
	if req.Inputs.RoleID != nil {
		roleID = *req.Inputs.RoleID
	}
	token, err := config.GetClient(ctx).CreateOrgAccessToken(
		ctx, req.Inputs.Name, req.Inputs.OrganizationName, desc, admin, expires, roleID,
	)
	if err != nil {
		return infer.CreateResponse[OrgAccessTokenState]{}, fmt.Errorf(
			"error creating org access token %q: %w", req.Inputs.Name, err,
		)
	}
	return infer.CreateResponse[OrgAccessTokenState]{
		ID: orgAccessTokenID(req.Inputs.OrganizationName, req.Inputs.Name, token.ID),
		Output: OrgAccessTokenState{
			OrgAccessTokenInput: req.Inputs,
			Value:               token.TokenValue,
		},
	}, nil
}

func (*OrgAccessToken) Delete(
	ctx context.Context,
	req infer.DeleteRequest[OrgAccessTokenState],
) (infer.DeleteResponse, error) {
	orgName, _, tokenID, err := splitOrgAccessTokenID(req.ID)
	if err != nil {
		return infer.DeleteResponse{}, err
	}
	return infer.DeleteResponse{}, config.GetClient(ctx).DeleteOrgAccessToken(ctx, tokenID, orgName)
}

func (*OrgAccessToken) Read(
	ctx context.Context,
	req infer.ReadRequest[OrgAccessTokenInput, OrgAccessTokenState],
) (infer.ReadResponse[OrgAccessTokenInput, OrgAccessTokenState], error) {
	orgName, _, tokenID, err := splitOrgAccessTokenID(req.ID)
	if err != nil {
		return infer.ReadResponse[OrgAccessTokenInput, OrgAccessTokenState]{}, err
	}

	token, err := config.GetClient(ctx).GetOrgAccessToken(ctx, tokenID, orgName)
	if err != nil {
		return infer.ReadResponse[OrgAccessTokenInput, OrgAccessTokenState]{}, err
	}
	if token == nil {
		return infer.ReadResponse[OrgAccessTokenInput, OrgAccessTokenState]{}, nil
	}

	admin := token.Admin
	inputs := OrgAccessTokenInput{
		Name:             token.Name,
		OrganizationName: orgName,
		Description:      stringPtrIfNonEmpty(token.Description),
		Admin:            &admin,
		Expires:          readTokenExpiry(token.Expires, req.State.Expires),
		RoleID:           stringPtrIfNonEmpty(token.RoleID),
	}
	return infer.ReadResponse[OrgAccessTokenInput, OrgAccessTokenState]{
		ID:     req.ID,
		Inputs: inputs,
		State: OrgAccessTokenState{
			OrgAccessTokenInput: inputs,
			// Token values aren't retrievable from the API after creation; carry
			// the existing secret from state so refresh does not erase it.
			Value: req.State.Value,
		},
	}, nil
}

func (*OrgAccessToken) Check(
	ctx context.Context,
	req infer.CheckRequest,
) (infer.CheckResponse[OrgAccessTokenInput], error) {
	in, failures, err := infer.DefaultCheck[OrgAccessTokenInput](ctx, req.NewInputs)
	if err != nil {
		return infer.CheckResponse[OrgAccessTokenInput]{}, err
	}

	if in.Expires != nil && !isUnknownInput(req.NewInputs, gcExpires) {
		// Only range-check a new or changed value. An expiry that was valid
		// when the token was created will eventually be in the past, and that
		// must not make every later preview fail.
		var old *string
		if v, ok := req.OldInputs.GetOk(gcExpires); ok && v.IsString() {
			s := v.AsString()
			old = &s
		}
		changed := old == nil || *old != *in.Expires
		if reason := checkTokenExpiry(*in.Expires, changed, tokenExpiryNow()); reason != "" {
			failures = append(failures, p.CheckFailure{Property: gcExpires, Reason: reason})
		}
	}

	if in.Admin != nil && *in.Admin && in.RoleID != nil && *in.RoleID != "" &&
		!isUnknownInput(req.NewInputs, gcAdmin) {
		failures = append(failures, p.CheckFailure{
			Property: gcRoleIDKey,
			Reason:   "roleId can't be combined with admin: true; set one or the other",
		})
	}

	return infer.CheckResponse[OrgAccessTokenInput]{Inputs: in, Failures: failures}, nil
}

// StateMigrations strips the legacy "__inputs" outputs property that the
// pre-infer OrgAccessToken resource embedded in state. infer rejects unknown
// fields when decoding state, so without this migration a refresh against an
// existing stack errors with "Unrecognized field '__inputs'".
func (*OrgAccessToken) StateMigrations(context.Context) []infer.StateMigrationFunc[OrgAccessTokenState] {
	return []infer.StateMigrationFunc[OrgAccessTokenState]{
		infer.StateMigration(migrateOrgAccessTokenLegacyInputs),
	}
}

func migrateOrgAccessTokenLegacyInputs(
	_ context.Context, old property.Map,
) (infer.MigrationResult[OrgAccessTokenState], error) {
	if _, ok := old.GetOk(gcInputs); !ok {
		return infer.MigrationResult[OrgAccessTokenState]{}, nil
	}
	state := OrgAccessTokenState{}
	if v, ok := old.GetOk(gcName); ok && v.IsString() {
		state.Name = v.AsString()
	}
	if v, ok := old.GetOk(gcOrganizationName); ok && v.IsString() {
		state.OrganizationName = v.AsString()
	}
	if v, ok := old.GetOk(gcDescription); ok && v.IsString() {
		s := v.AsString()
		state.Description = &s
	}
	if v, ok := old.GetOk(gcAdmin); ok && v.IsBool() {
		b := v.AsBool()
		state.Admin = &b
	}
	if v, ok := old.GetOk(gcValue); ok && v.IsString() {
		state.Value = v.AsString()
	}
	return infer.MigrationResult[OrgAccessTokenState]{Result: &state}, nil
}

const (
	gcExpires   = "expires"
	gcRoleIDKey = "roleId"

	expiryDateLayout = "2006-01-02"
)

// tokenExpiryMaxYears is the furthest ahead Pulumi Cloud lets a token expire.
const tokenExpiryMaxYears = 2

// tokenExpiryNow is a variable so tests can pin the clock.
var tokenExpiryNow = time.Now

// parseTokenExpiry accepts an RFC 3339 date-time or a YYYY-MM-DD date, which
// it treats as midnight UTC.
func parseTokenExpiry(s string) (time.Time, error) {
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, nil
	}
	if t, err := time.Parse(expiryDateLayout, s); err == nil {
		return t, nil
	}
	return time.Time{}, fmt.Errorf(
		"%q is not a valid expiry; use an RFC 3339 date-time such as \"2027-06-30T17:00:00Z\" "+
			"or a date such as \"2027-06-30\"", s,
	)
}

// checkTokenExpiry returns a failure reason, or "" when s is acceptable. The
// range check only applies when checkRange is set.
func checkTokenExpiry(s string, checkRange bool, now time.Time) string {
	t, err := parseTokenExpiry(s)
	if err != nil {
		return err.Error()
	}
	if !checkRange {
		return ""
	}
	if !t.After(now) {
		return fmt.Sprintf("expires %q is in the past", s)
	}
	if t.After(now.AddDate(tokenExpiryMaxYears, 0, 0)) {
		return fmt.Sprintf("expires %q is more than %d years away, which Pulumi Cloud doesn't allow",
			s, tokenExpiryMaxYears)
	}
	return ""
}

// formatTokenExpiry renders unix seconds as a date when the time is midnight
// UTC, and as an RFC 3339 date-time otherwise.
func formatTokenExpiry(epoch int64) string {
	t := time.Unix(epoch, 0).UTC()
	if t.Hour() == 0 && t.Minute() == 0 && t.Second() == 0 {
		return t.Format(expiryDateLayout)
	}
	return t.Format(time.RFC3339)
}

// readTokenExpiry converts the API's expiry to the input value. When the prior
// value denotes the same instant, it keeps the user's spelling (for example
// "2027-06-30T12:00:00-05:00") so refresh doesn't report a change.
func readTokenExpiry(epoch int64, prior *string) *string {
	if epoch == 0 {
		return nil
	}
	if prior != nil {
		if t, err := parseTokenExpiry(*prior); err == nil && t.Unix() == epoch {
			return prior
		}
	}
	s := formatTokenExpiry(epoch)
	return &s
}

func orgAccessTokenID(org, name, tokenID string) string {
	return fmt.Sprintf("%s/%s/%s", org, name, tokenID)
}

func splitOrgAccessTokenID(id string) (string, string, string, error) {
	// format: organization/name/tokenID. Name may itself contain slashes,
	// so org is the first segment and tokenID is the last.
	s := strings.Split(id, "/")
	if len(s) < 3 {
		return "", "", "", fmt.Errorf(
			"%q is invalid, must be of the form organization/name/tokenId",
			id,
		)
	}
	org := s[0]
	tokenID := s[len(s)-1]
	name := strings.Join(s[1:len(s)-1], "/")
	return org, name, tokenID, nil
}
