package provider

import (
	"context"
	"fmt"
	"sort"
	"strings"

	pbempty "github.com/golang/protobuf/ptypes/empty"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/internal/pulumiapi"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
)

type PulumiServiceWebhookResource struct {
	config PulumiServiceConfig
	client *pulumiapi.Client
}

type PulumiServiceWebhookInput struct {
	Active           bool
	DisplayName      string
	PayloadUrl       string
	Secret           *string
	Name             string
	OrganizationName string
	ProjectName      *string
	StackName        *string
}

func (i *PulumiServiceWebhookInput) ToPropertyMap() resource.PropertyMap {
	pm := resource.PropertyMap{}
	pm["active"] = resource.NewPropertyValue(i.Active)
	pm["displayName"] = resource.NewPropertyValue(i.DisplayName)
	pm["payloadUrl"] = resource.NewPropertyValue(i.PayloadUrl)
	if i.Secret != nil {
		pm["secret"] = resource.NewPropertyValue(*i.Secret)
	}

	pm["organizationName"] = resource.NewPropertyValue(i.OrganizationName)

	if i.ProjectName != nil {
		pm["projectName"] = resource.NewPropertyValue(*i.ProjectName)
	}
	if i.StackName != nil {
		pm["stackName"] = resource.NewPropertyValue(*i.StackName)
	}

	pm["name"] = resource.NewPropertyValue(i.Name)
	return pm
}

func (wh *PulumiServiceWebhookResource) ToPulumiServiceWebhookInput(inputMap resource.PropertyMap) PulumiServiceWebhookInput {
	input := PulumiServiceWebhookInput{}

	if inputMap["active"].HasValue() && inputMap["active"].IsBool() {
		input.Active = inputMap["active"].BoolValue()
	}

	if inputMap["displayName"].HasValue() && inputMap["displayName"].IsString() {
		input.DisplayName = inputMap["displayName"].StringValue()
	}

	if inputMap["payloadUrl"].HasValue() && inputMap["payloadUrl"].IsString() {
		input.PayloadUrl = inputMap["payloadUrl"].StringValue()
	}

	if secretVal := inputMap["secret"]; secretVal.HasValue() && secretVal.IsString() {
		secretStr := secretVal.StringValue()
		input.Secret = &secretStr
	}

	if inputMap["organizationName"].HasValue() && inputMap["organizationName"].IsString() {
		input.OrganizationName = inputMap["organizationName"].StringValue()
	}
	if inputMap["projectName"].HasValue() && inputMap["projectName"].IsString() {
		projectNameStr := inputMap["projectName"].StringValue()
		input.ProjectName = &projectNameStr
	}
	if inputMap["stackName"].HasValue() && inputMap["stackName"].IsString() {
		stackNameStr := inputMap["stackName"].StringValue()
		input.StackName = &stackNameStr
	}

	if nameVal, ok := inputMap["name"]; ok && nameVal.IsString() {
		input.Name = nameVal.StringValue()
	}

	return input
}

func (wh *PulumiServiceWebhookResource) Name() string {
	return "pulumiservice:index:Webhook"
}

func (wh *PulumiServiceWebhookResource) Configure(config PulumiServiceConfig) {
	wh.config = config
}

