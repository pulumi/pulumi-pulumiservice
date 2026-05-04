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
	"time"

	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/config"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
)

type TTLSchedule struct{}

var (
	_ infer.CustomCreate[TTLScheduleInput, TTLScheduleState] = &TTLSchedule{}
	_ infer.CustomUpdate[TTLScheduleInput, TTLScheduleState] = &TTLSchedule{}
	_ infer.CustomDelete[TTLScheduleState]                   = &TTLSchedule{}
	_ infer.CustomRead[TTLScheduleInput, TTLScheduleState]   = &TTLSchedule{}
)

func (*TTLSchedule) Annotate(a infer.Annotator) {
	a.Describe(&TTLSchedule{}, "A scheduled stack destroy run.")
	a.SetToken("index", "TtlSchedule")
}

type TTLScheduleInput struct {
	Organization       string `pulumi:"organization" provider:"replaceOnChanges"`
	Project            string `pulumi:"project"      provider:"replaceOnChanges"`
	Stack              string `pulumi:"stack"        provider:"replaceOnChanges"`
	Timestamp          string `pulumi:"timestamp"    provider:"replaceOnChanges"`
	DeleteAfterDestroy *bool  `pulumi:"deleteAfterDestroy,optional"`
}

func (i *TTLScheduleInput) Annotate(a infer.Annotator) {
	a.Describe(&i.Organization, "Organization name.")
	a.Describe(&i.Project, "Project name.")
	a.Describe(&i.Stack, "Stack name.")
	a.Describe(&i.Timestamp, "The time at which the schedule should run, in ISO 8601 format. Eg: 2020-01-01T00:00:00Z.")
	a.Describe(
		&i.DeleteAfterDestroy,
		"True if the stack and all associated history and settings should be deleted.",
	)
	a.SetDefault(&i.DeleteAfterDestroy, false)
}

type TTLScheduleState struct {
	TTLScheduleInput
	ScheduleID string `pulumi:"scheduleId"`
}

func (s *TTLScheduleState) Annotate(a infer.Annotator) {
	a.Describe(&s.ScheduleID, "Schedule ID of the created schedule, assigned by Pulumi Cloud.")
}

func (*TTLSchedule) Check(
	ctx context.Context, req infer.CheckRequest,
) (infer.CheckResponse[TTLScheduleInput], error) {
	i, failures, err := infer.DefaultCheck[TTLScheduleInput](ctx, req.NewInputs)
	if err != nil {
		return infer.CheckResponse[TTLScheduleInput]{}, err
	}
	if i.Timestamp != "" {
		if _, perr := time.Parse(time.RFC3339, i.Timestamp); perr != nil {
			failures = append(failures, p.CheckFailure{
				Property: "timestamp",
				Reason:   fmt.Sprintf("timestamp must be in RFC 3339 format: %s", perr),
			})
		}
	}
	return infer.CheckResponse[TTLScheduleInput]{Inputs: i, Failures: failures}, nil
}

func (*TTLSchedule) Create(
	ctx context.Context,
	req infer.CreateRequest[TTLScheduleInput],
) (infer.CreateResponse[TTLScheduleState], error) {
	if req.DryRun {
		return infer.CreateResponse[TTLScheduleState]{
			Output: TTLScheduleState{TTLScheduleInput: req.Inputs},
		}, nil
	}
	stack, ts, deleteAfterDestroy, err := req.Inputs.toAPI()
	if err != nil {
		return infer.CreateResponse[TTLScheduleState]{}, err
	}
	scheduleID, err := config.GetClient(ctx).CreateTTLSchedule(ctx, stack, pulumiapi.CreateTTLScheduleRequest{
		Timestamp:          ts,
		DeleteAfterDestroy: deleteAfterDestroy,
	})
	if err != nil {
		return infer.CreateResponse[TTLScheduleState]{}, fmt.Errorf("error creating TTL schedule: %w", err)
	}
	return infer.CreateResponse[TTLScheduleState]{
		ID: stackScheduleID(stack, "ttl", *scheduleID),
		Output: TTLScheduleState{
			TTLScheduleInput: req.Inputs,
			ScheduleID:       *scheduleID,
		},
	}, nil
}

