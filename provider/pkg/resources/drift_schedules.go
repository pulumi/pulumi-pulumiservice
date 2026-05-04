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

	"github.com/pulumi/pulumi-go-provider/infer"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/config"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
)

type DriftSchedule struct{}

var (
	_ infer.CustomCreate[DriftScheduleInput, DriftScheduleState] = &DriftSchedule{}
	_ infer.CustomUpdate[DriftScheduleInput, DriftScheduleState] = &DriftSchedule{}
	_ infer.CustomDelete[DriftScheduleState]                     = &DriftSchedule{}
	_ infer.CustomRead[DriftScheduleInput, DriftScheduleState]   = &DriftSchedule{}
)

func (*DriftSchedule) Annotate(a infer.Annotator) {
	a.Describe(&DriftSchedule{}, "A cron schedule to run drift detection.")
	a.SetToken("index", "DriftSchedule")
}

type DriftScheduleInput struct {
	Organization  string `pulumi:"organization" provider:"replaceOnChanges"`
	Project       string `pulumi:"project"      provider:"replaceOnChanges"`
	Stack         string `pulumi:"stack"        provider:"replaceOnChanges"`
	ScheduleCron  string `pulumi:"scheduleCron"`
	AutoRemediate *bool  `pulumi:"autoRemediate,optional"`
}

func (i *DriftScheduleInput) Annotate(a infer.Annotator) {
	a.Describe(&i.Organization, "Organization name.")
	a.Describe(&i.Project, "Project name.")
	a.Describe(&i.Stack, "Stack name.")
	a.Describe(&i.ScheduleCron, "Cron expression for when to run drift detection.")
	a.Describe(&i.AutoRemediate, "Whether any drift detected should be remediated after a drift run.")
	a.SetDefault(&i.AutoRemediate, false)
}

type DriftScheduleState struct {
	DriftScheduleInput
	ScheduleID string `pulumi:"scheduleId"`
}

func (s *DriftScheduleState) Annotate(a infer.Annotator) {
	a.Describe(&s.ScheduleID, "Schedule ID of the created schedule, assigned by Pulumi Cloud.")
}

func (*DriftSchedule) Create(
	ctx context.Context,
	req infer.CreateRequest[DriftScheduleInput],
) (infer.CreateResponse[DriftScheduleState], error) {
	if req.DryRun {
		return infer.CreateResponse[DriftScheduleState]{
			Output: DriftScheduleState{DriftScheduleInput: req.Inputs},
		}, nil
	}
	stack, autoRemediate := req.Inputs.toAPI()
	scheduleID, err := config.GetClient(ctx).CreateDriftSchedule(ctx, stack, pulumiapi.CreateDriftScheduleRequest{
		ScheduleCron:  req.Inputs.ScheduleCron,
		AutoRemediate: autoRemediate,
	})
	if err != nil {
		return infer.CreateResponse[DriftScheduleState]{}, fmt.Errorf("error creating drift schedule: %w", err)
	}
	return infer.CreateResponse[DriftScheduleState]{
		ID: stackScheduleID(stack, "drift", *scheduleID),
		Output: DriftScheduleState{
			DriftScheduleInput: req.Inputs,
			ScheduleID:         *scheduleID,
		},
	}, nil
}

func (*DriftSchedule) Update(
	ctx context.Context,
	req infer.UpdateRequest[DriftScheduleInput, DriftScheduleState],
) (infer.UpdateResponse[DriftScheduleState], error) {
	if req.DryRun {
		return infer.UpdateResponse[DriftScheduleState]{
			Output: DriftScheduleState{
				DriftScheduleInput: req.Inputs,
				ScheduleID:         req.State.ScheduleID,
			},
		}, nil
	}
	stack, autoRemediate := req.Inputs.toAPI()
	scheduleID, err := config.GetClient(ctx).UpdateDriftSchedule(ctx, stack, pulumiapi.CreateDriftScheduleRequest{
		ScheduleCron:  req.Inputs.ScheduleCron,
		AutoRemediate: autoRemediate,
	}, req.State.ScheduleID)
	if err != nil {
		return infer.UpdateResponse[DriftScheduleState]{}, fmt.Errorf("error updating drift schedule: %w", err)
	}
	return infer.UpdateResponse[DriftScheduleState]{
		Output: DriftScheduleState{
			DriftScheduleInput: req.Inputs,
			ScheduleID:         *scheduleID,
		},
	}, nil
}

func (*DriftSchedule) Delete(
	ctx context.Context,
	req infer.DeleteRequest[DriftScheduleState],
) (infer.DeleteResponse, error) {
	stack := pulumiapi.StackIdentifier{
		OrgName:     req.State.Organization,
		ProjectName: req.State.Project,
		StackName:   req.State.Stack,
	}
	return infer.DeleteResponse{}, config.GetClient(ctx).DeleteStackSchedule(ctx, stack, req.State.ScheduleID)
}

func (*DriftSchedule) Read(
	ctx context.Context,
	req infer.ReadRequest[DriftScheduleInput, DriftScheduleState],
) (infer.ReadResponse[DriftScheduleInput, DriftScheduleState], error) {
	stack, scheduleID, err := parseStackScheduleID(req.ID, "drift")
	if err != nil {
		return infer.ReadResponse[DriftScheduleInput, DriftScheduleState]{}, err
	}

	resp, err := config.GetClient(ctx).GetStackSchedule(ctx, stack, scheduleID)
	if err != nil {
		return infer.ReadResponse[DriftScheduleInput, DriftScheduleState]{}, fmt.Errorf(
			"failed to read DriftSchedule (%q): %w", req.ID, err,
		)
	}
	if resp == nil {
		return infer.ReadResponse[DriftScheduleInput, DriftScheduleState]{}, nil
	}
	if resp.ScheduleCron == nil {
		return infer.ReadResponse[DriftScheduleInput, DriftScheduleState]{}, fmt.Errorf(
			"DriftSchedule (%q) has no scheduleCron", req.ID,
		)
	}
	autoRemediate := resp.Definition.Request.OperationContext.Options.AutoRemediate
	inputs := DriftScheduleInput{
		Organization:  stack.OrgName,
		Project:       stack.ProjectName,
		Stack:         stack.StackName,
		ScheduleCron:  *resp.ScheduleCron,
		AutoRemediate: &autoRemediate,
	}
	return infer.ReadResponse[DriftScheduleInput, DriftScheduleState]{
		ID:     req.ID,
		Inputs: inputs,
		State: DriftScheduleState{
			DriftScheduleInput: inputs,
			ScheduleID:         scheduleID,
		},
	}, nil
}

func (i DriftScheduleInput) toAPI() (pulumiapi.StackIdentifier, bool) {
	stack := pulumiapi.StackIdentifier{
		OrgName:     i.Organization,
		ProjectName: i.Project,
		StackName:   i.Stack,
	}
	autoRemediate := false
	if i.AutoRemediate != nil {
		autoRemediate = *i.AutoRemediate
	}
	return stack, autoRemediate
}
