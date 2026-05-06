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

type PulumiOperation string

const (
	PulumiOperationUpdate  PulumiOperation = "update"
	PulumiOperationPreview PulumiOperation = "preview"
	PulumiOperationRefresh PulumiOperation = "refresh"
	PulumiOperationDestroy PulumiOperation = "destroy"
)

func (PulumiOperation) Values() []infer.EnumValue[PulumiOperation] {
	return []infer.EnumValue[PulumiOperation]{
		{Name: "update", Value: PulumiOperationUpdate, Description: "Analogous to `pulumi up` command."},
		{Name: "preview", Value: PulumiOperationPreview, Description: "Analogous to `pulumi preview` command."},
		{Name: "refresh", Value: PulumiOperationRefresh, Description: "Analogous to `pulumi refresh` command."},
		{Name: "destroy", Value: PulumiOperationDestroy, Description: "Analogous to `pulumi destroy` command."},
	}
}

type DeploymentSchedule struct{}

type (
	depIn  = DeploymentScheduleInput
	depOut = DeploymentScheduleState
)

var (
	_ infer.CustomCheck[depIn]          = &DeploymentSchedule{}
	_ infer.CustomCreate[depIn, depOut] = &DeploymentSchedule{}
	_ infer.CustomUpdate[depIn, depOut] = &DeploymentSchedule{}
	_ infer.CustomDelete[depOut]        = &DeploymentSchedule{}
	_ infer.CustomRead[depIn, depOut]   = &DeploymentSchedule{}
)

func (*DeploymentSchedule) Annotate(a infer.Annotator) {
	a.Describe(&DeploymentSchedule{}, "A scheduled recurring or single time run of a pulumi command.")
	a.SetToken("index", "DeploymentSchedule")
}

type DeploymentScheduleInput struct {
	Organization    string          `pulumi:"organization"          provider:"replaceOnChanges"`
	Project         string          `pulumi:"project"               provider:"replaceOnChanges"`
	Stack           string          `pulumi:"stack"                 provider:"replaceOnChanges"`
	ScheduleCron    *string         `pulumi:"scheduleCron,optional"`
	Timestamp       *string         `pulumi:"timestamp,optional"    provider:"replaceOnChanges"`
	PulumiOperation PulumiOperation `pulumi:"pulumiOperation"`
}

func (i *DeploymentScheduleInput) Annotate(a infer.Annotator) {
	a.Describe(&i.Organization, "Organization name.")
	a.Describe(&i.Project, "Project name.")
	a.Describe(&i.Stack, "Stack name.")
	a.Describe(
		&i.ScheduleCron,
		"Cron expression for recurring scheduled runs. If you are supplying this, do not supply timestamp.",
	)
	a.Describe(
		&i.Timestamp,
		"The time at which the schedule should run, in ISO 8601 format. "+
			"Eg: 2020-01-01T00:00:00Z. If you are supplying this, do not supply scheduleCron.",
	)
	a.Describe(&i.PulumiOperation, "Which command to run.")
}

type DeploymentScheduleState struct {
	DeploymentScheduleInput
	ScheduleID string `pulumi:"scheduleId"`
}

func (s *DeploymentScheduleState) Annotate(a infer.Annotator) {
	a.Describe(&s.ScheduleID, "Schedule ID of the created schedule, assigned by Pulumi Cloud.")
}

