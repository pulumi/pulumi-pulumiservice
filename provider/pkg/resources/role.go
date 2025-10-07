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
	"strings"

	pbempty "google.golang.org/protobuf/types/known/emptypb"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/util"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
)

type PulumiServiceRoleResource struct {
	Client pulumiapi.RoleClient
}

type PulumiServiceRole struct {
	OrganizationName string                 `json:"organizationName"`
	Name             string                 `json:"name"`
	Description      string                 `json:"description"`
	UXPurpose        string                 `json:"uxPurpose"`
	Details          map[string]interface{} `json:"details"`
}

func (i *PulumiServiceRole) ToPropertyMap() resource.PropertyMap {
	pm := resource.PropertyMap{}
	pm["organizationName"] = resource.NewPropertyValue(i.OrganizationName)
	pm["name"] = resource.NewPropertyValue(i.Name)
	pm["description"] = resource.NewPropertyValue(i.Description)

	if i.UXPurpose != "" {
		pm["uxPurpose"] = resource.NewPropertyValue(i.UXPurpose)
	}

	// Convert details map to PropertyValue
	if i.Details != nil {
		detailsValue := mapToPropertyValue(i.Details)
		pm["details"] = detailsValue
	}

	return pm
}

func (s *PulumiServiceRoleResource) ToPulumiServiceRoleInput(inputMap resource.PropertyMap) (*PulumiServiceRole, error) {
	role := PulumiServiceRole{}

	if inputMap["organizationName"].HasValue() && inputMap["organizationName"].IsString() {
		role.OrganizationName = inputMap["organizationName"].StringValue()
	}

	if inputMap["name"].HasValue() && inputMap["name"].IsString() {
		role.Name = inputMap["name"].StringValue()
	}

	if inputMap["description"].HasValue() && inputMap["description"].IsString() {
		role.Description = inputMap["description"].StringValue()
	}

	if inputMap["uxPurpose"].HasValue() && inputMap["uxPurpose"].IsString() {
		role.UXPurpose = inputMap["uxPurpose"].StringValue()
	}

	// Parse details
	if inputMap["details"].HasValue() && inputMap["details"].IsObject() {
		role.Details = propertyValueToMap(inputMap["details"])
	}

	return &role, nil
}

func (s *PulumiServiceRoleResource) Name() string {
	return "pulumiservice:index:Role"
}

func buildRoleID(orgName string, roleID string) string {
	return fmt.Sprintf("%s/%s", orgName, roleID)
}

func parseRoleID(compositeID string) (string, string, error) {
	parts := strings.Split(compositeID, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid role ID format: expected '{orgName}/{roleID}', got %q", compositeID)
	}

	return parts[0], parts[1], nil
}

func (s *PulumiServiceRoleResource) Diff(req *pulumirpc.DiffRequest) (*pulumirpc.DiffResponse, error) {
	replaceProperties := map[string]bool{
		"organizationName": true,
	}
	return util.StandardDiff(req, replaceProperties)
}

func (s *PulumiServiceRoleResource) Delete(req *pulumirpc.DeleteRequest) (*pbempty.Empty, error) {
	ctx := context.Background()

	orgName, roleID, err := parseRoleID(req.GetId())
	if err != nil {
		return nil, err
	}

	err = s.Client.DeleteRole(ctx, orgName, roleID)
	if err != nil {
		return nil, err
	}

	return &pbempty.Empty{}, nil
}

func (s *PulumiServiceRoleResource) Create(req *pulumirpc.CreateRequest) (*pulumirpc.CreateResponse, error) {
	ctx := context.Background()
	inputs, err := plugin.UnmarshalProperties(req.GetProperties(), util.StandardUnmarshal)
	if err != nil {
		return nil, err
	}

	role, err := s.ToPulumiServiceRoleInput(inputs)
	if err != nil {
		return nil, err
	}

	createReq := pulumiapi.CreateRoleRequest{
		Name:        role.Name,
		Description: role.Description,
		UXPurpose:   role.UXPurpose,
		Details:     role.Details,
	}

	// Set default uxPurpose if not provided
	if createReq.UXPurpose == "" {
		createReq.UXPurpose = "role"
	}

	createdRole, err := s.Client.CreateRole(ctx, role.OrganizationName, createReq)
	if err != nil {
		return nil, err
	}

	outputProperties, err := plugin.MarshalProperties(
		role.ToPropertyMap(),
		util.StandardMarshal,
	)
	if err != nil {
		return nil, err
	}

	return &pulumirpc.CreateResponse{
		Id:         buildRoleID(role.OrganizationName, createdRole.ID),
		Properties: outputProperties,
	}, nil
}

