// Copyright 2016-2025, Pulumi Corporation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package resources contains implementations of Pulumi Service resources.
package resources

import (
	"context"
	"fmt"

	pbempty "google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
)

// PulumiServiceOrgMemberResource manages organization member membership and roles in Pulumi Cloud.
type PulumiServiceOrgMemberResource struct {
	Client *pulumiapi.Client
}

// PulumiServiceOrgMemberInput represents the input properties for an organization member resource.
type PulumiServiceOrgMemberInput struct {
	OrganizationName string
	UserName         string
	Role             string
}

func (i *PulumiServiceOrgMemberInput) ToPropertyMap() resource.PropertyMap {
	pm := resource.PropertyMap{}
	pm["organizationName"] = resource.NewPropertyValue(i.OrganizationName)
	pm["userName"] = resource.NewPropertyValue(i.UserName)
	pm["role"] = resource.NewPropertyValue(i.Role)
	return pm
}

func (r *PulumiServiceOrgMemberResource) ToPulumiServiceOrgMemberInput(inputMap resource.PropertyMap) PulumiServiceOrgMemberInput {
	input := PulumiServiceOrgMemberInput{}

	if inputMap["organizationName"].HasValue() && inputMap["organizationName"].IsString() {
		input.OrganizationName = inputMap["organizationName"].StringValue()
	}

	if inputMap["userName"].HasValue() && inputMap["userName"].IsString() {
		input.UserName = inputMap["userName"].StringValue()
	}

	if inputMap["role"].HasValue() && inputMap["role"].IsString() {
		input.Role = inputMap["role"].StringValue()
	}

	return input
}

func GenerateOrgMemberProperties(input PulumiServiceOrgMemberInput, member pulumiapi.Member) (outputs *structpb.Struct, inputs *structpb.Struct, err error) {
	inputMap := input.ToPropertyMap()

	outputMap := inputMap.Copy()
	outputMap["__inputs"] = resource.NewObjectProperty(inputMap)
	// Use the actual role from the API response, which may differ from input if the API
	// normalized it or if the user has special permissions (e.g., virtual admin)
	outputMap["role"] = resource.NewPropertyValue(member.Role)
	outputMap["knownToPulumi"] = resource.NewPropertyValue(member.KnownToPulumi)
	outputMap["virtualAdmin"] = resource.NewPropertyValue(member.VirtualAdmin)

	inputs, err = plugin.MarshalProperties(inputMap, plugin.MarshalOptions{})
	if err != nil {
		return nil, nil, err
	}

	outputs, err = plugin.MarshalProperties(outputMap, plugin.MarshalOptions{})
	if err != nil {
		return nil, nil, err
	}

	return outputs, inputs, nil
}

func (r *PulumiServiceOrgMemberResource) Name() string {
	return "pulumiservice:index:OrgMember"
}

