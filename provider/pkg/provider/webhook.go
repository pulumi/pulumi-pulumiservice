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

// Not the best to create a second source of truth here, but this will likely not change for years
var defaultWebhookGroups = map[string][]string{
	"organization": {"deployments", "environments", "stacks"},
	"stack":        {"deployments", "stacks"},
	"environment":  {"environments"},
}

type PulumiServiceWebhookResource struct {
	config PulumiServiceConfig
	client pulumiapi.WebhookClient
}

type PulumiServiceWebhookInput struct {
	Active           bool
	DisplayName      string
	PayloadUrl       string
	Secret           *string
	OrganizationName string
	ProjectName      *string
	StackName        *string
	EnvironmentName  *string
	Format           *string
	Filters          []string
	Groups           []string
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
	if i.EnvironmentName != nil {
		pm["environmentName"] = resource.NewPropertyValue(*i.EnvironmentName)
	}
	if i.Format != nil {
		pm["format"] = resource.NewPropertyValue(*i.Format)
	}
	if len(i.Filters) > 0 {
		pm["filters"] = resource.NewPropertyValue(i.Filters)
	}
	if len(i.Groups) > 0 {
		pm["groups"] = resource.NewPropertyValue(i.Groups)
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
	if propMap["environmentName"].HasValue() && propMap["environmentName"].IsString() {
		environmentNameStr := propMap["environmentName"].StringValue()
		props.EnvironmentName = &environmentNameStr
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
	if propMap["groups"].HasValue() && propMap["groups"].IsArray() {
		groupsInput := propMap["groups"].ArrayValue()
		groups := make([]string, len(groupsInput))

		for i, v := range groupsInput {
			groups[i] = getSecretOrStringValue(v)
		}

		props.Groups = groups
	}

	if nameVal, ok := propMap["name"]; ok && nameVal.IsString() {
		props.Name = nameVal.StringValue()
	}

	return props
}

func (wh *PulumiServiceWebhookResource) Name() string {
	return "pulumiservice:index:Webhook"
}

func (wh *PulumiServiceWebhookResource) Configure(config PulumiServiceConfig) {
	wh.config = config
}

func (wh *PulumiServiceWebhookResource) Check(req *pulumirpc.CheckRequest) (*pulumirpc.CheckResponse, error) {
	news, err := plugin.UnmarshalProperties(req.GetNews(), plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true, KeepSecrets: true})
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

	if news["stackName"].HasValue() && !news["projectName"].HasValue() {
		failures = append(failures, &pulumirpc.CheckFailure{
			Reason:   "projectName and stackName must both be specified for stack webhooks",
			Property: "projectName",
		})
	}
	if news["environmentName"].HasValue() && !news["projectName"].HasValue() {
		failures = append(failures, &pulumirpc.CheckFailure{
			Reason:   "projectName and environmentName must both be specified for environment webhooks",
			Property: "projectName",
		})
	}
	if news["environmentName"].HasValue() && news["stackName"].HasValue() {
		failures = append(failures, &pulumirpc.CheckFailure{
			Reason: "stackName needs to be empty if this is meant to be an environment webhook; " +
				"environmentName needs to be empty if this is meant to be a stack webhook",
			Property: "stackName",
		})
	}
	if news["projectName"].HasValue() && !news["stackName"].HasValue() && !news["environmentName"].HasValue() {
		failures = append(failures, &pulumirpc.CheckFailure{
			Reason: "projectName needs to be empty if this is meant to be an organization webhook; " +
				"otherwise provide stackName for stack webhook or environmentName for environment webhook",
			Property: "projectName",
		})
	}

	// if the format is not specified, default to raw
	// this should work automatically because we have set the default in the schema,
	// but it isn't respected by the yaml provider
	// https://github.com/pulumi/pulumi-yaml/issues/458
	if !news["format"].HasValue() {
		news["format"] = resource.NewPropertyValue("raw")
	}

	// if neither filters nor groups are specified, set default groups
	if !news["filters"].HasValue() && !news["groups"].HasValue() {
		var groups []string
		if news["stackName"].HasValue() {
			groups = defaultWebhookGroups["stack"]
		} else if news["environmentName"].HasValue() {
			groups = defaultWebhookGroups["environment"]
		} else {
			groups = defaultWebhookGroups["organization"]
		}

		var groupProps []resource.PropertyValue
		for _, group := range groups {
			groupProps = append(groupProps, resource.NewStringProperty(group))
		}

		news["groups"] = resource.NewArrayProperty(groupProps)
	}

	inputNews, err := plugin.MarshalProperties(
		news,
		plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true, KeepSecrets: true},
	)
	if err != nil {
		return nil, err
	}

	return &pulumirpc.CheckResponse{Inputs: inputNews, Failures: failures}, nil
}