func (s *PulumiServiceRoleResource) Check(req *pulumirpc.CheckRequest) (*pulumirpc.CheckResponse, error) {
	inputs, err := plugin.UnmarshalProperties(req.GetNews(), util.StandardUnmarshal)
	if err != nil {
		return nil, err
	}

	var failures []*pulumirpc.CheckFailure

	// Validate required fields
	if !inputs["organizationName"].HasValue() || !inputs["organizationName"].IsString() || inputs["organizationName"].StringValue() == "" {
		failures = append(failures, &pulumirpc.CheckFailure{
			Property: "organizationName",
			Reason:   "organizationName is required",
		})
	}

	if !inputs["name"].HasValue() || !inputs["name"].IsString() || inputs["name"].StringValue() == "" {
		failures = append(failures, &pulumirpc.CheckFailure{
			Property: "name",
			Reason:   "name is required",
		})
	}

	if !inputs["description"].HasValue() || !inputs["description"].IsString() || inputs["description"].StringValue() == "" {
		failures = append(failures, &pulumirpc.CheckFailure{
			Property: "description",
			Reason:   "description is required",
		})
	}

	if !inputs["details"].HasValue() || !inputs["details"].IsObject() {
		failures = append(failures, &pulumirpc.CheckFailure{
			Property: "details",
			Reason:   "details is required and must be an object",
		})
	}

	return &pulumirpc.CheckResponse{Inputs: req.News, Failures: failures}, nil
}

func (s *PulumiServiceRoleResource) Update(req *pulumirpc.UpdateRequest) (*pulumirpc.UpdateResponse, error) {
	ctx := context.Background()
	inputs, err := plugin.UnmarshalProperties(req.GetNews(), util.StandardUnmarshal)
	if err != nil {
		return nil, err
	}

	role, err := s.ToPulumiServiceRoleInput(inputs)
	if err != nil {
		return nil, err
	}

	orgName, roleID, err := parseRoleID(req.GetId())
	if err != nil {
		return nil, err
	}

	updateReq := pulumiapi.UpdateRoleRequest{
		Name:        role.Name,
		Description: role.Description,
		Details:     role.Details,
	}

	_, err = s.Client.UpdateRole(ctx, orgName, roleID, updateReq)
	if err != nil {
		return nil, err
	}

	outputProperties, err := plugin.MarshalProperties(
		role.ToPropertyMap(),
		util.StandardMarshal,
	)
	if err != nil {
		return nil, err
	}

	return &pulumirpc.UpdateResponse{
		Properties: outputProperties,
	}, nil
}

func (s *PulumiServiceRoleResource) Read(req *pulumirpc.ReadRequest) (*pulumirpc.ReadResponse, error) {
	ctx := context.Background()

	orgName, roleID, err := parseRoleID(req.GetId())
	if err != nil {
		return nil, err
	}

	apiRole, err := s.Client.GetRole(ctx, orgName, roleID)
	if err != nil {
		return nil, fmt.Errorf("failure while reading role %q: %w", req.Id, err)
	}
	if apiRole == nil {
		return &pulumirpc.ReadResponse{}, nil
	}

	// Convert API response back to our input format
	readRole := PulumiServiceRole{
		OrganizationName: orgName,
		Name:             apiRole.Name,
		Description:      apiRole.Description,
		UXPurpose:        apiRole.UXPurpose,
		Details:          apiRole.Details,
	}

	outputs, err := plugin.MarshalProperties(
		readRole.ToPropertyMap(),
		util.StandardMarshal,
	)
	if err != nil {
		return nil, err
	}

	return &pulumirpc.ReadResponse{
		Id:         buildRoleID(orgName, apiRole.ID),
		Properties: outputs,
		Inputs:     outputs,
	}, nil
}

// mapToPropertyValue converts a map[string]interface{} to a resource.PropertyValue
func mapToPropertyValue(m map[string]interface{}) resource.PropertyValue {
	pm := resource.PropertyMap{}
	for k, v := range m {
		pm[resource.PropertyKey(k)] = interfaceToPropertyValue(v)
	}
	return resource.NewPropertyValue(pm)
}

// propertyValueToMap converts a resource.PropertyValue to a map[string]interface{}
func propertyValueToMap(pv resource.PropertyValue) map[string]interface{} {
	if !pv.IsObject() {
		return nil
	}

	result := make(map[string]interface{})
	for k, v := range pv.ObjectValue() {
		result[string(k)] = propertyValueToInterface(v)
	}
	return result
}

// interfaceToPropertyValue converts an interface{} to a resource.PropertyValue
func interfaceToPropertyValue(v interface{}) resource.PropertyValue {
	switch val := v.(type) {
	case bool:
		return resource.NewPropertyValue(val)
	case float64:
		return resource.NewPropertyValue(val)
	case string:
		return resource.NewPropertyValue(val)
	case []interface{}:
		arr := []resource.PropertyValue{}
		for _, item := range val {
			arr = append(arr, interfaceToPropertyValue(item))
		}
		return resource.NewPropertyValue(arr)
	case map[string]interface{}:
		return mapToPropertyValue(val)
	case nil:
		return resource.NewNullProperty()
	default:
		// For unknown types, try to convert to string
		return resource.NewPropertyValue(fmt.Sprintf("%v", val))
	}
}

// propertyValueToInterface converts a resource.PropertyValue to an interface{}
func propertyValueToInterface(pv resource.PropertyValue) interface{} {
	if pv.IsBool() {
		return pv.BoolValue()
	}
	if pv.IsNumber() {
		return pv.NumberValue()
	}
	if pv.IsString() {
		return pv.StringValue()
	}
	if pv.IsArray() {
		arr := []interface{}{}
		for _, item := range pv.ArrayValue() {
			arr = append(arr, propertyValueToInterface(item))
		}
		return arr
	}
	if pv.IsObject() {
		return propertyValueToMap(pv)
	}
	if pv.IsNull() {
		return nil
	}
	// For unknown types, return empty string
	return ""
}
