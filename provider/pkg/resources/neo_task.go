package resources

import (
	"context"
	"fmt"
	"time"

	pbempty "google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
)

type PulumiServiceNeoTaskResource struct {
	Client pulumiapi.TaskClient
}

type PulumiServiceNeoTaskInput struct {
	Content          string
	OrganizationName string
	EntityAdd        []PulumiServiceTaskEntity
	EntityRemove     []PulumiServiceTaskEntity
	Timestamp        string
}

type PulumiServiceTaskEntity struct {
	Type string
	ID   string
}

func (i *PulumiServiceNeoTaskInput) ToPropertyMap() resource.PropertyMap {
	pm := resource.PropertyMap{}
	pm["content"] = resource.NewPropertyValue(i.Content)
	pm["organizationName"] = resource.NewPropertyValue(i.OrganizationName)

	if len(i.EntityAdd) > 0 {
		entities := make([]resource.PropertyValue, len(i.EntityAdd))
		for idx, entity := range i.EntityAdd {
			entityMap := resource.PropertyMap{
				"type": resource.NewPropertyValue(entity.Type),
				"id":   resource.NewPropertyValue(entity.ID),
			}
			entities[idx] = resource.NewObjectProperty(entityMap)
		}
		pm["entityAdd"] = resource.NewArrayProperty(entities)
	}

	if len(i.EntityRemove) > 0 {
		entities := make([]resource.PropertyValue, len(i.EntityRemove))
		for idx, entity := range i.EntityRemove {
			entityMap := resource.PropertyMap{
				"type": resource.NewPropertyValue(entity.Type),
				"id":   resource.NewPropertyValue(entity.ID),
			}
			entities[idx] = resource.NewObjectProperty(entityMap)
		}
		pm["entityRemove"] = resource.NewArrayProperty(entities)
	}

	if i.Timestamp != "" {
		pm["timestamp"] = resource.NewPropertyValue(i.Timestamp)
	}

	return pm
}

func (i *PulumiServiceNeoTaskInput) ToRpc() (*structpb.Struct, error) {
	return plugin.MarshalProperties(i.ToPropertyMap(), plugin.MarshalOptions{
		KeepOutputValues: true,
	})
}

func ToPulumiServiceNeoTaskInput(inputMap resource.PropertyMap) PulumiServiceNeoTaskInput {
	input := PulumiServiceNeoTaskInput{}

	if inputMap["content"].HasValue() && inputMap["content"].IsString() {
		input.Content = inputMap["content"].StringValue()
	}

	if inputMap["organizationName"].HasValue() && inputMap["organizationName"].IsString() {
		input.OrganizationName = inputMap["organizationName"].StringValue()
	}

	if inputMap["entityAdd"].HasValue() && inputMap["entityAdd"].IsArray() {
		for _, e := range inputMap["entityAdd"].ArrayValue() {
			if e.HasValue() && e.IsObject() {
				objMap := e.ObjectValue()
				entity := PulumiServiceTaskEntity{}
				if objMap["type"].HasValue() && objMap["type"].IsString() {
					entity.Type = objMap["type"].StringValue()
				}
				if objMap["id"].HasValue() && objMap["id"].IsString() {
					entity.ID = objMap["id"].StringValue()
				}
				input.EntityAdd = append(input.EntityAdd, entity)
			}
		}
	}

	if inputMap["entityRemove"].HasValue() && inputMap["entityRemove"].IsArray() {
		for _, e := range inputMap["entityRemove"].ArrayValue() {
			if e.HasValue() && e.IsObject() {
				objMap := e.ObjectValue()
				entity := PulumiServiceTaskEntity{}
				if objMap["type"].HasValue() && objMap["type"].IsString() {
					entity.Type = objMap["type"].StringValue()
				}
				if objMap["id"].HasValue() && objMap["id"].IsString() {
					entity.ID = objMap["id"].StringValue()
				}
				input.EntityRemove = append(input.EntityRemove, entity)
			}
		}
	}

	if inputMap["timestamp"].HasValue() && inputMap["timestamp"].IsString() {
		input.Timestamp = inputMap["timestamp"].StringValue()
	}

	return input
}

func (t *PulumiServiceNeoTaskResource) Name() string {
	return "pulumiservice:index:NeoTask"
}

