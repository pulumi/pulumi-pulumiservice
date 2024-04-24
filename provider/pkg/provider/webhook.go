package provider

import (
	"context"
	"fmt"
	"strings"

	pbempty "google.golang.org/protobuf/types/known/emptypb"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/internal/pulumiapi"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
)

type PulumiServiceWebhookResource struct{}

type PulumiServiceWebhookInput struct {
	Active           bool
	DisplayName      string
	PayloadUrl       string
	Secret           *string
	OrganizationName string
	ProjectName      *string
	StackName        *string
	Format           *string
	Filters          []string
}

type PulumiServiceWebhookProperties struct {
	PulumiServiceWebhookInput
	Name string
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
	if i.Format != nil {
		pm["format"] = resource.NewPropertyValue(*i.Format)
	}
	if len(i.Filters) > 0 {
		pm["filters"] = resource.NewPropertyValue(i.Filters)
	}

	return pm
}

func (i *PulumiServiceWebhookProperties) ToPropertyMap() resource.PropertyMap {
	pm := i.PulumiServiceWebhookInput.ToPropertyMap()

	pm["name"] = resource.NewPropertyValue(i.Name)
	return pm
}

func (wh *PulumiServiceWebhookResource) ToPulumiServiceWebhookProperties(propMap resource.PropertyMap) PulumiServiceWebhookProperties {
	props := PulumiServiceWebhookProperties{}

	if propMap["active"].HasValue() && propMap["active"].IsBool() {
		props.Active = propMap["active"].BoolValue()
	}

	if propMap["displayName"].HasValue() && propMap["displayName"].IsString() {
		props.DisplayName = propMap["displayName"].StringValue()
	}

	if propMap["payloadUrl"].HasValue() && propMap["payloadUrl"].IsString() {
		props.PayloadUrl = propMap["payloadUrl"].StringValue()
	}

	if secretVal := propMap["secret"]; secretVal.HasValue() && secretVal.IsString() {
		secretStr := secretVal.StringValue()
		props.Secret = &secretStr
	}

	if propMap["organizationName"].HasValue() && propMap["organizationName"].IsString() {
		props.OrganizationName = propMap["organizationName"].StringValue()
	}
	if propMap["projectName"].HasValue() && propMap["projectName"].IsString() {
		projectNameStr := propMap["projectName"].StringValue()
		props.ProjectName = &projectNameStr
	}
	if propMap["stackName"].HasValue() && propMap["stackName"].IsString() {
		stackNameStr := propMap["stackName"].StringValue()
		props.StackName = &stackNameStr
	}
	if propMap["format"].HasValue() && propMap["format"].IsString() {
		formatStr := propMap["format"].StringValue()
		props.Format = &formatStr
	}
	if propMap["filters"].HasValue() && propMap["filters"].IsArray() {
		filtersInput := propMap["filters"].ArrayValue()
		filters := make([]string, len(filtersInput))

		for i, v := range filtersInput {
			filters[i] = getSecretOrStringValue(v)
		}

		props.Filters = filters
	}

	if nameVal, ok := propMap["name"]; ok && nameVal.IsString() {
		props.Name = nameVal.StringValue()
	}

	return props
}

func (wh *PulumiServiceWebhookResource) Name() string {
	return "pulumiservice:index:Webhook"
}

func (wh *PulumiServiceWebhookResource) Check(_ context.Context, req *pulumirpc.CheckRequest) (*pulumirpc.CheckResponse, error) {
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

	stackWebhookError := "projectName and stackName must both be specified for stack webhooks, or both unspecified for org webhooks"
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

	// if the format is not specified, default to raw
	// this should work automatically because we have set the default in the schema,
	// but it isn't respected by the yaml provider
	// https://github.com/pulumi/pulumi-yaml/issues/458
	if !news["format"].HasValue() {
		news["format"] = resource.NewPropertyValue("raw")
	}

	inputNews, err := plugin.MarshalProperties(
		news,
		plugin.MarshalOptions{},
	)
	if err != nil {
		return nil, err
	}

	return &pulumirpc.CheckResponse{Inputs: inputNews, Failures: failures}, nil
}

