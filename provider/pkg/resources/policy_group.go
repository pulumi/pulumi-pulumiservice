// Copyright 2016-2025, Pulumi Corporation.
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
	"reflect"

	"golang.org/x/exp/slices"
	"google.golang.org/grpc/codes"
	pbempty "google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	"github.com/pulumi/pulumi/sdk/v3/go/common/util/rpcutil/rpcerror"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
)

type PulumiServicePolicyGroupResource struct {
	Client pulumiapi.PolicyGroupClient
}

type PulumiServicePolicyGroupInput struct {
	Name             string
	OrganizationName string
	EntityType       string
	Mode             string
	Stacks           []pulumiapi.StackReference
	PolicyPacks      []pulumiapi.PolicyPackMetadata
}

func (i *PulumiServicePolicyGroupInput) ToPropertyMap() resource.PropertyMap {
	// Convert the entire struct to a map first, then use helper
	inputMap := map[string]interface{}{
		"name":             i.Name,
		"organizationName": i.OrganizationName,
		"entityType":       i.EntityType,
		"mode":             i.Mode,
		"stacks":           convertStacksToInterfaceArray(i.Stacks),
		"policyPacks":      convertPolicyPacksToInterfaceArray(i.PolicyPacks),
	}

	return resource.NewPropertyMapFromMap(inputMap)
}

// convertStacksToInterfaceArray converts []pulumiapi.StackReference to []interface{}
func convertStacksToInterfaceArray(stacks []pulumiapi.StackReference) []interface{} {
	result := make([]interface{}, len(stacks))
	for i, stack := range stacks {
		result[i] = map[string]interface{}{
			"name":           stack.Name,
			"routingProject": stack.RoutingProject,
		}
	}
	return result
}

// convertPolicyPacksToInterfaceArray converts []pulumiapi.PolicyPackMetadata to []interface{}
func convertPolicyPacksToInterfaceArray(policyPacks []pulumiapi.PolicyPackMetadata) []interface{} {
	result := make([]interface{}, len(policyPacks))
	for i, pp := range policyPacks {
		ppMap := map[string]interface{}{
			"name":        pp.Name,
			"displayName": pp.DisplayName,
			"version":     float64(pp.Version),
			"versionTag":  pp.VersionTag,
		}
		if pp.Config != nil {
			ppMap["config"] = pp.Config
		}
		result[i] = ppMap
	}
	return result
}

// convertInterfaceArrayToStacks converts []interface{} to []pulumiapi.StackReference using helpers
func convertInterfaceArrayToStacks(arr []interface{}) []pulumiapi.StackReference {
	result := make([]pulumiapi.StackReference, 0, len(arr))
	for _, item := range arr {
		if stackMap, ok := item.(map[string]interface{}); ok {
			stack := pulumiapi.StackReference{}
			if name, ok := stackMap["name"].(string); ok {
				stack.Name = name
			}
			if routingProject, ok := stackMap["routingProject"].(string); ok {
				stack.RoutingProject = routingProject
			}
			result = append(result, stack)
		}
	}
	return result
}

// convertInterfaceArrayToPolicyPacks converts []interface{} to []pulumiapi.PolicyPackMetadata using helpers
func convertInterfaceArrayToPolicyPacks(arr []interface{}) []pulumiapi.PolicyPackMetadata {
	result := make([]pulumiapi.PolicyPackMetadata, 0, len(arr))
	for _, item := range arr {
		if ppMap, ok := item.(map[string]interface{}); ok {
			pp := pulumiapi.PolicyPackMetadata{}
			if name, ok := ppMap["name"].(string); ok {
				pp.Name = name
			}
			if displayName, ok := ppMap["displayName"].(string); ok {
				pp.DisplayName = displayName
			}
			if version, ok := ppMap["version"].(float64); ok {
				pp.Version = int(version)
			}
			if versionTag, ok := ppMap["versionTag"].(string); ok {
				pp.VersionTag = versionTag
			}
			if config, ok := ppMap["config"].(map[string]interface{}); ok {
				pp.Config = config
			}
			result = append(result, pp)
		}
	}
	return result
}

func (i *PulumiServicePolicyGroupInput) ToRpc() (*structpb.Struct, error) {
	return plugin.MarshalProperties(i.ToPropertyMap(), plugin.MarshalOptions{
		KeepOutputValues: true,
	})
}