func (t *PulumiServiceNeoTaskResource) Check(req *pulumirpc.CheckRequest) (*pulumirpc.CheckResponse, error) {
	news := req.GetNews()
	newsMap, err := plugin.UnmarshalProperties(news, plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	var failures []*pulumirpc.CheckFailure

	if !newsMap["content"].HasValue() {
		failures = append(failures, &pulumirpc.CheckFailure{
			Reason:   "content is required",
			Property: "content",
		})
	}

	if !newsMap["organizationName"].HasValue() {
		failures = append(failures, &pulumirpc.CheckFailure{
			Reason:   "organizationName is required",
			Property: "organizationName",
		})
	}

	// Set default timestamp to now if not provided
	if !newsMap["timestamp"].HasValue() {
		newsMap["timestamp"] = resource.NewPropertyValue(time.Now().Format(time.RFC3339))
	}

	inputs, err := plugin.MarshalProperties(newsMap, plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	return &pulumirpc.CheckResponse{Inputs: inputs, Failures: failures}, nil
}

func (t *PulumiServiceNeoTaskResource) Delete(req *pulumirpc.DeleteRequest) (*pbempty.Empty, error) {
	// Tasks cannot be deleted once created, they exist as immutable records
	// Return success without error
	return &pbempty.Empty{}, nil
}

func (t *PulumiServiceNeoTaskResource) Diff(req *pulumirpc.DiffRequest) (*pulumirpc.DiffResponse, error) {
	// Tasks are immutable once created, any change requires replacement
	olds, err := plugin.UnmarshalProperties(req.GetOldInputs(), plugin.MarshalOptions{KeepUnknowns: false, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	news, err := plugin.UnmarshalProperties(req.GetNews(), plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: false})
	if err != nil {
		return nil, err
	}

	oldTask := ToPulumiServiceNeoTaskInput(olds)
	newTask := ToPulumiServiceNeoTaskInput(news)

	changes := pulumirpc.DiffResponse_DIFF_NONE
	var replaces []string

	// Any change to a task requires replacement since tasks are immutable
	if oldTask.Content != newTask.Content {
		changes = pulumirpc.DiffResponse_DIFF_SOME
		replaces = append(replaces, "content")
	}

	if oldTask.OrganizationName != newTask.OrganizationName {
		changes = pulumirpc.DiffResponse_DIFF_SOME
		replaces = append(replaces, "organizationName")
	}

	if oldTask.Timestamp != newTask.Timestamp {
		changes = pulumirpc.DiffResponse_DIFF_SOME
		replaces = append(replaces, "timestamp")
	}

	return &pulumirpc.DiffResponse{
		Changes:             changes,
		Replaces:            replaces,
		Stables:             []string{},
		DeleteBeforeReplace: true,
	}, nil
}

func (t *PulumiServiceNeoTaskResource) Read(req *pulumirpc.ReadRequest) (*pulumirpc.ReadResponse, error) {
	ctx := context.Background()

	orgName, taskID, err := splitSingleSlashString(req.Id)
	if err != nil {
		return nil, err
	}

	task, err := t.Client.GetTask(ctx, orgName, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to read Task (%q): %w", req.Id, err)
	}
	if task == nil {
		return &pulumirpc.ReadResponse{}, nil
	}

	inputs := PulumiServiceNeoTaskInput{
		Content:          task.Name,
		OrganizationName: orgName,
		Timestamp:        task.CreatedAt.Format(time.RFC3339),
	}

	for _, entity := range task.Entities {
		inputs.EntityAdd = append(inputs.EntityAdd, PulumiServiceTaskEntity{
			Type: entity.Type,
			ID:   entity.ID,
		})
	}

	props, err := plugin.MarshalProperties(inputs.ToPropertyMap(), plugin.MarshalOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal inputs to properties: %w", err)
	}
	return &pulumirpc.ReadResponse{
		Id:         req.Id,
		Properties: props,
		Inputs:     props,
	}, nil
}

func (t *PulumiServiceNeoTaskResource) Update(req *pulumirpc.UpdateRequest) (*pulumirpc.UpdateResponse, error) {
	// Tasks are immutable, updates should trigger replacement instead
	return nil, fmt.Errorf("tasks are immutable and cannot be updated")
}

func (t *PulumiServiceNeoTaskResource) Create(req *pulumirpc.CreateRequest) (*pulumirpc.CreateResponse, error) {
	ctx := context.Background()
	inputs, err := plugin.UnmarshalProperties(req.GetProperties(), plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	inputsTask := ToPulumiServiceNeoTaskInput(inputs)

	// Parse timestamp
	timestamp, err := time.Parse(time.RFC3339, inputsTask.Timestamp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse timestamp: %w", err)
	}

	// Build entity diff
	var entityDiff *pulumiapi.EntityDiff
	if len(inputsTask.EntityAdd) > 0 || len(inputsTask.EntityRemove) > 0 {
		entityDiff = &pulumiapi.EntityDiff{}

		for _, e := range inputsTask.EntityAdd {
			entityDiff.Add = append(entityDiff.Add, pulumiapi.TaskEntity{
				Type: e.Type,
				ID:   e.ID,
			})
		}

		for _, e := range inputsTask.EntityRemove {
			entityDiff.Remove = append(entityDiff.Remove, pulumiapi.TaskEntity{
				Type: e.Type,
				ID:   e.ID,
			})
		}
	}

	createReq := pulumiapi.CreateTaskRequest{
		Content:    inputsTask.Content,
		EntityDiff: entityDiff,
		Timestamp:  timestamp,
	}

	task, err := t.Client.CreateTask(ctx, inputsTask.OrganizationName, createReq)
	if err != nil {
		return nil, fmt.Errorf("error creating task: %s", err.Error())
	}

	taskID := fmt.Sprintf("%s/%s", inputsTask.OrganizationName, task.ID)

	outputs := PulumiServiceNeoTaskInput{
		Content:          inputsTask.Content,
		OrganizationName: inputsTask.OrganizationName,
		EntityAdd:        inputsTask.EntityAdd,
		EntityRemove:     inputsTask.EntityRemove,
		Timestamp:        inputsTask.Timestamp,
	}

	outputProperties, err := outputs.ToRpc()
	if err != nil {
		return nil, err
	}

	return &pulumirpc.CreateResponse{
		Id:         taskID,
		Properties: outputProperties,
	}, nil
}