func (wh *PulumiServiceWebhookResource) Create(ctx context.Context, req *pulumirpc.CreateRequest) (*pulumirpc.CreateResponse, error) {
	inputs, err := plugin.UnmarshalProperties(req.GetProperties(), plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	props := wh.ToPulumiServiceWebhookProperties(inputs)

	idString, err := wh.createWebhook(ctx, props.PulumiServiceWebhookInput)
	if err != nil {
		return nil, err
	}

	hookID, err := splitWebhookID(*idString)
	if err != nil {
		return nil, err
	}

	props.Name = hookID.webhookName

	outputProperties, err := plugin.MarshalProperties(
		props.ToPropertyMap(),
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

func (wh *PulumiServiceWebhookResource) createWebhook(ctx context.Context, input PulumiServiceWebhookInput) (*string, error) {
	req := pulumiapi.WebhookRequest{
		OrganizationName: input.OrganizationName,
		ProjectName:      input.ProjectName,
		StackName:        input.StackName,
		DisplayName:      input.DisplayName,
		PayloadURL:       input.PayloadUrl,
		Secret:           input.Secret,
		Active:           input.Active,
		Format:           input.Format,
		Filters:          input.Filters,
	}
	webhook, err := GetClient[pulumiapi.WebhookClient](ctx).CreateWebhook(ctx, req)
	if err != nil {
		return nil, err
	}

	var hookID string
	if input.ProjectName != nil && input.StackName != nil {
		hookID = fmt.Sprintf("%s/%s/%s/%s", input.OrganizationName, *input.ProjectName, *input.StackName,
			webhook.Name)
	} else {
		hookID = fmt.Sprintf("%s/%s", input.OrganizationName, webhook.Name)
	}

	return &hookID, nil
}

func (wh *PulumiServiceWebhookResource) Diff(_ context.Context, req *pulumirpc.DiffRequest) (*pulumirpc.DiffResponse, error) {
	olds, err := plugin.UnmarshalProperties(req.GetOlds(), plugin.MarshalOptions{KeepUnknowns: false, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	// preprocess olds to remove the `name` property since it's only an output and shouldn't cause a diff
	if olds["name"].HasValue() {
		delete(olds, "name")
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
		"projectName":      true,
		"stackName":        true,
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
	if len(detailedDiffs) > 0 {
		changes = pulumirpc.DiffResponse_DIFF_SOME
	}
	return &pulumirpc.DiffResponse{
		Changes:         changes,
		DetailedDiff:    detailedDiffs,
		HasDetailedDiff: true,
	}, nil
}

func (wh *PulumiServiceWebhookResource) Update(ctx context.Context, req *pulumirpc.UpdateRequest) (*pulumirpc.UpdateResponse, error) {
	// we only care about news because we validated that everything was correctly set in Check() & Diff()
	inputsNew, err := plugin.UnmarshalProperties(req.GetNews(), plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	webhookNew := wh.ToPulumiServiceWebhookProperties(inputsNew)

	// ignore orgName because if that changed, we would have done a replace, so update would never have been called
	hookID, err := splitWebhookID(req.GetId())
	if err != nil {
		return nil, fmt.Errorf("invalid resource id: %v", err)
	}
	webhookNew.Name = hookID.webhookName

	updateReq := pulumiapi.UpdateWebhookRequest{
		WebhookRequest: pulumiapi.WebhookRequest{
			OrganizationName: webhookNew.OrganizationName,
			ProjectName:      webhookNew.ProjectName,
			StackName:        webhookNew.StackName,
			DisplayName:      webhookNew.DisplayName,
			PayloadURL:       webhookNew.PayloadUrl,
			Secret:           webhookNew.Secret,
			Active:           webhookNew.Active,
			Format:           webhookNew.Format,
			Filters:          webhookNew.Filters,
		},
		Name: webhookNew.Name,
	}
	err = GetClient[pulumiapi.WebhookClient](ctx).UpdateWebhook(ctx, updateReq)
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

func (wh *PulumiServiceWebhookResource) Delete(ctx context.Context, req *pulumirpc.DeleteRequest) (*pbempty.Empty, error) {
	return &pbempty.Empty{}, wh.deleteWebhook(ctx, req.Id)
}

func (wh *PulumiServiceWebhookResource) deleteWebhook(ctx context.Context, id string) error {
	hookID, err := splitWebhookID(id)
	if err != nil {
		return err
	}
	return GetClient[pulumiapi.WebhookClient](ctx).DeleteWebhook(ctx, hookID.organizationName,
		hookID.projectName, hookID.stackName, hookID.webhookName)
}

func (wh *PulumiServiceWebhookResource) Read(ctx context.Context, req *pulumirpc.ReadRequest) (*pulumirpc.ReadResponse, error) {
	webhook, err := wh.getWebhook(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	if webhook == nil {
		return &pulumirpc.ReadResponse{}, nil
	}

	hookID, err := splitWebhookID(req.Id)
	if err != nil {
		return nil, err
	}

	properties := PulumiServiceWebhookProperties{
		PulumiServiceWebhookInput: PulumiServiceWebhookInput{
			Active:           webhook.Active,
			DisplayName:      webhook.DisplayName,
			PayloadUrl:       webhook.PayloadUrl,
			Secret:           webhook.Secret,
			Format:           &webhook.Format,
			Filters:          webhook.Filters,
			OrganizationName: hookID.organizationName,
			ProjectName:      hookID.projectName,
			StackName:        hookID.stackName,
		},
		Name: hookID.webhookName,
	}

	outputs, err := plugin.MarshalProperties(
		properties.ToPropertyMap(),
		plugin.MarshalOptions{},
	)
	if err != nil {
		return nil, err
	}

	inputs, err := plugin.MarshalProperties(
		properties.PulumiServiceWebhookInput.ToPropertyMap(),
		plugin.MarshalOptions{},
	)
	if err != nil {
		return nil, err
	}

	return &pulumirpc.ReadResponse{
		Id:         req.Id,
		Properties: outputs,
		Inputs:     inputs,
	}, nil
}

func (wh *PulumiServiceWebhookResource) getWebhook(ctx context.Context, id string) (*pulumiapi.Webhook, error) {
	hookID, err := splitWebhookID(id)
	if err != nil {
		return nil, err
	}
	webhook, err := GetClient[pulumiapi.WebhookClient](ctx).GetWebhook(ctx,
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
