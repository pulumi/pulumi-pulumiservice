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

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/config"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
)

type EnvironmentPermission string

const (
	EnvironmentPermissionNone  EnvironmentPermission = "none"
	EnvironmentPermissionRead  EnvironmentPermission = "read"
	EnvironmentPermissionOpen  EnvironmentPermission = "open"
	EnvironmentPermissionWrite EnvironmentPermission = "write"
	EnvironmentPermissionAdmin EnvironmentPermission = gcAdmin
)

func (EnvironmentPermission) Values() []infer.EnumValue[EnvironmentPermission] {
	return []infer.EnumValue[EnvironmentPermission]{
		{Value: EnvironmentPermissionNone, Description: "No permissions."},
		{Value: EnvironmentPermissionRead, Description: "Permission to read environment definition only."},
		{Value: EnvironmentPermissionOpen, Description: "Permission to open and read the environment."},
		{Value: EnvironmentPermissionWrite, Description: "Permission to open, read and update the environment."},
		{Value: EnvironmentPermissionAdmin, Description: "Permission for all operations on the environment."},
	}
}

type TeamEnvironmentPermission struct{}

type (
	tepIn  = TeamEnvironmentPermissionInput
	tepOut = TeamEnvironmentPermissionState
)

var (
	_ infer.CustomCheck[tepIn]          = &TeamEnvironmentPermission{}
	_ infer.CustomDiff[tepIn, tepOut]   = &TeamEnvironmentPermission{}
	_ infer.CustomCreate[tepIn, tepOut] = &TeamEnvironmentPermission{}
	_ infer.CustomDelete[tepOut]        = &TeamEnvironmentPermission{}
	_ infer.CustomRead[tepIn, tepOut]   = &TeamEnvironmentPermission{}
)

func (*TeamEnvironmentPermission) Annotate(a infer.Annotator) {
	a.Describe(&TeamEnvironmentPermission{}, "A permission for a team to use an environment.")
	a.SetToken("index", "TeamEnvironmentPermission")
}

type TeamEnvironmentPermissionInput struct {
	Organization    string                `pulumi:"organization"              provider:"replaceOnChanges"`
	Team            string                `pulumi:"team"                      provider:"replaceOnChanges"`
	Project         string                `pulumi:"project,optional"          provider:"replaceOnChanges"`
	Environment     string                `pulumi:"environment"               provider:"replaceOnChanges"`
	Permission      EnvironmentPermission `pulumi:"permission"                provider:"replaceOnChanges"`
	MaxOpenDuration *string               `pulumi:"maxOpenDuration,optional"  provider:"replaceOnChanges"`
}

func (i *TeamEnvironmentPermissionInput) Annotate(a infer.Annotator) {
	a.Describe(&i.Organization, "Organization name.")
	a.Describe(&i.Team, "Team name.")
	a.Describe(&i.Project, "Project name.")
	a.SetDefault(&i.Project, "default")
	a.Describe(&i.Environment, "Environment name.")
	a.Describe(&i.Permission, "Which permission level to grant to the specified team.")
	a.Describe(
		&i.MaxOpenDuration,
		"The maximum duration for which members of this team may open the environment.",
	)
}

type TeamEnvironmentPermissionState struct {
	TeamEnvironmentPermissionInput
}

func (*TeamEnvironmentPermission) Check(
	ctx context.Context, req infer.CheckRequest,
) (infer.CheckResponse[TeamEnvironmentPermissionInput], error) {
	// Strip an explicit empty-string maxOpenDuration before decoding.
	// Provider versions 0.29.3–0.36.0 wrote `""` into state when the user
	// did not set the field; treating that as "unset" keeps re-applies
	// against that state from forcing a spurious replacement.
	if v, ok := req.NewInputs.GetOk("maxOpenDuration"); ok && v.IsString() && v.AsString() == "" {
		req.NewInputs = req.NewInputs.Delete("maxOpenDuration")
	}

	i, failures, err := infer.DefaultCheck[TeamEnvironmentPermissionInput](ctx, req.NewInputs)
	if err != nil {
		return infer.CheckResponse[TeamEnvironmentPermissionInput]{}, err
	}
	if i.MaxOpenDuration != nil {
		d, perr := time.ParseDuration(*i.MaxOpenDuration)
		if perr != nil {
			failures = append(failures, p.CheckFailure{
				Property: "maxOpenDuration",
				Reason:   fmt.Sprintf("malformed duration: %v", perr),
			})
		} else if normalized := d.String(); normalized != *i.MaxOpenDuration {
			i.MaxOpenDuration = &normalized
		}
	}
	return infer.CheckResponse[TeamEnvironmentPermissionInput]{Inputs: i, Failures: failures}, nil
}

