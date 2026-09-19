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

	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/config"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
)

type Stack struct{}

var (
	_ infer.CustomCreate[StackInput, StackState] = &Stack{}
	_ infer.CustomDelete[StackState]             = &Stack{}
	_ infer.CustomRead[StackInput, StackState]   = &Stack{}
	_ infer.CustomUpdate[StackInput, StackState] = &Stack{}
	_ infer.CustomDiff[StackInput, StackState]   = &Stack{}
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
	// forceDestroy is consulted only at delete time, so a change to it must not
	// replace the stack; Diff and Update below carry that distinction, since
	// without an Update method infer treats every change as a replacement.
	ForceDestroy bool `pulumi:"forceDestroy,optional"`
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

// Diff compares typed values rather than raw property maps so that state
// written by a provider before v1.1.0, which omitted forceDestroy when false,
// decodes to the same false the current inputs carry and produces no diff.
func (*Stack) Diff(
	_ context.Context,
	req infer.DiffRequest[StackInput, StackState],
) (infer.DiffResponse, error) {
	diff := map[string]p.PropertyDiff{}
	replace := func(key string) { diff[key] = p.PropertyDiff{Kind: p.UpdateReplace, InputDiff: true} }

	if req.State.OrganizationName != req.Inputs.OrganizationName {
		replace("organizationName")
	}
	if req.State.ProjectName != req.Inputs.ProjectName {
		replace("projectName")
	}
	if req.State.StackName != req.Inputs.StackName {
		replace("stackName")
	}
	if req.State.ForceDestroy != req.Inputs.ForceDestroy {
		diff["forceDestroy"] = p.PropertyDiff{Kind: p.Update, InputDiff: true}
	}

	return infer.DiffResponse{
		HasChanges:   len(diff) > 0,
		DetailedDiff: diff,
	}, nil
}

// Update handles the one non-replacing input, forceDestroy, which the Pulumi
// Cloud API does not store: the only thing to do is record the new value.
// Anything else reaching here escaped Diff's replace path, so refuse it
// rather than record a change the API never saw.
func (*Stack) Update(
	_ context.Context,
	req infer.UpdateRequest[StackInput, StackState],
) (infer.UpdateResponse[StackState], error) {
	old := req.State.StackInput
	if req.Inputs.OrganizationName != old.OrganizationName ||
		req.Inputs.ProjectName != old.ProjectName ||
		req.Inputs.StackName != old.StackName {
		return infer.UpdateResponse[StackState]{}, fmt.Errorf(
			"stack %q cannot be updated in place; identity changes require a replacement", req.ID)
	}
	return infer.UpdateResponse[StackState]{
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
