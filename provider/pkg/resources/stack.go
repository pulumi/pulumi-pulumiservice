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
	"github.com/pulumi/pulumi/sdk/v3/go/property"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/config"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
)

type Stack struct{}

var (
	_ infer.CustomCreate[StackInput, StackState] = &Stack{}
	_ infer.CustomDelete[StackState]             = &Stack{}
	_ infer.CustomRead[StackInput, StackState]   = &Stack{}
	_ infer.CustomStateMigrations[StackState]    = &Stack{}
)

func (*Stack) Annotate(a infer.Annotator) {
	a.Describe(
		&Stack{},
		"A stack is a collection of resources that share a common lifecycle. "+
			"Stacks are uniquely identified by their name and the project they belong to.",
	)
	a.SetToken("index", "Stack")
}

type StackInput struct {
	OrganizationName string `pulumi:"organizationName" provider:"replaceOnChanges"`
	ProjectName      string `pulumi:"projectName"      provider:"replaceOnChanges"`
	StackName        string `pulumi:"stackName"        provider:"replaceOnChanges"`
	ForceDestroy     bool   `pulumi:"forceDestroy,optional" provider:"replaceOnChanges"`
}

func (i *StackInput) Annotate(a infer.Annotator) {
	a.Describe(&i.OrganizationName, "The name of the organization.")
	a.Describe(&i.ProjectName, "The name of the project.")
	a.Describe(&i.StackName, "The name of the stack.")
	a.Describe(
		&i.ForceDestroy,
		"Optional. Flag indicating whether to delete the stack even if it still contains resources.",
	)
}

type StackState struct {
	StackInput
}

func (*Stack) Create(
	ctx context.Context,
	req infer.CreateRequest[StackInput],
) (infer.CreateResponse[StackState], error) {
	if req.DryRun {
		return infer.CreateResponse[StackState]{
			Output: StackState{StackInput: req.Inputs},
		}, nil
	}
	stackID := pulumiapi.StackIdentifier{
		OrgName:     req.Inputs.OrganizationName,
		ProjectName: req.Inputs.ProjectName,
		StackName:   req.Inputs.StackName,
	}
	if err := config.GetClient(ctx).CreateStack(ctx, stackID); err != nil {
		return infer.CreateResponse[StackState]{}, fmt.Errorf(
			"error creating stack %q: %w", stackID, err,
		)
	}
	return infer.CreateResponse[StackState]{
		ID:     stackResourceID(stackID),
		Output: StackState{StackInput: req.Inputs},
	}, nil
}

func (*Stack) Delete(
	ctx context.Context,
	req infer.DeleteRequest[StackState],
) (infer.DeleteResponse, error) {
	stackID := pulumiapi.StackIdentifier{
		OrgName:     req.State.OrganizationName,
		ProjectName: req.State.ProjectName,
		StackName:   req.State.StackName,
	}
	return infer.DeleteResponse{}, config.GetClient(ctx).DeleteStack(ctx, stackID, req.State.ForceDestroy)
}

func (*Stack) Read(
	ctx context.Context,
	req infer.ReadRequest[StackInput, StackState],
) (infer.ReadResponse[StackInput, StackState], error) {
	orgName, projectName, stackName, err := splitStackResourceID(req.ID)
	if err != nil {
		return infer.ReadResponse[StackInput, StackState]{}, err
	}
	stackID := pulumiapi.StackIdentifier{
		OrgName:     orgName,
		ProjectName: projectName,
		StackName:   stackName,
	}
	exists, err := config.GetClient(ctx).StackExists(ctx, stackID)
	if err != nil {
		return infer.ReadResponse[StackInput, StackState]{}, fmt.Errorf(
			"failure while checking if stack %q exists: %w", req.ID, err,
		)
	}
	if !exists {
		return infer.ReadResponse[StackInput, StackState]{}, nil
	}
	inputs := StackInput{
		OrganizationName: orgName,
		ProjectName:      projectName,
		StackName:        stackName,
		// forceDestroy is a write-only delete-time hint that does not round-trip
		// through the Pulumi Cloud API; preserve whatever the user configured.
		ForceDestroy: req.Inputs.ForceDestroy,
	}
	return infer.ReadResponse[StackInput, StackState]{
		ID:     req.ID,
		Inputs: inputs,
		State:  StackState{StackInput: inputs},
	}, nil
}

// StateMigrations strips the legacy `__inputs` field that the pre-infer Stack
// resource embedded in state. infer rejects unknown fields when decoding state,
// so without this migration a refresh against an existing stack errors with
// "Unrecognized field '__inputs'".
func (*Stack) StateMigrations(context.Context) []infer.StateMigrationFunc[StackState] {
	return []infer.StateMigrationFunc[StackState]{
		infer.StateMigration(migrateStackLegacyState),
	}
}

func migrateStackLegacyState(
	_ context.Context, old property.Map,
) (infer.MigrationResult[StackState], error) {
	if _, ok := old.GetOk("__inputs"); !ok {
		return infer.MigrationResult[StackState]{}, nil
	}
	state := StackState{}
	if v, ok := old.GetOk("organizationName"); ok && v.IsString() {
		state.OrganizationName = v.AsString()
	}
	if v, ok := old.GetOk("projectName"); ok && v.IsString() {
		state.ProjectName = v.AsString()
	}
	if v, ok := old.GetOk("stackName"); ok && v.IsString() {
		state.StackName = v.AsString()
	}
	if v, ok := old.GetOk("forceDestroy"); ok && v.IsBool() {
		state.ForceDestroy = v.AsBool()
	}
	return infer.MigrationResult[StackState]{Result: &state}, nil
}

func stackResourceID(stack pulumiapi.StackIdentifier) string {
	return fmt.Sprintf("%s/%s/%s", stack.OrgName, stack.ProjectName, stack.StackName)
}

func splitStackResourceID(id string) (orgName, projectName, stackName string, err error) {
	s := strings.Split(id, "/")
	if len(s) != 3 {
		return "", "", "", fmt.Errorf(
			"%q is invalid, must be in the format: organization/project/stack", id,
		)
	}
	return s[0], s[1], s[2], nil
}