func (*TeamEnvironmentPermission) Diff(
	_ context.Context,
	req infer.DiffRequest[TeamEnvironmentPermissionInput, TeamEnvironmentPermissionState],
) (infer.DiffResponse, error) {
	diff := map[string]p.PropertyDiff{}
	add := func(key string) { diff[key] = p.PropertyDiff{Kind: p.UpdateReplace, InputDiff: true} }

	if req.State.Organization != req.Inputs.Organization {
		add(gcOrganization)
	}
	if req.State.Team != req.Inputs.Team {
		add(gcTeam)
	}
	if req.State.Project != req.Inputs.Project {
		add(gcProject)
	}
	if req.State.Environment != req.Inputs.Environment {
		add(gcEnvironment)
	}
	if req.State.Permission != req.Inputs.Permission {
		add("permission")
	}
	if normalizedDuration(req.State.MaxOpenDuration) != normalizedDuration(req.Inputs.MaxOpenDuration) {
		add("maxOpenDuration")
	}

	return infer.DiffResponse{
		HasChanges:          len(diff) > 0,
		DetailedDiff:        diff,
		DeleteBeforeReplace: true,
	}, nil
}

// normalizedDuration treats nil and an empty string as equivalent so a
// pre-#752 empty-string in state does not diff against an absent field.
func normalizedDuration(d *string) string {
	if d == nil {
		return ""
	}
	return *d
}

func (*TeamEnvironmentPermission) Create(
	ctx context.Context,
	req infer.CreateRequest[TeamEnvironmentPermissionInput],
) (infer.CreateResponse[TeamEnvironmentPermissionState], error) {
	if req.DryRun {
		return infer.CreateResponse[TeamEnvironmentPermissionState]{
			Output: TeamEnvironmentPermissionState{TeamEnvironmentPermissionInput: req.Inputs},
		}, nil
	}
	apiReq, err := req.Inputs.toCreateRequest()
	if err != nil {
		return infer.CreateResponse[TeamEnvironmentPermissionState]{}, err
	}
	if err := config.GetClient(ctx).AddEnvironmentSettings(ctx, apiReq); err != nil {
		return infer.CreateResponse[TeamEnvironmentPermissionState]{},
			fmt.Errorf("error granting team environment permission: %w", err)
	}
	id := teamEnvironmentPermissionID{
		Organization: req.Inputs.Organization,
		Team:         req.Inputs.Team,
		Project:      req.Inputs.Project,
		Environment:  req.Inputs.Environment,
	}
	return infer.CreateResponse[TeamEnvironmentPermissionState]{
		ID:     id.String(),
		Output: TeamEnvironmentPermissionState{TeamEnvironmentPermissionInput: req.Inputs},
	}, nil
}

func (*TeamEnvironmentPermission) Delete(
	ctx context.Context,
	req infer.DeleteRequest[TeamEnvironmentPermissionState],
) (infer.DeleteResponse, error) {
	apiReq := pulumiapi.TeamEnvironmentSettingsRequest{
		Organization: req.State.Organization,
		Team:         req.State.Team,
		Project:      req.State.Project,
		Environment:  req.State.Environment,
	}
	return infer.DeleteResponse{}, config.GetClient(ctx).RemoveEnvironmentSettings(ctx, apiReq)
}

