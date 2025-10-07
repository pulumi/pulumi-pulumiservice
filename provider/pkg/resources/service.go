package resources

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	pbempty "google.golang.org/protobuf/types/known/emptypb"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
)

type PulumiServiceServiceResource struct {
	Client pulumiapi.ServiceClient
}

type PulumiServiceServiceInput struct {
	OrganizationName string
	OwnerType        string
	OwnerName        string
	Name             string
	Description      string
	Properties       map[string]string
	Items            []pulumiapi.ServiceItem
}

func (i *PulumiServiceServiceInput) ToPropertyMap() resource.PropertyMap {
	pm := resource.PropertyMap{}
	pm["organizationName"] = resource.NewPropertyValue(i.OrganizationName)
	pm["ownerType"] = resource.NewPropertyValue(i.OwnerType)
	pm["ownerName"] = resource.NewPropertyValue(i.OwnerName)
	pm["name"] = resource.NewPropertyValue(i.Name)

	if i.Description != "" {
		pm["description"] = resource.NewPropertyValue(i.Description)
	}

	if len(i.Properties) > 0 {
		propsMap := resource.PropertyMap{}
		for k, v := range i.Properties {
			propsMap[resource.PropertyKey(k)] = resource.NewPropertyValue(v)
		}
		pm["properties"] = resource.NewObjectProperty(propsMap)
	}

	if len(i.Items) > 0 {
		itemsArray := make([]resource.PropertyValue, len(i.Items))
		for idx, item := range i.Items {
			itemMap := resource.PropertyMap{
				"itemType": resource.NewPropertyValue(item.ItemType),
				"name":     resource.NewPropertyValue(item.Name),
			}
			itemsArray[idx] = resource.NewObjectProperty(itemMap)
		}
		pm["items"] = resource.NewArrayProperty(itemsArray)
	}

	return pm
}

func ToPulumiServiceServiceInput(inputMap resource.PropertyMap) PulumiServiceServiceInput {
	input := PulumiServiceServiceInput{}

	if inputMap["organizationName"].HasValue() && inputMap["organizationName"].IsString() {
		input.OrganizationName = inputMap["organizationName"].StringValue()
	}

	if inputMap["ownerType"].HasValue() && inputMap["ownerType"].IsString() {
		input.OwnerType = inputMap["ownerType"].StringValue()
	}

	if inputMap["ownerName"].HasValue() && inputMap["ownerName"].IsString() {
		input.OwnerName = inputMap["ownerName"].StringValue()
	}

	if inputMap["name"].HasValue() && inputMap["name"].IsString() {
		input.Name = inputMap["name"].StringValue()
	}

	if inputMap["description"].HasValue() && inputMap["description"].IsString() {
		input.Description = inputMap["description"].StringValue()
	}

	if inputMap["properties"].HasValue() && inputMap["properties"].IsObject() {
		input.Properties = make(map[string]string)
		for k, v := range inputMap["properties"].ObjectValue() {
			if v.HasValue() && v.IsString() {
				input.Properties[string(k)] = v.StringValue()
			}
		}
	}

	if inputMap["items"].HasValue() && inputMap["items"].IsArray() {
		itemsArray := inputMap["items"].ArrayValue()
		input.Items = make([]pulumiapi.ServiceItem, len(itemsArray))
		for idx, itemVal := range itemsArray {
			if itemVal.HasValue() && itemVal.IsObject() {
				itemObj := itemVal.ObjectValue()
				item := pulumiapi.ServiceItem{}
				if itemObj["itemType"].HasValue() && itemObj["itemType"].IsString() {
					item.ItemType = itemObj["itemType"].StringValue()
				}
				if itemObj["name"].HasValue() && itemObj["name"].IsString() {
					item.Name = itemObj["name"].StringValue()
				}
				input.Items[idx] = item
			}
		}
	}

	return input
}

func (s *PulumiServiceServiceResource) Name() string {
	return "pulumiservice:index:Service"
}

