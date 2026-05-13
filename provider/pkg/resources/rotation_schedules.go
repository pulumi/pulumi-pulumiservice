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

type EnvironmentRotationSchedule struct{}

type (
	rotIn  = EnvironmentRotationScheduleInput
	rotOut = EnvironmentRotationScheduleState
)

var (
	_ infer.CustomCheck[rotIn]          = &EnvironmentRotationSchedule{}
	_ infer.CustomCreate[rotIn, rotOut] = &EnvironmentRotationSchedule{}
	_ infer.CustomUpdate[rotIn, rotOut] = &EnvironmentRotationSchedule{}
	_ infer.CustomDelete[rotOut]        = &EnvironmentRotationSchedule{}
	_ infer.CustomRead[rotIn, rotOut]   = &EnvironmentRotationSchedule{}
)

func (*EnvironmentRotationSchedule) Annotate(a infer.Annotator) {
	a.Describe(&EnvironmentRotationSchedule{}, "A scheduled recurring or single time environment rotation.")
	a.SetToken("index", "EnvironmentRotationSchedule")
}

type EnvironmentRotationScheduleInput struct {
	Organization string  `pulumi:"organization" provider:"replaceOnChanges"`
	Project      string  `pulumi:"project"      provider:"replaceOnChanges"`
	Environment  string  `pulumi:"environment"  provider:"replaceOnChanges"`
	ScheduleCron *string `pulumi:"scheduleCron,optional"`
	Timestamp    *string `pulumi:"timestamp,optional"    provider:"replaceOnChanges"`
}

func (i *EnvironmentRotationScheduleInput) Annotate(a infer.Annotator) {
	a.Describe(&i.Organization, "Organization name.")
	a.Describe(&i.Project, "Project name.")
	a.Describe(&i.Environment, "Environment name.")
	a.Describe(
		&i.ScheduleCron,
		"Cron expression for recurring scheduled rotations. If you are supplying this, do not supply timestamp.",
	)
	a.Describe(
		&i.Timestamp,
		"The time at which the rotation should run, in ISO 8601 format. "+
			"Eg: 2020-01-01T00:00:00Z. If you are supplying this, do not supply scheduleCron.",
	)
}

type EnvironmentRotationScheduleState struct {
	EnvironmentRotationScheduleInput
	ScheduleID string `pulumi:"scheduleId"`
}

func (s *EnvironmentRotationScheduleState) Annotate(a infer.Annotator) {
	a.Describe(&s.ScheduleID, "Schedule ID of the created rotation schedule, assigned by Pulumi Cloud.")
}

func (*EnvironmentRotationSchedule) Check(
	ctx context.Context, req infer.CheckRequest,
) (infer.CheckResponse[EnvironmentRotationScheduleInput], error) {
	i, failures, err := infer.DefaultCheck[EnvironmentRotationScheduleInput](ctx, req.NewInputs)
	if err != nil {
		return infer.CheckResponse[EnvironmentRotationScheduleInput]{}, err
	}
	hasCron := i.ScheduleCron != nil && *i.ScheduleCron != ""
	hasTimestamp := i.Timestamp != nil && *i.Timestamp != ""
	if hasCron == hasTimestamp {
		failures = append(failures, p.CheckFailure{
			Property: "scheduleCron",
			Reason:   "exactly one of scheduleCron or timestamp must be specified",
		})
	}
	if hasTimestamp {
		if _, perr := time.Parse(time.RFC3339, *i.Timestamp); perr != nil {
			failures = append(failures, p.CheckFailure{
				Property: "timestamp",
				Reason:   fmt.Sprintf("timestamp must be in RFC 3339 format: %s", perr),
			})
		}
	}
	return infer.CheckResponse[EnvironmentRotationScheduleInput]{Inputs: i, Failures: failures}, nil
}

func (*EnvironmentRotationSchedule) Create(
	ctx context.Context,
	req infer.CreateRequest[EnvironmentRotationScheduleInput],
) (infer.CreateResponse[EnvironmentRotationScheduleState], error) {
	if req.DryRun {
		return infer.CreateResponse[EnvironmentRotationScheduleState]{
			Output: EnvironmentRotationScheduleState{EnvironmentRotationScheduleInput: req.Inputs},
		}, nil
	}
	env, scheduleReq, err := req.Inputs.toAPI()
	if err != nil {
		return infer.CreateResponse[EnvironmentRotationScheduleState]{}, err
	}
	scheduleID, err := config.GetClient(ctx).CreateEnvironmentRotationSchedule(ctx, env, scheduleReq)
	if err != nil {
		return infer.CreateResponse[EnvironmentRotationScheduleState]{}, fmt.Errorf(
			"error creating environment rotation schedule: %w", err,
		)
	}
	return infer.CreateResponse[EnvironmentRotationScheduleState]{
		ID: environmentScheduleID(env, "rotations", *scheduleID),
		Output: EnvironmentRotationScheduleState{
			EnvironmentRotationScheduleInput: req.Inputs,
			ScheduleID:                       *scheduleID,
		},
	}, nil
}