func (*TeamEnvironmentPermission) Read(
	ctx context.Context,
	req infer.ReadRequest[TeamEnvironmentPermissionInput, TeamEnvironmentPermissionState],
) (infer.ReadResponse[TeamEnvironmentPermissionInput, TeamEnvironmentPermissionState], error) {
	permID, err := splitTeamEnvironmentPermissionID(req.ID)
	if err != nil {
		return infer.ReadResponse[TeamEnvironmentPermissionInput, TeamEnvironmentPermissionState]{}, err
	}
	apiReq := pulumiapi.TeamEnvironmentSettingsRequest{
		Organization: permID.Organization,
		Team:         permID.Team,
		Project:      permID.Project,
		Environment:  permID.Environment,
	}
	permission, maxOpenDuration, err := config.GetClient(ctx).GetTeamEnvironmentSettings(ctx, apiReq)
	if err != nil {
		return infer.ReadResponse[TeamEnvironmentPermissionInput, TeamEnvironmentPermissionState]{},
			fmt.Errorf("failed to get team environment permission: %w", err)
	}
	if permission == nil {
		return infer.ReadResponse[TeamEnvironmentPermissionInput, TeamEnvironmentPermissionState]{}, nil
	}
	inputs := TeamEnvironmentPermissionInput{
		Organization: permID.Organization,
		Team:         permID.Team,
		Project:      permID.Project,
		Environment:  permID.Environment,
		Permission:   EnvironmentPermission(*permission),
	}
	if maxOpenDuration != nil {
		s := time.Duration(*maxOpenDuration).String()
		inputs.MaxOpenDuration = &s
	}
	return infer.ReadResponse[TeamEnvironmentPermissionInput, TeamEnvironmentPermissionState]{
		ID:     req.ID,
		Inputs: inputs,
		State:  TeamEnvironmentPermissionState{TeamEnvironmentPermissionInput: inputs},
	}, nil
}

func (i TeamEnvironmentPermissionInput) toCreateRequest() (
	pulumiapi.CreateTeamEnvironmentSettingsRequest, error,
) {
	apiReq := pulumiapi.CreateTeamEnvironmentSettingsRequest{
		TeamEnvironmentSettingsRequest: pulumiapi.TeamEnvironmentSettingsRequest{
			Organization: i.Organization,
			Team:         i.Team,
			Project:      i.Project,
			Environment:  i.Environment,
		},
		Permission: string(i.Permission),
	}
	if i.MaxOpenDuration != nil {
		d, err := time.ParseDuration(*i.MaxOpenDuration)
		if err != nil {
			return pulumiapi.CreateTeamEnvironmentSettingsRequest{},
				fmt.Errorf("invalid maxOpenDuration %q: %w", *i.MaxOpenDuration, err)
		}
		mod := pulumiapi.Duration(d)
		apiReq.MaxOpenDuration = &mod
	}
	return apiReq, nil
}

type teamEnvironmentPermissionID struct {
	Organization string
	Team         string
	Project      string
	Environment  string
}

func (s teamEnvironmentPermissionID) String() string {
	return fmt.Sprintf("%s/%s/%s+%s", s.Organization, s.Team, s.Project, s.Environment)
}

func splitTeamEnvironmentPermissionID(id string) (teamEnvironmentPermissionID, error) {
	split := strings.Split(id, "/")
	if len(split) != 3 {
		return teamEnvironmentPermissionID{}, fmt.Errorf("invalid id %q, expected 3 parts", id)
	}
	splitProjectEnv := strings.Split(split[2], "+")
	switch len(splitProjectEnv) {
	case 1:
		return teamEnvironmentPermissionID{
			Organization: split[0],
			Team:         split[1],
			Project:      "default",
			Environment:  splitProjectEnv[0],
		}, nil
	case 2:
		return teamEnvironmentPermissionID{
			Organization: split[0],
			Team:         split[1],
			Project:      splitProjectEnv[0],
			Environment:  splitProjectEnv[1],
		}, nil
	}
	return teamEnvironmentPermissionID{}, fmt.Errorf(
		"invalid id %q, expected environment name or project/environment in last part", id,
	)
}