func (s *PulumiServiceServiceResource) Check(req *pulumirpc.CheckRequest) (*pulumirpc.CheckResponse, error) {
	news := req.GetNews()
	newsMap, err := plugin.UnmarshalProperties(news, plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	var failures []*pulumirpc.CheckFailure

	for _, p := range []resource.PropertyKey{"organizationName", "ownerType", "ownerName", "name"} {
		if !newsMap[p].HasValue() {
			failures = append(failures, &pulumirpc.CheckFailure{
				Reason:   fmt.Sprintf("missing required property '%s'", p),
				Property: string(p),
			})
		}
	}

	// Validate ownerType is either "user" or "team"
	if newsMap["ownerType"].HasValue() && newsMap["ownerType"].IsString() {
		ownerType := newsMap["ownerType"].StringValue()
		if ownerType != "user" && ownerType != "team" {
			failures = append(failures, &pulumirpc.CheckFailure{
				Reason:   fmt.Sprintf("ownerType must be either 'user' or 'team', got '%s'", ownerType),
				Property: "ownerType",
			})
		}
	}

	inputs, err := plugin.MarshalProperties(newsMap, plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	return &pulumirpc.CheckResponse{Inputs: inputs, Failures: failures}, nil
}

func (s *PulumiServiceServiceResource) Create(req *pulumirpc.CreateRequest) (*pulumirpc.CreateResponse, error) {
	ctx := context.Background()
	inputs, err := plugin.UnmarshalProperties(req.GetProperties(), plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	inputsService := ToPulumiServiceServiceInput(inputs)

	createReq := pulumiapi.CreateServiceRequest{
		OrganizationName: inputsService.OrganizationName,
		OwnerType:        inputsService.OwnerType,
		OwnerName:        inputsService.OwnerName,
		Name:             inputsService.Name,
		Description:      inputsService.Description,
		Properties:       inputsService.Properties,
		Items:            inputsService.Items,
	}

	service, err := s.Client.CreateService(ctx, createReq)
	if err != nil {
		return nil, fmt.Errorf("error creating service '%s': %w", inputsService.Name, err)
	}

	// Build the output from the created service
	outputs := PulumiServiceServiceInput{
		OrganizationName: inputsService.OrganizationName,
		OwnerType:        service.OwnerType,
		OwnerName:        service.OwnerName,
		Name:             service.Name,
		Description:      service.Description,
		Properties:       service.Properties,
		Items:            service.Items,
	}

	outputProperties, err := plugin.MarshalProperties(
		outputs.ToPropertyMap(),
		plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true},
	)
	if err != nil {
		return nil, err
	}

	serviceID := generateServiceID(inputsService.OrganizationName, service.OwnerType, service.OwnerName, service.Name)

	return &pulumirpc.CreateResponse{
		Id:         serviceID,
		Properties: outputProperties,
	}, nil
}

func (s *PulumiServiceServiceResource) Diff(req *pulumirpc.DiffRequest) (*pulumirpc.DiffResponse, error) {
	olds, err := plugin.UnmarshalProperties(req.GetOldInputs(), plugin.MarshalOptions{KeepUnknowns: false, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	news, err := plugin.UnmarshalProperties(req.GetNews(), plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	oldService := ToPulumiServiceServiceInput(olds)
	newService := ToPulumiServiceServiceInput(news)

	diffs := olds.Diff(news)
	if diffs == nil {
		return &pulumirpc.DiffResponse{
			Changes: pulumirpc.DiffResponse_DIFF_NONE,
		}, nil
	}

	dd := plugin.NewDetailedDiffFromObjectDiff(diffs, false)

	detailedDiffs := map[string]*pulumirpc.PropertyDiff{}
	replaceProperties := map[string]bool{
		"organizationName": true,
		"ownerType":        true,
		"ownerName":        true,
		"name":             true,
	}

	for k, v := range dd {
		if _, ok := replaceProperties[k]; ok {
			v.Kind = v.Kind.AsReplace()
		}
		detailedDiffs[k] = &pulumirpc.PropertyDiff{
			Kind:      pulumirpc.PropertyDiff_Kind(v.Kind),
			InputDiff: v.InputDiff,
		}
	}

	changes := pulumirpc.DiffResponse_DIFF_NONE
	if len(detailedDiffs) > 0 || !reflect.DeepEqual(oldService, newService) {
		changes = pulumirpc.DiffResponse_DIFF_SOME
	}

	return &pulumirpc.DiffResponse{
		Changes:         changes,
		DetailedDiff:    detailedDiffs,
		HasDetailedDiff: true,
	}, nil
}

func (s *PulumiServiceServiceResource) Update(req *pulumirpc.UpdateRequest) (*pulumirpc.UpdateResponse, error) {
	ctx := context.Background()

	inputsOld, err := plugin.UnmarshalProperties(req.GetOlds(), plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	inputsNew, err := plugin.UnmarshalProperties(req.GetNews(), plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	serviceOld := ToPulumiServiceServiceInput(inputsOld)
	serviceNew := ToPulumiServiceServiceInput(inputsNew)

	// Update description and properties if they changed
	if serviceOld.Description != serviceNew.Description || !reflect.DeepEqual(serviceOld.Properties, serviceNew.Properties) {
		updateReq := pulumiapi.UpdateServiceRequest{
			OrganizationName: serviceNew.OrganizationName,
			OwnerType:        serviceNew.OwnerType,
			OwnerName:        serviceNew.OwnerName,
			ServiceName:      serviceNew.Name,
			Properties:       serviceNew.Properties,
		}

		if serviceOld.Description != serviceNew.Description {
			updateReq.Description = &serviceNew.Description
		}

		_, err := s.Client.UpdateService(ctx, updateReq)
		if err != nil {
			return nil, fmt.Errorf("failed to update service: %w", err)
		}
	}

	// Reconcile items - remove old items that are not in the new list
	for _, oldItem := range serviceOld.Items {
		found := false
		for _, newItem := range serviceNew.Items {
			if oldItem.ItemType == newItem.ItemType && oldItem.Name == newItem.Name {
				found = true
				break
			}
		}
		if !found {
			removeReq := pulumiapi.RemoveServiceItemRequest{
				OrganizationName: serviceNew.OrganizationName,
				OwnerType:        serviceNew.OwnerType,
				OwnerName:        serviceNew.OwnerName,
				ServiceName:      serviceNew.Name,
				ItemType:         oldItem.ItemType,
				ItemName:         oldItem.Name,
			}
			if err := s.Client.RemoveServiceItem(ctx, removeReq); err != nil {
				return nil, fmt.Errorf("failed to remove service item: %w", err)
			}
		}
	}

	// Add new items that are not in the old list
	for _, newItem := range serviceNew.Items {
		found := false
		for _, oldItem := range serviceOld.Items {
			if oldItem.ItemType == newItem.ItemType && oldItem.Name == newItem.Name {
				found = true
				break
			}
		}
		if !found {
			addReq := pulumiapi.AddServiceItemRequest{
				OrganizationName: serviceNew.OrganizationName,
				OwnerType:        serviceNew.OwnerType,
				OwnerName:        serviceNew.OwnerName,
				ServiceName:      serviceNew.Name,
				Item:             newItem,
			}
			if err := s.Client.AddServiceItem(ctx, addReq); err != nil {
				return nil, fmt.Errorf("failed to add service item: %w", err)
			}
		}
	}

	outputProperties, err := plugin.MarshalProperties(
		serviceNew.ToPropertyMap(),
		plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true},
	)
	if err != nil {
		return nil, err
	}

	return &pulumirpc.UpdateResponse{
		Properties: outputProperties,
	}, nil
}

func (s *PulumiServiceServiceResource) Delete(req *pulumirpc.DeleteRequest) (*pbempty.Empty, error) {
	ctx := context.Background()

	orgName, ownerType, ownerName, serviceName, err := splitServiceID(req.Id)
	if err != nil {
		return nil, err
	}

	err = s.Client.DeleteService(ctx, orgName, ownerType, ownerName, serviceName, false)
	if err != nil {
		return nil, err
	}

	return &pbempty.Empty{}, nil
}

func (s *PulumiServiceServiceResource) Read(req *pulumirpc.ReadRequest) (*pulumirpc.ReadResponse, error) {
	ctx := context.Background()

	orgName, ownerType, ownerName, serviceName, err := splitServiceID(req.Id)
	if err != nil {
		return nil, err
	}

	service, err := s.Client.GetService(ctx, orgName, ownerType, ownerName, serviceName)
	if err != nil {
		return nil, fmt.Errorf("failed to read service (%q): %w", req.Id, err)
	}

	if service == nil {
		return &pulumirpc.ReadResponse{}, nil
	}

	outputs := PulumiServiceServiceInput{
		OrganizationName: orgName,
		OwnerType:        service.OwnerType,
		OwnerName:        service.OwnerName,
		Name:             service.Name,
		Description:      service.Description,
		Properties:       service.Properties,
		Items:            service.Items,
	}

	properties, err := plugin.MarshalProperties(outputs.ToPropertyMap(), plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal properties: %w", err)
	}

	return &pulumirpc.ReadResponse{
		Id:         req.Id,
		Properties: properties,
		Inputs:     properties,
	}, nil
}

func generateServiceID(orgName, ownerType, ownerName, serviceName string) string {
	return fmt.Sprintf("%s/%s/%s/%s", orgName, ownerType, ownerName, serviceName)
}

func splitServiceID(id string) (string, string, string, string, error) {
	// format: organization/ownerType/ownerName/serviceName
	parts := strings.Split(id, "/")
	if len(parts) != 4 {
		return "", "", "", "", fmt.Errorf("%q is not a valid service ID (expected format: org/ownerType/ownerName/serviceName)", id)
	}
	return parts[0], parts[1], parts[2], parts[3], nil
}