func (*TTLSchedule) Update(
	ctx context.Context,
	req infer.UpdateRequest[TTLScheduleInput, TTLScheduleState],
) (infer.UpdateResponse[TTLScheduleState], error) {
	if req.DryRun {
		return infer.UpdateResponse[TTLScheduleState]{
			Output: TTLScheduleState{
				TTLScheduleInput: req.Inputs,
				ScheduleID:       req.State.ScheduleID,
			},
		}, nil
	}
	stack, ts, deleteAfterDestroy, err := req.Inputs.toAPI()
	if err != nil {
		return infer.UpdateResponse[TTLScheduleState]{}, err
	}
	scheduleID, err := config.GetClient(ctx).UpdateTTLSchedule(ctx, stack, pulumiapi.CreateTTLScheduleRequest{
		Timestamp:          ts,
		DeleteAfterDestroy: deleteAfterDestroy,
	}, req.State.ScheduleID)
	if err != nil {
		return infer.UpdateResponse[TTLScheduleState]{}, fmt.Errorf("error updating TTL schedule: %w", err)
	}
	return infer.UpdateResponse[TTLScheduleState]{
		Output: TTLScheduleState{
			TTLScheduleInput: req.Inputs,
			ScheduleID:       *scheduleID,
		},
	}, nil
}

func (*TTLSchedule) Delete(
	ctx context.Context,
	req infer.DeleteRequest[TTLScheduleState],
) (infer.DeleteResponse, error) {
	stack := pulumiapi.StackIdentifier{
		OrgName:     req.State.Organization,
		ProjectName: req.State.Project,
		StackName:   req.State.Stack,
	}
	return infer.DeleteResponse{}, config.GetClient(ctx).DeleteStackSchedule(ctx, stack, req.State.ScheduleID)
}

func (*TTLSchedule) Read(
	ctx context.Context,
	req infer.ReadRequest[TTLScheduleInput, TTLScheduleState],
) (infer.ReadResponse[TTLScheduleInput, TTLScheduleState], error) {
	stack, scheduleID, err := parseStackScheduleID(req.ID, "ttl")
	if err != nil {
		return infer.ReadResponse[TTLScheduleInput, TTLScheduleState]{}, err
	}

	resp, err := config.GetClient(ctx).GetStackSchedule(ctx, stack, scheduleID)
	if err != nil {
		return infer.ReadResponse[TTLScheduleInput, TTLScheduleState]{}, fmt.Errorf(
			"failed to read TtlSchedule (%q): %w", req.ID, err,
		)
	}
	if resp == nil {
		return infer.ReadResponse[TTLScheduleInput, TTLScheduleState]{}, nil
	}
	if resp.ScheduleOnce == nil {
		return infer.ReadResponse[TTLScheduleInput, TTLScheduleState]{}, fmt.Errorf(
			"TtlSchedule (%q) has no scheduleOnce timestamp", req.ID,
		)
	}
	parsed, err := time.Parse(time.DateTime, *resp.ScheduleOnce)
	if err != nil {
		return infer.ReadResponse[TTLScheduleInput, TTLScheduleState]{}, fmt.Errorf(
			"failed to read TtlSchedule (%q): %w", req.ID, err,
		)
	}
	deleteAfterDestroy := resp.Definition.Request.OperationContext.Options.DeleteAfterDestroy
	inputs := TTLScheduleInput{
		Organization:       stack.OrgName,
		Project:            stack.ProjectName,
		Stack:              stack.StackName,
		Timestamp:          parsed.UTC().Format(time.RFC3339),
		DeleteAfterDestroy: &deleteAfterDestroy,
	}
	return infer.ReadResponse[TTLScheduleInput, TTLScheduleState]{
		ID:     req.ID,
		Inputs: inputs,
		State: TTLScheduleState{
			TTLScheduleInput: inputs,
			ScheduleID:       scheduleID,
		},
	}, nil
}

func (i TTLScheduleInput) toAPI() (pulumiapi.StackIdentifier, time.Time, bool, error) {
	stack := pulumiapi.StackIdentifier{
		OrgName:     i.Organization,
		ProjectName: i.Project,
		StackName:   i.Stack,
	}
	ts, err := time.Parse(time.RFC3339, i.Timestamp)
	if err != nil {
		return pulumiapi.StackIdentifier{}, time.Time{}, false, fmt.Errorf("invalid timestamp %q: %w", i.Timestamp, err)
	}
	deleteAfterDestroy := false
	if i.DeleteAfterDestroy != nil {
		deleteAfterDestroy = *i.DeleteAfterDestroy
	}
	return stack, ts, deleteAfterDestroy, nil
}

func stackScheduleID(stack pulumiapi.StackIdentifier, scheduleType, scheduleID string) string {
	return fmt.Sprintf("%s/%s/%s/%s/%s", stack.OrgName, stack.ProjectName, stack.StackName, scheduleType, scheduleID)
}

func parseStackScheduleID(id, scheduleType string) (pulumiapi.StackIdentifier, string, error) {
	stack, sched, err := ParseStackScheduleID(id, scheduleType)
	if err != nil {
		return pulumiapi.StackIdentifier{}, "", err
	}
	return *stack, *sched, nil
}