func (*EnvironmentRotationSchedule) Update(
	ctx context.Context,
	req infer.UpdateRequest[EnvironmentRotationScheduleInput, EnvironmentRotationScheduleState],
) (infer.UpdateResponse[EnvironmentRotationScheduleState], error) {
	if req.DryRun {
		return infer.UpdateResponse[EnvironmentRotationScheduleState]{
			Output: EnvironmentRotationScheduleState{
				EnvironmentRotationScheduleInput: req.Inputs,
				ScheduleID:                       req.State.ScheduleID,
			},
		}, nil
	}
	env, scheduleReq, err := req.Inputs.toAPI()
	if err != nil {
		return infer.UpdateResponse[EnvironmentRotationScheduleState]{}, err
	}
	scheduleID, err := config.GetClient(ctx).UpdateEnvironmentRotationSchedule(
		ctx, env, scheduleReq, req.State.ScheduleID,
	)
	if err != nil {
		return infer.UpdateResponse[EnvironmentRotationScheduleState]{}, fmt.Errorf(
			"error updating environment rotation schedule: %w", err,
		)
	}
	return infer.UpdateResponse[EnvironmentRotationScheduleState]{
		Output: EnvironmentRotationScheduleState{
			EnvironmentRotationScheduleInput: req.Inputs,
			ScheduleID:                       *scheduleID,
		},
	}, nil
}

func (*EnvironmentRotationSchedule) Delete(
	ctx context.Context,
	req infer.DeleteRequest[EnvironmentRotationScheduleState],
) (infer.DeleteResponse, error) {
	env := pulumiapi.EnvironmentIdentifier{
		OrgName:     req.State.Organization,
		ProjectName: req.State.Project,
		EnvName:     req.State.Environment,
	}
	return infer.DeleteResponse{}, config.GetClient(ctx).DeleteEnvironmentSchedule(ctx, env, req.State.ScheduleID)
}

func (*EnvironmentRotationSchedule) Read(
	ctx context.Context,
	req infer.ReadRequest[EnvironmentRotationScheduleInput, EnvironmentRotationScheduleState],
) (infer.ReadResponse[EnvironmentRotationScheduleInput, EnvironmentRotationScheduleState], error) {
	env, scheduleID, err := parseEnvironmentScheduleID(req.ID, "rotations")
	if err != nil {
		return infer.ReadResponse[EnvironmentRotationScheduleInput, EnvironmentRotationScheduleState]{}, err
	}

	resp, err := config.GetClient(ctx).GetEnvironmentSchedule(ctx, env, scheduleID)
	if err != nil {
		return infer.ReadResponse[EnvironmentRotationScheduleInput, EnvironmentRotationScheduleState]{},
			fmt.Errorf("failed to read EnvironmentRotationSchedule (%q): %w", req.ID, err)
	}
	if resp == nil {
		return infer.ReadResponse[EnvironmentRotationScheduleInput, EnvironmentRotationScheduleState]{}, nil
	}

	inputs := EnvironmentRotationScheduleInput{
		Organization: env.OrgName,
		Project:      env.ProjectName,
		Environment:  env.EnvName,
		ScheduleCron: resp.ScheduleCron,
	}
	if resp.ScheduleOnce != nil {
		parsed, err := time.Parse(time.DateTime, *resp.ScheduleOnce)
		if err != nil {
			return infer.ReadResponse[EnvironmentRotationScheduleInput, EnvironmentRotationScheduleState]{},
				fmt.Errorf("failed to read EnvironmentRotationSchedule (%q): %w", req.ID, err)
		}
		ts := parsed.UTC().Format(time.RFC3339)
		inputs.Timestamp = &ts
	}
	return infer.ReadResponse[EnvironmentRotationScheduleInput, EnvironmentRotationScheduleState]{
		ID:     req.ID,
		Inputs: inputs,
		State: EnvironmentRotationScheduleState{
			EnvironmentRotationScheduleInput: inputs,
			ScheduleID:                       scheduleID,
		},
	}, nil
}

func (i EnvironmentRotationScheduleInput) toAPI() (
	pulumiapi.EnvironmentIdentifier, pulumiapi.CreateEnvironmentRotationScheduleRequest, error,
) {
	env := pulumiapi.EnvironmentIdentifier{
		OrgName:     i.Organization,
		ProjectName: i.Project,
		EnvName:     i.Environment,
	}
	scheduleReq := pulumiapi.CreateEnvironmentRotationScheduleRequest{
		ScheduleCron: i.ScheduleCron,
	}
	if i.Timestamp != nil && *i.Timestamp != "" {
		ts, err := time.Parse(time.RFC3339, *i.Timestamp)
		if err != nil {
			return pulumiapi.EnvironmentIdentifier{}, pulumiapi.CreateEnvironmentRotationScheduleRequest{},
				fmt.Errorf("invalid timestamp %q: %w", *i.Timestamp, err)
		}
		scheduleReq.ScheduleOnce = &ts
	}
	return env, scheduleReq, nil
}

func environmentScheduleID(env pulumiapi.EnvironmentIdentifier, scheduleType, scheduleID string) string {
	return fmt.Sprintf("%s/%s/%s/%s/%s", env.OrgName, env.ProjectName, env.EnvName, scheduleType, scheduleID)
}

func parseEnvironmentScheduleID(id, scheduleType string) (pulumiapi.EnvironmentIdentifier, string, error) {
	parts := strings.Split(id, "/")
	if len(parts) != 5 || parts[3] != scheduleType {
		return pulumiapi.EnvironmentIdentifier{}, "",
			fmt.Errorf("%q is invalid, expected organization/project/environment/%s/scheduleId", id, scheduleType)
	}
	return pulumiapi.EnvironmentIdentifier{
		OrgName:     parts[0],
		ProjectName: parts[1],
		EnvName:     parts[2],
	}, parts[4], nil
}