func ToPulumiServicePolicyGroupInput(inputMap resource.PropertyMap) PulumiServicePolicyGroupInput {
	// Convert PropertyMap to regular map using helper, then extract fields
	interfaceMap := inputMap.Mappable()

	input := PulumiServicePolicyGroupInput{}

	if name, ok := interfaceMap["name"].(string); ok {
		input.Name = name
	}

	if organizationName, ok := interfaceMap["organizationName"].(string); ok {
		input.OrganizationName = organizationName
	}

	if entityType, ok := interfaceMap["entityType"].(string); ok {
		input.EntityType = entityType
	}

	if mode, ok := interfaceMap["mode"].(string); ok {
		input.Mode = mode
	}

	// Parse stacks using helper
	if stacksInterface, ok := interfaceMap["stacks"].([]interface{}); ok {
		input.Stacks = convertInterfaceArrayToStacks(stacksInterface)
	}

	// Parse policy packs using helper
	if policyPacksInterface, ok := interfaceMap["policyPacks"].([]interface{}); ok {
		input.PolicyPacks = convertInterfaceArrayToPolicyPacks(policyPacksInterface)
	}

	return input
}

func (p *PulumiServicePolicyGroupResource) Name() string {
	return "pulumiservice:index:PolicyGroup"
}

func (p *PulumiServicePolicyGroupResource) Check(req *pulumirpc.CheckRequest) (*pulumirpc.CheckResponse, error) {
	news := req.GetNews()
	newsMap, err := plugin.UnmarshalProperties(news, plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	var failures []*pulumirpc.CheckFailure

	if !newsMap["name"].HasValue() {
		failures = append(failures, &pulumirpc.CheckFailure{
			Reason:   "missing required property 'name'",
			Property: "name",
		})
	}

	if !newsMap["organizationName"].HasValue() {
		failures = append(failures, &pulumirpc.CheckFailure{
			Reason:   "missing required property 'organizationName'",
			Property: "organizationName",
		})
	}

	// Apply defaults if not provided
	if !newsMap["entityType"].HasValue() {
		newsMap["entityType"] = resource.NewPropertyValue("stacks")
	}
	if !newsMap["mode"].HasValue() {
		newsMap["mode"] = resource.NewPropertyValue("audit")
	}

	// Validate enum values
	if newsMap["entityType"].HasValue() {
		entityType := newsMap["entityType"].StringValue()
		if entityType != "stacks" && entityType != "accounts" {
			failures = append(failures, &pulumirpc.CheckFailure{
				Reason:   "entityType must be either 'stacks' or 'accounts'",
				Property: "entityType",
			})
		}
	}

	if newsMap["mode"].HasValue() {
		mode := newsMap["mode"].StringValue()
		if mode != "audit" && mode != "preventative" {
			failures = append(failures, &pulumirpc.CheckFailure{
				Reason:   "mode must be either 'audit' or 'preventative'",
				Property: "mode",
			})
		}
	}

	inputs, err := plugin.MarshalProperties(newsMap, plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	return &pulumirpc.CheckResponse{Inputs: inputs, Failures: failures}, nil
}

func (p *PulumiServicePolicyGroupResource) Delete(req *pulumirpc.DeleteRequest) (*pbempty.Empty, error) {
	ctx := context.Background()
	orgName, policyGroupName, err := splitSingleSlashString(req.Id)
	if err != nil {
		return &pbempty.Empty{}, err
	}

	err = p.Client.DeletePolicyGroup(ctx, orgName, policyGroupName)
	if err != nil {
		return &pbempty.Empty{}, fmt.Errorf("failed to delete policy group %q: %w", req.Id, err)
	}

	return &pbempty.Empty{}, nil
}

func (p *PulumiServicePolicyGroupResource) Diff(req *pulumirpc.DiffRequest) (*pulumirpc.DiffResponse, error) {
	olds, err := plugin.UnmarshalProperties(req.GetOldInputs(), plugin.MarshalOptions{KeepUnknowns: false, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	news, err := plugin.UnmarshalProperties(req.GetNews(), plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: false})
	if err != nil {
		return nil, err
	}

	oldPolicyGroup := ToPulumiServicePolicyGroupInput(olds)
	newPolicyGroup := ToPulumiServicePolicyGroupInput(news)

	changes := pulumirpc.DiffResponse_DIFF_NONE

	if !reflect.DeepEqual(oldPolicyGroup, newPolicyGroup) {
		changes = pulumirpc.DiffResponse_DIFF_SOME
	}

	// Name change requires replacement
	replaces := []string{}
	if oldPolicyGroup.Name != newPolicyGroup.Name {
		replaces = append(replaces, "name")
	}

	// EntityType change requires replacement
	if oldPolicyGroup.EntityType != newPolicyGroup.EntityType {
		replaces = append(replaces, "entityType")
	}

	// Mode change requires replacement (API doesn't support update yet)
	if oldPolicyGroup.Mode != newPolicyGroup.Mode {
		replaces = append(replaces, "mode")
	}

	return &pulumirpc.DiffResponse{
		Changes:             changes,
		Replaces:            replaces,
		Stables:             []string{},
		DeleteBeforeReplace: true,
	}, nil
}

func (p *PulumiServicePolicyGroupResource) Read(req *pulumirpc.ReadRequest) (*pulumirpc.ReadResponse, error) {
	ctx := context.Background()

	orgName, policyGroupName, err := splitSingleSlashString(req.Id)
	if err != nil {
		return nil, err
	}

	policyGroup, err := p.Client.GetPolicyGroup(ctx, orgName, policyGroupName)
	if err != nil {
		return nil, fmt.Errorf("failed to read policy group (%q): %w", req.Id, err)
	}
	if policyGroup == nil {
		return &pulumirpc.ReadResponse{}, nil
	}

	// Convert API response directly to PropertyMap using consistent helpers
	inputMap := map[string]interface{}{
		"name":             policyGroup.Name,
		"organizationName": orgName,
		"entityType":       policyGroup.EntityType,
		"mode":             policyGroup.Mode,
		"stacks":           convertStacksToInterfaceArray(policyGroup.Stacks),
		"policyPacks":      convertPolicyPacksToInterfaceArray(policyGroup.AppliedPolicyPacks),
	}

	propertyMap := resource.NewPropertyMapFromMap(inputMap)

	props, err := plugin.MarshalProperties(propertyMap, plugin.MarshalOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal inputs to properties: %w", err)
	}

	return &pulumirpc.ReadResponse{
		Id:         req.Id,
		Properties: props,
		Inputs:     props,
	}, nil
}

func (p *PulumiServicePolicyGroupResource) Update(req *pulumirpc.UpdateRequest) (*pulumirpc.UpdateResponse, error) {
	ctx := context.Background()
	inputsOld, err := plugin.UnmarshalProperties(req.GetOlds(), plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
	if err != nil {
		return nil, err
	}
	inputsNew, err := plugin.UnmarshalProperties(req.GetNews(), plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	policyGroupOld := ToPulumiServicePolicyGroupInput(inputsOld)
	policyGroupNew := ToPulumiServicePolicyGroupInput(inputsNew)

	// Handle stack changes
	if !stackReferencesEqual(policyGroupOld.Stacks, policyGroupNew.Stacks) {
		// Remove stacks that are no longer in the new list
		for _, oldStack := range policyGroupOld.Stacks {
			if !containsStackReference(policyGroupNew.Stacks, oldStack) {
				updateReq := pulumiapi.UpdatePolicyGroupRequest{
					RemoveStack: &oldStack,
				}
				err := p.Client.UpdatePolicyGroup(ctx, policyGroupNew.OrganizationName, policyGroupNew.Name, updateReq)
				if err != nil {
					return nil, fmt.Errorf("failed to remove stack from policy group: %w", err)
				}
			}
		}

		// Add stacks that are new in the new list
		for _, newStack := range policyGroupNew.Stacks {
			if !containsStackReference(policyGroupOld.Stacks, newStack) {
				updateReq := pulumiapi.UpdatePolicyGroupRequest{
					AddStack: &newStack,
				}
				err := p.Client.UpdatePolicyGroup(ctx, policyGroupNew.OrganizationName, policyGroupNew.Name, updateReq)
				if err != nil {
					return nil, fmt.Errorf("failed to add stack to policy group: %w", err)
				}
			}
		}
	}

	// Handle policy pack changes
	if !policyPacksEqual(policyGroupOld.PolicyPacks, policyGroupNew.PolicyPacks) {
		// Remove policy packs that are no longer in the new list
		for _, oldPP := range policyGroupOld.PolicyPacks {
			if !containsPolicyPack(policyGroupNew.PolicyPacks, oldPP) {
				updateReq := pulumiapi.UpdatePolicyGroupRequest{
					RemovePolicyPack: &oldPP,
				}
				err := p.Client.UpdatePolicyGroup(ctx, policyGroupNew.OrganizationName, policyGroupNew.Name, updateReq)
				if err != nil {
					return nil, fmt.Errorf("failed to remove policy pack from policy group: %w", err)
				}
			}
		}

		// Add policy packs that are new in the new list
		for _, newPP := range policyGroupNew.PolicyPacks {
			if !containsPolicyPack(policyGroupOld.PolicyPacks, newPP) {
				updateReq := pulumiapi.UpdatePolicyGroupRequest{
					AddPolicyPack: &newPP,
				}
				err := p.Client.UpdatePolicyGroup(ctx, policyGroupNew.OrganizationName, policyGroupNew.Name, updateReq)
				if err != nil {
					return nil, fmt.Errorf("failed to add policy pack to policy group: %w", err)
				}
			}
		}
	}

	outputProperties, err := policyGroupNew.ToRpc()
	if err != nil {
		return nil, err
	}

	return &pulumirpc.UpdateResponse{
		Properties: outputProperties,
	}, nil
}

func (p *PulumiServicePolicyGroupResource) Create(req *pulumirpc.CreateRequest) (*pulumirpc.CreateResponse, error) {
	ctx := context.Background()
	inputs, err := plugin.UnmarshalProperties(req.GetProperties(), plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	inputsPolicyGroup := ToPulumiServicePolicyGroupInput(inputs)

	// Create the policy group
	err = p.Client.CreatePolicyGroup(ctx, inputsPolicyGroup.OrganizationName, inputsPolicyGroup.Name, inputsPolicyGroup.EntityType, inputsPolicyGroup.Mode)
	if err != nil {
		return nil, fmt.Errorf("error creating policy group '%s': %w", inputsPolicyGroup.Name, err)
	}

	policyGroupID := fmt.Sprintf("%s/%s", inputsPolicyGroup.OrganizationName, inputsPolicyGroup.Name)

	// Add stacks to the policy group
	for _, stack := range inputsPolicyGroup.Stacks {
		updateReq := pulumiapi.UpdatePolicyGroupRequest{
			AddStack: &stack,
		}
		err := p.Client.UpdatePolicyGroup(ctx, inputsPolicyGroup.OrganizationName, inputsPolicyGroup.Name, updateReq)
		if err != nil {
			return nil, partialErrorPolicyGroup(policyGroupID, fmt.Errorf("failed to add stack to policy group: %w", err), inputsPolicyGroup, inputsPolicyGroup)
		}
	}

	// Add policy packs to the policy group
	for _, pp := range inputsPolicyGroup.PolicyPacks {
		updateReq := pulumiapi.UpdatePolicyGroupRequest{
			AddPolicyPack: &pp,
		}
		err := p.Client.UpdatePolicyGroup(ctx, inputsPolicyGroup.OrganizationName, inputsPolicyGroup.Name, updateReq)
		if err != nil {
			return nil, partialErrorPolicyGroup(policyGroupID, fmt.Errorf("failed to add policy pack to policy group: %w", err), inputsPolicyGroup, inputsPolicyGroup)
		}
	}

	// Read back the policy group to get the full state
	policyGroup, err := p.Client.GetPolicyGroup(ctx, inputsPolicyGroup.OrganizationName, inputsPolicyGroup.Name)
	if err != nil {
		return nil, partialErrorPolicyGroup(policyGroupID, err, inputsPolicyGroup, inputsPolicyGroup)
	}

	outputs := PulumiServicePolicyGroupInput{
		Name:             policyGroup.Name,
		OrganizationName: inputsPolicyGroup.OrganizationName,
		Stacks:           policyGroup.Stacks,
		PolicyPacks:      policyGroup.AppliedPolicyPacks,
		EntityType:       policyGroup.EntityType,
		Mode:             policyGroup.Mode,
	}

	outputProperties, err := outputs.ToRpc()
	if err != nil {
		return nil, partialErrorPolicyGroup(policyGroupID, err, outputs, inputsPolicyGroup)
	}

	return &pulumirpc.CreateResponse{
		Id:         policyGroupID,
		Properties: outputProperties,
	}, nil
}

// Helper functions

func stackReferencesEqual(a, b []pulumiapi.StackReference) bool {
	if len(a) != len(b) {
		return false
	}
	aCopy := make([]pulumiapi.StackReference, len(a))
	bCopy := make([]pulumiapi.StackReference, len(b))
	copy(aCopy, a)
	copy(bCopy, b)

	slices.SortFunc(aCopy, func(i, j pulumiapi.StackReference) int {
		if i.RoutingProject != j.RoutingProject {
			if i.RoutingProject < j.RoutingProject {
				return -1
			}
			return 1
		}
		if i.Name < j.Name {
			return -1
		}
		if i.Name > j.Name {
			return 1
		}
		return 0
	})
	slices.SortFunc(bCopy, func(i, j pulumiapi.StackReference) int {
		if i.RoutingProject != j.RoutingProject {
			if i.RoutingProject < j.RoutingProject {
				return -1
			}
			return 1
		}
		if i.Name < j.Name {
			return -1
		}
		if i.Name > j.Name {
			return 1
		}
		return 0
	})

	return slices.EqualFunc(aCopy, bCopy, func(x, y pulumiapi.StackReference) bool {
		return x.Name == y.Name && x.RoutingProject == y.RoutingProject
	})
}

func containsStackReference(stacks []pulumiapi.StackReference, target pulumiapi.StackReference) bool {
	for _, stack := range stacks {
		if stack.Name == target.Name && stack.RoutingProject == target.RoutingProject {
			return true
		}
	}
	return false
}

func policyPacksEqual(a, b []pulumiapi.PolicyPackMetadata) bool {
	if len(a) != len(b) {
		return false
	}
	aCopy := make([]pulumiapi.PolicyPackMetadata, len(a))
	bCopy := make([]pulumiapi.PolicyPackMetadata, len(b))
	copy(aCopy, a)
	copy(bCopy, b)

	slices.SortFunc(aCopy, func(i, j pulumiapi.PolicyPackMetadata) int {
		if i.Name < j.Name {
			return -1
		}
		if i.Name > j.Name {
			return 1
		}
		if i.Version < j.Version {
			return -1
		}
		if i.Version > j.Version {
			return 1
		}
		return 0
	})
	slices.SortFunc(bCopy, func(i, j pulumiapi.PolicyPackMetadata) int {
		if i.Name < j.Name {
			return -1
		}
		if i.Name > j.Name {
			return 1
		}
		if i.Version < j.Version {
			return -1
		}
		if i.Version > j.Version {
			return 1
		}
		return 0
	})

	return slices.EqualFunc(aCopy, bCopy, func(x, y pulumiapi.PolicyPackMetadata) bool {
		return x.Name == y.Name && x.Version == y.Version && x.VersionTag == y.VersionTag
	})
}

func containsPolicyPack(packs []pulumiapi.PolicyPackMetadata, target pulumiapi.PolicyPackMetadata) bool {
	for _, pp := range packs {
		if pp.Name == target.Name && pp.Version == target.Version {
			return true
		}
	}
	return false
}

// partialErrorPolicyGroup creates an error for resources that did not complete an operation in progress.
// The last known state of the object is included in the error so that it can be checkpointed.
func partialErrorPolicyGroup(id string, err error, state PulumiServicePolicyGroupInput, inputs PulumiServicePolicyGroupInput) error {
	stateRpc, stateSerErr := state.ToRpc()
	inputRpc, inputSerErr := inputs.ToRpc()

	// combine errors if we can't serialize state or inputs for some reason
	if stateSerErr != nil {
		err = fmt.Errorf("err serializing state: %v, (src error: %v)", stateSerErr, err)
	}
	if inputSerErr != nil {
		err = fmt.Errorf("err serializing inputs: %v (src error: %v)", inputSerErr, err)
	}
	detail := pulumirpc.ErrorResourceInitFailed{
		Id:         id,
		Properties: stateRpc,
		Reasons:    []string{err.Error()},
		Inputs:     inputRpc,
	}
	return rpcerror.WithDetails(rpcerror.New(codes.Unknown, err.Error()), &detail)
}
