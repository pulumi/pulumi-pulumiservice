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

package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
)

// invokeFunctionCreateTask implements the createTask function
func (k *pulumiserviceProvider) invokeFunctionCreateTask(ctx context.Context, req *pulumirpc.InvokeRequest) (*pulumirpc.InvokeResponse, error) {
	if k.client == nil {
		return nil, fmt.Errorf("provider not configured")
	}

	// Parse inputs
	inputs, err := plugin.UnmarshalProperties(req.GetArgs(), plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	// Extract required fields
	content := inputs["content"].StringValue()
	if content == "" {
		return nil, fmt.Errorf("content is required")
	}

	orgName := inputs["organizationName"].StringValue()
	if orgName == "" {
		return nil, fmt.Errorf("organizationName is required")
	}

	// Extract optional timestamp, default to now
	var timestamp time.Time
	if inputs["timestamp"].HasValue() && inputs["timestamp"].IsString() {
		timestampStr := inputs["timestamp"].StringValue()
		timestamp, err = time.Parse(time.RFC3339, timestampStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse timestamp: %w", err)
		}
	} else {
		timestamp = time.Now()
	}

	// Build entity diff from entityAdd and entityRemove
	var entityDiff *pulumiapi.EntityDiff
	if (inputs["entityAdd"].HasValue() && inputs["entityAdd"].IsArray()) ||
		(inputs["entityRemove"].HasValue() && inputs["entityRemove"].IsArray()) {
		entityDiff = &pulumiapi.EntityDiff{}

		// Process entityAdd
		if inputs["entityAdd"].HasValue() && inputs["entityAdd"].IsArray() {
			for _, e := range inputs["entityAdd"].ArrayValue() {
				if e.HasValue() && e.IsObject() {
					objMap := e.ObjectValue()
					entity := pulumiapi.TaskEntity{}
					if objMap["type"].HasValue() && objMap["type"].IsString() {
						entity.Type = objMap["type"].StringValue()
					}
					if objMap["id"].HasValue() && objMap["id"].IsString() {
						entity.ID = objMap["id"].StringValue()
					}
					entityDiff.Add = append(entityDiff.Add, entity)
				}
			}
		}

		// Process entityRemove
		if inputs["entityRemove"].HasValue() && inputs["entityRemove"].IsArray() {
			for _, e := range inputs["entityRemove"].ArrayValue() {
				if e.HasValue() && e.IsObject() {
					objMap := e.ObjectValue()
					entity := pulumiapi.TaskEntity{}
					if objMap["type"].HasValue() && objMap["type"].IsString() {
						entity.Type = objMap["type"].StringValue()
					}
					if objMap["id"].HasValue() && objMap["id"].IsString() {
						entity.ID = objMap["id"].StringValue()
					}
					entityDiff.Remove = append(entityDiff.Remove, entity)
				}
			}
		}
	}

	// Create the task via API
	createReq := pulumiapi.CreateTaskRequest{
		Content:    content,
		EntityDiff: entityDiff,
		Timestamp:  timestamp,
	}

	task, err := k.client.CreateTask(ctx, orgName, createReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	// Build output properties
	outputProps := resource.PropertyMap{
		"id":               resource.NewStringProperty(task.ID),
		"content":          resource.NewStringProperty(task.Name),
		"organizationName": resource.NewStringProperty(orgName),
		"timestamp":        resource.NewStringProperty(task.CreatedAt.Format(time.RFC3339)),
		"status":           resource.NewStringProperty(task.Status),
	}

	// Add entities to output (always include, even if empty)
	entities := make([]resource.PropertyValue, len(task.Entities))
	for idx, entity := range task.Entities {
		entityMap := resource.PropertyMap{
			"type": resource.NewStringProperty(entity.Type),
			"id":   resource.NewStringProperty(entity.ID),
		}
		entities[idx] = resource.NewObjectProperty(entityMap)
	}
	outputProps["entities"] = resource.NewArrayProperty(entities)

	outputProperties, err := plugin.MarshalProperties(outputProps, plugin.MarshalOptions{})
	if err != nil {
		return nil, err
	}

	return &pulumirpc.InvokeResponse{
		Return: outputProperties,
	}, nil
}