func (wh *PulumiServiceWebhookResource) Check(req *pulumirpc.CheckRequest) (*pulumirpc.CheckResponse, error) {
	news, err := plugin.UnmarshalProperties(req.GetNews(), plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	var failures []*pulumirpc.CheckFailure
	for _, p := range []resource.PropertyKey{"organizationName", "payloadUrl", "displayName", "active"} {
		if !news[(p)].HasValue() {
			failures = append(failures, &pulumirpc.CheckFailure{
				Reason:   fmt.Sprintf("missing required property '%s'", p),
				Property: string(p),
			})
		}
	}

	stackWebhookError := "both projectName and stackName must both be specified for stack webhooks, or both unspecified for org webhooks"
	if !news["projectName"].HasValue() && news["stackName"].HasValue() {
		failures = append(failures, &pulumirpc.CheckFailure{
			Reason:   stackWebhookError,
			Property: "projectName",
		})
	}
	if news["projectName"].HasValue() && !news["stackName"].HasValue() {
		failures = append(failures, &pulumirpc.CheckFailure{
			Reason:   stackWebhookError,
			Property: "stackName",
		})
	}

	return &pulumirpc.CheckResponse{Inputs: req.News, Failures: failures}, nil
}

func (wh *PulumiServiceWebhookResource) Create(req *pulumirpc.CreateRequest) (*pulumirpc.CreateResponse, error) {
	inputs, err := plugin.UnmarshalProperties(req.GetProperties(), plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	inputsWebhook := wh.ToPulumiServiceWebhookInput(inputs)

	idString, err := wh.createWebhook(inputsWebhook)
	if err != nil {
		return nil, err
	}

	hookID, err := splitWebhookID(*idString)
	if err != nil {
		return nil, err
	}

	inputsWebhook.Name = hookID.webhookName

	outputProperties, err := plugin.MarshalProperties(
		inputsWebhook.ToPropertyMap(),
		plugin.MarshalOptions{
			KeepUnknowns: true,
			SkipNulls:    true,
		},
	)
	if err != nil {
		return nil, err
	}

	return &pulumirpc.CreateResponse{
		Id:         *idString,
		Properties: outputProperties,
	}, nil
}

func (wh *PulumiServiceWebhookResource) createWebhook(input PulumiServiceWebhookInput) (*string, error) {
	ctx := context.Background()
	req := pulumiapi.CreateWebhookRequest{
		OrganizationName: input.OrganizationName,
		ProjectName:      input.ProjectName,
		StackName:        input.StackName,
		DisplayName:      input.DisplayName,
		PayloadURL:       input.PayloadUrl,
		Secret:           input.Secret,
		Active:           input.Active,
	}
	webhook, err := wh.client.CreateWebhook(ctx, req)
	if err != nil {
		return nil, err
	}

	var webhookID string
	if input.ProjectName != nil && input.StackName != nil {
		webhookID = fmt.Sprintf("%s/%s/%s/%s", input.OrganizationName, *input.ProjectName, *input.StackName,
			webhook.Name)
	} else {
		webhookID = fmt.Sprintf("%s/%s", input.OrganizationName, webhook.Name)
	}

	return &webhookID, nil
}

func (wh *PulumiServiceWebhookResource) Diff(req *pulumirpc.DiffRequest) (*pulumirpc.DiffResponse, error) {
	olds, err := plugin.UnmarshalProperties(req.GetOlds(), plugin.MarshalOptions{KeepUnknowns: false, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	news, err := plugin.UnmarshalProperties(req.GetNews(), plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	// previous versions of the provider used "__inputs" key to store inputs in output properties
	// to maintain backwards compatibility, we still need to handle this case
	// so we just lift up those values to the top level
	if oldInputs, ok := olds["__inputs"]; ok && oldInputs.IsObject() {
		for k, v := range oldInputs.ObjectValue() {
			olds[k] = v
		}
	}

	changes := pulumirpc.DiffResponse_DIFF_NONE
	var diffs, replaces []string
	properties := map[string]bool{
		"active":           false,
		"displayName":      false,
		"payloadUrl":       false,
		"secret":           false,
		"organizationName": true,
		"projectName":      true,
		"stackName":        true,
	}
	if d := olds.Diff(news); d != nil {
		for key, replace := range properties {
			i := sort.SearchStrings(req.IgnoreChanges, key)
			if i < len(req.IgnoreChanges) && req.IgnoreChanges[i] == key {
				continue
			}

			if d.Changed(resource.PropertyKey(key)) {
				changes = pulumirpc.DiffResponse_DIFF_SOME
				diffs = append(diffs, key)

				if replace {
					replaces = append(replaces, key)
				}
			}
		}
	}

	return &pulumirpc.DiffResponse{
		Diffs:               diffs,
		Changes:             changes,
		Replaces:            replaces,
		DeleteBeforeReplace: false,
	}, nil
}

func (wh *PulumiServiceWebhookResource) Update(req *pulumirpc.UpdateRequest) (*pulumirpc.UpdateResponse, error) {
	// we only care about news because we validated that everything was correctly set in Check() & Diff()
	inputsNew, err := plugin.UnmarshalProperties(req.GetNews(), plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	webhookNew := wh.ToPulumiServiceWebhookInput(inputsNew)

	// ignore orgName because if that changed, we would have done a replace, so update would never have been called
	hookID, err := splitWebhookID(req.GetId())
	if err != nil {
		return nil, fmt.Errorf("invalid resource id: %v", err)
	}
	webhookNew.Name = hookID.webhookName

	updateReq := pulumiapi.UpdateWebhookRequest{
		OrganizationName: webhookNew.OrganizationName,
		ProjectName:      webhookNew.ProjectName,
		StackName:        webhookNew.StackName,
		DisplayName:      webhookNew.DisplayName,
		PayloadURL:       webhookNew.PayloadUrl,
		Secret:           webhookNew.Secret,
		Active:           webhookNew.Active,
		Name:             webhookNew.Name,
	}
	err = wh.client.UpdateWebhook(context.Background(), updateReq)
	if err != nil {
		return nil, err
	}

	outputStore := webhookNew.ToPropertyMap()

	outputProperties, err := plugin.MarshalProperties(
		outputStore,
		plugin.MarshalOptions{},
	)
	if err != nil {
		return nil, err
	}
	return &pulumirpc.UpdateResponse{
		Properties: outputProperties,
	}, nil

}

func (wh *PulumiServiceWebhookResource) Delete(req *pulumirpc.DeleteRequest) (*pbempty.Empty, error) {
	err := wh.deleteWebhook(req.Id)
	return &pbempty.Empty{}, err
}

func (wh *PulumiServiceWebhookResource) deleteWebhook(id string) error {
	hookID, err := splitWebhookID(id)
	if err != nil {
		return err
	}
	return wh.client.DeleteWebhook(context.Background(), hookID.organizationName,
		hookID.projectName, hookID.stackName, hookID.webhookName)
}

func (wh *PulumiServiceWebhookResource) Read(req *pulumirpc.ReadRequest) (*pulumirpc.ReadResponse, error) {
	inputs, err := plugin.UnmarshalProperties(req.GetProperties(), plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	webhook, err := wh.getWebhook(req.Id)
	if err != nil {
		return nil, err
	}

	hookID, err := splitWebhookID(req.Id)
	if err != nil {
		return nil, err
	}

	webhookInput := PulumiServiceWebhookInput{
		Active:           webhook.Active,
		DisplayName:      webhook.DisplayName,
		PayloadUrl:       webhook.PayloadUrl,
		Secret:           webhook.Secret,
		OrganizationName: hookID.organizationName,
		ProjectName:      hookID.projectName,
		StackName:        hookID.stackName,
		Name:             hookID.webhookName,
	}

	properties, err := plugin.MarshalProperties(
		webhookInput.ToPropertyMap(),
		plugin.MarshalOptions{},
	)

	if err != nil {
		return nil, err
	}

	inputProperties, err := plugin.MarshalProperties(
		inputs,
		plugin.MarshalOptions{},
	)

	if err != nil {
		return nil, err
	}

	return &pulumirpc.ReadResponse{
		Id:         req.Id,
		Properties: properties,
		Inputs:     inputProperties,
	}, nil
}

func (wh *PulumiServiceWebhookResource) getWebhook(id string) (*pulumiapi.Webhook, error) {
	hookID, err := splitWebhookID(id)
	if err != nil {
		return nil, err
	}
	webhook, err := wh.client.GetWebhook(context.Background(),
		hookID.organizationName, hookID.projectName, hookID.stackName, hookID.webhookName)
	if err != nil {
		return nil, err
	}
	return webhook, nil
}

func splitWebhookID(id string) (*webhookID, error) {
	// format: organization/project/stack/webhookName (stack webhook) or organization/webhookName (org webhook)
	s := strings.Split(id, "/")
	switch len(s) {
	case 2:
		return &webhookID{
			organizationName: s[0],
			webhookName:      s[1],
		}, nil
	case 4:
		return &webhookID{
			organizationName: s[0],
			projectName:      &s[1],
			stackName:        &s[2],
			webhookName:      s[3],
		}, nil
	default:
		return nil, fmt.Errorf("%q is not a valid webhook ID", id)
	}
}

type webhookID struct {
	organizationName string
	projectName      *string
	stackName        *string
	webhookName      string
}