func (r *PulumiServiceOrgMemberResource) Check(req *pulumirpc.CheckRequest) (*pulumirpc.CheckResponse, error) {
	inputMap, err := plugin.UnmarshalProperties(req.GetNews(), plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	input := r.ToPulumiServiceOrgMemberInput(inputMap)

	var failures []*pulumirpc.CheckFailure

	// Validate required fields
	if input.OrganizationName == "" {
		failures = append(failures, &pulumirpc.CheckFailure{
			Property: "organizationName",
			Reason:   "organizationName is required",
		})
	}

	if input.UserName == "" {
		failures = append(failures, &pulumirpc.CheckFailure{
			Property: "userName",
			Reason:   "userName is required",
		})
	}

	if input.Role == "" {
		failures = append(failures, &pulumirpc.CheckFailure{
			Property: "role",
			Reason:   "role is required",
		})
	} else if input.Role != "admin" && input.Role != "member" {
		// Validate role value if provided
		failures = append(failures, &pulumirpc.CheckFailure{
			Property: "role",
			Reason:   "role must be either 'admin' or 'member'",
		})
	}

	if len(failures) > 0 {
		return &pulumirpc.CheckResponse{
			Inputs:   req.News,
			Failures: failures,
		}, nil
	}

	return &pulumirpc.CheckResponse{Inputs: req.News, Failures: nil}, nil
}

func (r *PulumiServiceOrgMemberResource) Create(req *pulumirpc.CreateRequest) (*pulumirpc.CreateResponse, error) {
	ctx := context.Background()

	inputMap, err := plugin.UnmarshalProperties(req.GetProperties(), plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	input := r.ToPulumiServiceOrgMemberInput(inputMap)

	// Add member to organization
	err = r.Client.AddMemberToOrg(ctx, input.UserName, input.OrganizationName, input.Role)
	if err != nil {
		return nil, fmt.Errorf("failed to add member to organization: %w", err)
	}

	// Generate ID
	id := fmt.Sprintf("%s/%s", input.OrganizationName, input.UserName)

	// Read back to get full member details
	members, err := r.Client.ListOrgMembers(ctx, input.OrganizationName)
	if err != nil {
		return nil, fmt.Errorf("failed to read member after creation: %w", err)
	}

	// Find the member we just created
	var foundMember *pulumiapi.Member
	for _, member := range members.Members {
		if member.User.Name == input.UserName {
			foundMember = &member
			break
		}
	}

	if foundMember == nil {
		return nil, fmt.Errorf("failed to find member '%s' in organization '%s' after creation", input.UserName, input.OrganizationName)
	}

	// Generate properties. Note: inputs return value is intentionally unused here.
	// The pulumirpc.CreateResponse only requires Properties, and the inputs are
	// embedded in the outputs via __inputs field for state tracking.
	outputs, _, err := GenerateOrgMemberProperties(input, *foundMember)
	if err != nil {
		return nil, err
	}

	return &pulumirpc.CreateResponse{
		Id:         id,
		Properties: outputs,
	}, nil
}

func (r *PulumiServiceOrgMemberResource) Read(req *pulumirpc.ReadRequest) (*pulumirpc.ReadResponse, error) {
	ctx := context.Background()

	// Split ID into org and username
	orgName, userName, err := splitSingleSlashString(req.Id)
	if err != nil {
		return nil, fmt.Errorf("invalid ID format, must contain a single slash: %w", err)
	}

	// List all members and filter
	members, err := r.Client.ListOrgMembers(ctx, orgName)
	if err != nil {
		return nil, fmt.Errorf("failed to list organization members: %w", err)
	}

	// Find the specific member
	var foundMember *pulumiapi.Member
	for _, member := range members.Members {
		if member.User.Name == userName {
			foundMember = &member
			break
		}
	}

	// If member not found, return empty response (resource deleted)
	if foundMember == nil {
		return &pulumirpc.ReadResponse{}, nil
	}

	// Generate input for property generation
	input := PulumiServiceOrgMemberInput{
		OrganizationName: orgName,
		UserName:         userName,
		Role:             foundMember.Role,
	}

	outputs, inputs, err := GenerateOrgMemberProperties(input, *foundMember)
	if err != nil {
		return nil, fmt.Errorf("failed to generate properties: %w", err)
	}

	return &pulumirpc.ReadResponse{
		Id:         req.Id,
		Properties: outputs,
		Inputs:     inputs,
	}, nil
}

func (r *PulumiServiceOrgMemberResource) Update(req *pulumirpc.UpdateRequest) (*pulumirpc.UpdateResponse, error) {
	ctx := context.Background()

	// Unmarshal new properties
	newInputMap, err := plugin.UnmarshalProperties(req.GetNews(), plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	newInput := r.ToPulumiServiceOrgMemberInput(newInputMap)

	// Update role (uses same API endpoint as creation)
	err = r.Client.AddMemberToOrg(ctx, newInput.UserName, newInput.OrganizationName, newInput.Role)
	if err != nil {
		return nil, fmt.Errorf("failed to update member role: %w", err)
	}

	// Read back to get updated member details
	members, err := r.Client.ListOrgMembers(ctx, newInput.OrganizationName)
	if err != nil {
		return nil, fmt.Errorf("failed to read member after update: %w", err)
	}

	// Find the updated member
	var foundMember *pulumiapi.Member
	for _, member := range members.Members {
		if member.User.Name == newInput.UserName {
			foundMember = &member
			break
		}
	}

	if foundMember == nil {
		return nil, fmt.Errorf("failed to find member '%s' in organization '%s' after update", newInput.UserName, newInput.OrganizationName)
	}

	// Generate properties. inputs return value is intentionally unused (see Create for explanation).
	outputs, _, err := GenerateOrgMemberProperties(newInput, *foundMember)
	if err != nil {
		return nil, err
	}

	return &pulumirpc.UpdateResponse{
		Properties: outputs,
	}, nil
}

func (r *PulumiServiceOrgMemberResource) Delete(req *pulumirpc.DeleteRequest) (*pbempty.Empty, error) {
	ctx := context.Background()

	// Split ID into org and username
	orgName, userName, err := splitSingleSlashString(req.Id)
	if err != nil {
		return &pbempty.Empty{}, fmt.Errorf("invalid ID format: %w", err)
	}

	// Delete member from organization
	err = r.Client.DeleteMemberFromOrg(ctx, orgName, userName)
	if err != nil {
		return &pbempty.Empty{}, fmt.Errorf("failed to delete member from organization: %w", err)
	}

	return &pbempty.Empty{}, nil
}

func (r *PulumiServiceOrgMemberResource) Diff(req *pulumirpc.DiffRequest) (*pulumirpc.DiffResponse, error) {
	// Unmarshal old and new properties
	oldInputMap, err := plugin.UnmarshalProperties(req.GetOlds(), plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	newInputMap, err := plugin.UnmarshalProperties(req.GetNews(), plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	oldInput := r.ToPulumiServiceOrgMemberInput(oldInputMap)
	newInput := r.ToPulumiServiceOrgMemberInput(newInputMap)

	// Check for changes
	var replaces []string
	hasChanges := false

	// organizationName is immutable - triggers replacement
	if oldInput.OrganizationName != newInput.OrganizationName {
		replaces = append(replaces, "organizationName")
		hasChanges = true
	}

	// userName is immutable - triggers replacement
	if oldInput.UserName != newInput.UserName {
		replaces = append(replaces, "userName")
		hasChanges = true
	}

	// role is mutable - triggers update only
	if oldInput.Role != newInput.Role {
		hasChanges = true
	}

	if !hasChanges {
		return &pulumirpc.DiffResponse{
			Changes: pulumirpc.DiffResponse_DIFF_NONE,
		}, nil
	}

	return &pulumirpc.DiffResponse{
		Changes:  pulumirpc.DiffResponse_DIFF_SOME,
		Replaces: replaces,
	}, nil
}