func (*DeploymentSchedule) Check(
	ctx context.Context, req infer.CheckRequest,
) (infer.CheckResponse[DeploymentScheduleInput], error) {
	i, failures, err := infer.DefaultCheck[DeploymentScheduleInput](ctx, req.NewInputs)
	if err != nil {
		return infer.CheckResponse[DeploymentScheduleInput]{}, err
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
	return infer.CheckResponse[DeploymentScheduleInput]{Inputs: i, Failures: failures}, nil
}

func (*DeploymentSchedule) Create(
	ctx context.Context,
	req infer.CreateRequest[DeploymentScheduleInput],
) (infer.CreateResponse[DeploymentScheduleState], error) {
	if req.DryRun {
		return infer.CreateResponse[DeploymentScheduleState]{
			Output: DeploymentScheduleState{DeploymentScheduleInput: req.Inputs},
		}, nil
	}
	stack, scheduleReq, err := req.Inputs.toAPI()
	if err != nil {
		return infer.CreateResponse[DeploymentScheduleState]{}, err
	}
	scheduleID, err := config.GetClient(ctx).CreateDeploymentSchedule(ctx, stack, scheduleReq)
	if err != nil {
		return infer.CreateResponse[DeploymentScheduleState]{}, fmt.Errorf("error creating deployment schedule: %w", err)
	}
	return infer.CreateResponse[DeploymentScheduleState]{
		ID: deploymentScheduleID(stack, *scheduleID),
		Output: DeploymentScheduleState{
			DeploymentScheduleInput: req.Inputs,
			ScheduleID:              *scheduleID,
		},
	}, nil
}

func (*DeploymentSchedule) Update(
	ctx context.Context,
	req infer.UpdateRequest[DeploymentScheduleInput, DeploymentScheduleState],
) (infer.UpdateResponse[DeploymentScheduleState], error) {
	if req.DryRun {
		return infer.UpdateResponse[DeploymentScheduleState]{
			Output: DeploymentScheduleState{
				DeploymentScheduleInput: req.Inputs,
				ScheduleID:              req.State.ScheduleID,
			},
		}, nil
	}
	stack, scheduleReq, err := req.Inputs.toAPI()
	if err != nil {
		return infer.UpdateResponse[DeploymentScheduleState]{}, err
	}
	scheduleID, err := config.GetClient(ctx).UpdateDeploymentSchedule(
		ctx, stack, scheduleReq, req.State.ScheduleID,
	)
	if err != nil {
		return infer.UpdateResponse[DeploymentScheduleState]{}, fmt.Errorf("error updating deployment schedule: %w", err)
	}
	return infer.UpdateResponse[DeploymentScheduleState]{
		Output: DeploymentScheduleState{
			DeploymentScheduleInput: req.Inputs,
			ScheduleID:              *scheduleID,
		},
	}, nil
}

func (*DeploymentSchedule) Delete(
	ctx context.Context,
	req infer.DeleteRequest[DeploymentScheduleState],
) (infer.DeleteResponse, error) {
	stack := pulumiapi.StackIdentifier{
		OrgName:     req.State.Organization,
		ProjectName: req.State.Project,
		StackName:   req.State.Stack,
	}
	return infer.DeleteResponse{}, config.GetClient(ctx).DeleteStackSchedule(ctx, stack, req.State.ScheduleID)
}

func (*DeploymentSchedule) Read(
	ctx context.Context,
	req infer.ReadRequest[DeploymentScheduleInput, DeploymentScheduleState],
) (infer.ReadResponse[DeploymentScheduleInput, DeploymentScheduleState], error) {
	stack, scheduleID, err := parseDeploymentScheduleID(req.ID)
	if err != nil {
		return infer.ReadResponse[DeploymentScheduleInput, DeploymentScheduleState]{}, err
	}

	resp, err := config.GetClient(ctx).GetStackSchedule(ctx, stack, scheduleID)
	if err != nil {
		return infer.ReadResponse[DeploymentScheduleInput, DeploymentScheduleState]{}, fmt.Errorf(
			"failed to read DeploymentSchedule (%q): %w", req.ID, err,
		)
	}
	if resp == nil {
		return infer.ReadResponse[DeploymentScheduleInput, DeploymentScheduleState]{}, nil
	}

	inputs := DeploymentScheduleInput{
		Organization:    stack.OrgName,
		Project:         stack.ProjectName,
		Stack:           stack.StackName,
		ScheduleCron:    resp.ScheduleCron,
		PulumiOperation: PulumiOperation(resp.Definition.Request.PulumiOperation),
	}
	if resp.ScheduleOnce != nil {
		parsed, err := time.Parse(time.DateTime, *resp.ScheduleOnce)
		if err != nil {
			return infer.ReadResponse[DeploymentScheduleInput, DeploymentScheduleState]{}, fmt.Errorf(
				"failed to read DeploymentSchedule (%q): %w", req.ID, err,
			)
		}
		ts := parsed.UTC().Format(time.RFC3339)
		inputs.Timestamp = &ts
	}
	return infer.ReadResponse[DeploymentScheduleInput, DeploymentScheduleState]{
		ID:     req.ID,
		Inputs: inputs,
		State: DeploymentScheduleState{
			DeploymentScheduleInput: inputs,
			ScheduleID:              scheduleID,
		},
	}, nil
}

func (i DeploymentScheduleInput) toAPI() (
	pulumiapi.StackIdentifier, pulumiapi.CreateDeploymentScheduleRequest, error,
) {
	stack := pulumiapi.StackIdentifier{
		OrgName:     i.Organization,
		ProjectName: i.Project,
		StackName:   i.Stack,
	}
	scheduleReq := pulumiapi.CreateDeploymentScheduleRequest{
		ScheduleCron: i.ScheduleCron,
		Request: pulumiapi.CreateDeploymentRequest{
			PulumiOperation: string(i.PulumiOperation),
		},
	}
	if i.Timestamp != nil && *i.Timestamp != "" {
		ts, err := time.Parse(time.RFC3339, *i.Timestamp)
		if err != nil {
			return pulumiapi.StackIdentifier{}, pulumiapi.CreateDeploymentScheduleRequest{},
				fmt.Errorf("invalid timestamp %q: %w", *i.Timestamp, err)
		}
		scheduleReq.ScheduleOnce = &ts
	}
	return stack, scheduleReq, nil
}

func deploymentScheduleID(stack pulumiapi.StackIdentifier, scheduleID string) string {
	return fmt.Sprintf("%s/%s/%s/%s", stack.OrgName, stack.ProjectName, stack.StackName, scheduleID)
}

func parseDeploymentScheduleID(id string) (pulumiapi.StackIdentifier, string, error) {
	parts := strings.Split(id, "/")
	if len(parts) != 4 {
		return pulumiapi.StackIdentifier{}, "",
			fmt.Errorf("%q is invalid, expected organization/project/stack/scheduleId", id)
	}
	return pulumiapi.StackIdentifier{
		OrgName:     parts[0],
		ProjectName: parts[1],
		StackName:   parts[2],
	}, parts[3], nil
}