func (wh *PulumiServiceWebhookResource) Create(req *pulumirpc.CreateRequest) (*pulumirpc.CreateResponse, error) {
	inputs, err := plugin.UnmarshalProperties(req.GetProperties(), plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	props := wh.ToPulumiServiceWebhookProperties(inputs)

	idString, err := wh.createWebhook(props.PulumiServiceWebhookInput)
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

func (wh *PulumiServiceWebhookResource) createWebhook(input PulumiServiceWebhookInput) (*string, error) {
	ctx := context.Background()
	req := pulumiapi.WebhookRequest{
		OrganizationName: input.OrganizationName,
		ProjectName:      input.ProjectName,
		StackName:        input.StackName,
		EnvironmentName:  input.EnvironmentName,
		DisplayName:      input.DisplayName,
		PayloadURL:       input.PayloadUrl,
		Secret:           input.Secret,
		Active:           input.Active,
		Format:           input.Format,
		Filters:          input.Filters,
		Groups:           input.Groups,
	}
	webhook, err := wh.client.CreateWebhook(ctx, req)
	if err != nil {
		return nil, err
	}

	var hookID string
	if input.ProjectName != nil && input.StackName != nil {
		hookID = fmt.Sprintf("%s/%s/%s/%s", input.OrganizationName, *input.ProjectName, *input.StackName,
			webhook.Name)
	} else if input.ProjectName != nil && input.EnvironmentName != nil {
		// This is not ideal, but inserting "environment" string to distinguish from stack webhooks
		hookID = fmt.Sprintf("%s/%s/%s/%s/%s", input.OrganizationName, "environment", *input.ProjectName, *input.EnvironmentName,
			webhook.Name)
	} else {
		hookID = fmt.Sprintf("%s/%s", input.OrganizationName, webhook.Name)
	}

	return &hookID, nil
}

func (wh *PulumiServiceWebhookResource) Diff(req *pulumirpc.DiffRequest) (*pulumirpc.DiffResponse, error) {
	olds, err := plugin.UnmarshalProperties(req.GetOldInputs(), plugin.MarshalOptions{KeepUnknowns: false, SkipNulls: true})
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
		"environmentName":  true,
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

func (wh *PulumiServiceWebhookResource) Update(req *pulumirpc.UpdateRequest) (*pulumirpc.UpdateResponse, error) {
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
			EnvironmentName:  webhookNew.EnvironmentName,
			DisplayName:      webhookNew.DisplayName,
			PayloadURL:       webhookNew.PayloadUrl,
			Secret:           webhookNew.Secret,
			Active:           webhookNew.Active,
			Format:           webhookNew.Format,
			Filters:          webhookNew.Filters,
			Groups:           webhookNew.Groups,
		},
		Name: webhookNew.Name,
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
		hookID.projectName, hookID.stackName, hookID.environmentName, hookID.webhookName)
}

func (wh *PulumiServiceWebhookResource) Read(req *pulumirpc.ReadRequest) (*pulumirpc.ReadResponse, error) {
	webhook, err := wh.getWebhook(req.Id)
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
			Groups:           webhook.Groups,
			OrganizationName: hookID.organizationName,
			ProjectName:      hookID.projectName,
			StackName:        hookID.stackName,
			EnvironmentName:  hookID.environmentName,
		},
		Name: hookID.webhookName,
	}

	outputs, err := plugin.MarshalProperties(
		properties.ToPropertyMap(),
		plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true, KeepSecrets: true},
	)
	if err != nil {
		return nil, err
	}

	inputs, err := plugin.MarshalProperties(
		properties.PulumiServiceWebhookInput.ToPropertyMap(),
		plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true, KeepSecrets: true},
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

func (wh *PulumiServiceWebhookResource) getWebhook(id string) (*pulumiapi.Webhook, error) {
	hookID, err := splitWebhookID(id)
	if err != nil {
		return nil, err
	}
	webhook, err := wh.client.GetWebhook(context.Background(),
		hookID.organizationName, hookID.projectName, hookID.stackName, hookID.environmentName, hookID.webhookName)
	if err != nil {
		return nil, err
	}
	return webhook, nil
}

func splitWebhookID(id string) (*webhookID, error) {
	// format:
	// organization/project/stack/webhookName (stack webhook)
	// organization/webhookName (org webhook)
	// organization/environment/project/stack/webhookName (stack webhook)
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
	case 5:
		return &webhookID{
			organizationName: s[0],
			projectName:      &s[2],
			environmentName:  &s[3],
			webhookName:      s[4],
		}, nil
	default:
		return nil, fmt.Errorf("%q is not a valid webhook ID", id)
	}
}

type webhookID struct {
	organizationName string
	projectName      *string
	stackName        *string
	environmentName  *string
	webhookName      string
}
