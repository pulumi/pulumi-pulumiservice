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

// Not the best to create a second source of truth here, but this will likely not change for years
var defaultWebhookGroups = map[string][]string{
	"organization": {"deployments", "environments", "stacks"},
	"stack":        {"deployments", "stacks"},
	"environment":  {"environments"},
}

type PulumiServiceWebhookResource struct {
	Client pulumiapi.WebhookClient
}

type PulumiServiceWebhookInput struct {
	Active           bool
	DisplayName      string
	PayloadUrl       string
	Secret           *pulumiapi.SecretValue
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

func (i *PulumiServiceWebhookInput) ToPropertyMap(
	plaintextSecret *pulumiapi.SecretValue,
	cipherSecret *pulumiapi.SecretValue,
	isInput bool,
) resource.PropertyMap {
	createMode := plaintextSecret != nil && cipherSecret == nil
	mergeMode := plaintextSecret != nil && cipherSecret != nil

	pm := resource.PropertyMap{}
	pm["active"] = resource.NewPropertyValue(i.Active)
	pm["displayName"] = resource.NewPropertyValue(i.DisplayName)
	pm["payloadUrl"] = resource.NewPropertyValue(i.PayloadUrl)
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

	if i.Secret != nil {
		if mergeMode {
			util.MergeSecretValue(pm, "secret", *i.Secret, plaintextSecret, cipherSecret, isInput)
		} else if createMode {
			util.CreateSecretValue(pm, "secret", *i.Secret, *plaintextSecret, isInput)
		} else {
			util.ImportSecretValue(pm, "secret", *i.Secret, isInput)
		}
	}

	return pm
}

func (i *PulumiServiceWebhookProperties) ToPropertyMap(
	plaintextSecret *pulumiapi.SecretValue,
	cipherSecret *pulumiapi.SecretValue,
	isInput bool,
) resource.PropertyMap {
	pm := i.PulumiServiceWebhookInput.ToPropertyMap(plaintextSecret, cipherSecret, isInput)

	if !isInput {
		pm["name"] = resource.NewPropertyValue(i.Name)
	}

	return pm
}

func (wh *PulumiServiceWebhookResource) ToPulumiServiceWebhookProperties(
	propMap resource.PropertyMap,
) PulumiServiceWebhookProperties {
	props := PulumiServiceWebhookProperties{}

	props.DisplayName = util.GetSecretOrStringValue(propMap["displayName"])
	props.PayloadUrl = util.GetSecretOrStringValue(propMap["payloadUrl"])
	props.OrganizationName = util.GetSecretOrStringValue(propMap["organizationName"])
	props.ProjectName = util.GetSecretOrStringNullableValue(propMap["projectName"])
	props.StackName = util.GetSecretOrStringNullableValue(propMap["stackName"])
	props.EnvironmentName = util.GetSecretOrStringNullableValue(propMap["environmentName"])
	props.Format = util.GetSecretOrStringNullableValue(propMap["format"])
	props.Name = util.GetSecretOrStringValue(propMap["name"])

	if propMap["active"].HasValue() && propMap["active"].IsBool() {
		props.Active = propMap["active"].BoolValue()
	}

	if propMap["filters"].HasValue() && propMap["filters"].IsArray() {
		filtersInput := propMap["filters"].ArrayValue()
		filters := make([]string, len(filtersInput))

		for i, v := range filtersInput {
			filters[i] = util.GetSecretOrStringValue(v)
		}

		props.Filters = filters
	}

	if propMap["groups"].HasValue() && propMap["groups"].IsArray() {
		groupsInput := propMap["groups"].ArrayValue()
		groups := make([]string, len(groupsInput))

		for i, v := range groupsInput {
			groups[i] = util.GetSecretOrStringValue(v)
		}

		props.Groups = groups
	}

	if propMap["secret"].HasValue() || propMap["secretCipher"].HasValue() {
		props.Secret = &pulumiapi.SecretValue{
			Secret: true,
			Value:  util.GetSecretOrStringValue(propMap["secret"]),
		}
	}

	return props
}

func (wh *PulumiServiceWebhookResource) Name() string {
	return "pulumiservice:index:Webhook"
}

func (wh *PulumiServiceWebhookResource) Check(req *pulumirpc.CheckRequest) (*pulumirpc.CheckResponse, error) {
	news, err := plugin.UnmarshalProperties(
		req.GetNews(),
		plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true, KeepSecrets: true},
	)
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
	ctx := context.Background()
	inputMap, err := plugin.UnmarshalProperties(
		req.GetProperties(),
		plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true, KeepSecrets: true},
	)
	if err != nil {
		return nil, err
	}

	inputProps := wh.ToPulumiServiceWebhookProperties(inputMap)

	var secretStr *string = nil
	if inputProps.Secret != nil {
		secretStr = &inputProps.Secret.Value
	}

	request := pulumiapi.WebhookRequest{
		OrganizationName: inputProps.OrganizationName,
		ProjectName:      inputProps.ProjectName,
		StackName:        inputProps.StackName,
		EnvironmentName:  inputProps.EnvironmentName,
		DisplayName:      inputProps.DisplayName,
		PayloadURL:       inputProps.PayloadUrl,
		Secret:           secretStr,
		Active:           inputProps.Active,
		Format:           inputProps.Format,
		Filters:          inputProps.Filters,
		Groups:           inputProps.Groups,
	}
	webhook, err := wh.Client.CreateWebhook(ctx, request)
	if err != nil {
		return nil, err
	}
	props := inputProps
	props.Name = webhook.Name
	if secretStr != nil && webhook.HasSecret {
		props.Secret.Value = webhook.SecretCiphertext
	}

	properties, err := plugin.MarshalProperties(
		props.ToPropertyMap(inputProps.Secret, nil, false),
		plugin.MarshalOptions{
			KeepUnknowns: true,
			SkipNulls:    true,
		},
	)
	if err != nil {
		return nil, err
	}

	return &pulumirpc.CreateResponse{
		Id:         generateWebhookID(props.PulumiServiceWebhookInput, *webhook),
		Properties: properties,
	}, nil
}

func (wh *PulumiServiceWebhookResource) Diff(req *pulumirpc.DiffRequest) (*pulumirpc.DiffResponse, error) {
	olds, err := plugin.UnmarshalProperties(
		req.GetOldInputs(),
		plugin.MarshalOptions{KeepUnknowns: false, SkipNulls: true},
	)
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
	inputMap, err := plugin.UnmarshalProperties(
		req.GetNews(),
		plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true},
	)
	if err != nil {
		return nil, err
	}

	inputProps := wh.ToPulumiServiceWebhookProperties(inputMap)

	hookID, err := splitWebhookID(req.GetId())
	if err != nil {
		return nil, fmt.Errorf("invalid resource id: %v", err)
	}

	var secretStr *string = nil
	if inputProps.Secret != nil {
		secretStr = &inputProps.Secret.Value
	}

	updateReq := pulumiapi.UpdateWebhookRequest{
		WebhookRequest: pulumiapi.WebhookRequest{
			OrganizationName: inputProps.OrganizationName,
			ProjectName:      inputProps.ProjectName,
			StackName:        inputProps.StackName,
			EnvironmentName:  inputProps.EnvironmentName,
			DisplayName:      inputProps.DisplayName,
			PayloadURL:       inputProps.PayloadUrl,
			Secret:           secretStr,
			Active:           inputProps.Active,
			Format:           inputProps.Format,
			Filters:          inputProps.Filters,
			Groups:           inputProps.Groups,
		},
		Name: hookID.webhookName,
	}
	webhook, err := wh.Client.UpdateWebhook(context.Background(), updateReq)
	if err != nil {
		return nil, err
	}
	props := inputProps

	if secretStr != nil && webhook.HasSecret {
		props.Secret.Value = webhook.SecretCiphertext
	}

	properties, err := plugin.MarshalProperties(
		props.ToPropertyMap(inputProps.Secret, nil, false),
		plugin.MarshalOptions{
			KeepUnknowns: true,
			SkipNulls:    true,
			KeepSecrets:  true,
		},
	)
	if err != nil {
		return nil, err
	}
	return &pulumirpc.UpdateResponse{
		Properties: properties,
	}, nil

}

func (wh *PulumiServiceWebhookResource) Delete(req *pulumirpc.DeleteRequest) (*pbempty.Empty, error) {
	hookID, err := splitWebhookID(req.Id)
	if err != nil {
		return nil, err
	}

	err = wh.Client.DeleteWebhook(context.Background(), hookID.organizationName,
		hookID.projectName, hookID.stackName, hookID.environmentName, hookID.webhookName)

	return &pbempty.Empty{}, err
}

func (wh *PulumiServiceWebhookResource) Read(req *pulumirpc.ReadRequest) (*pulumirpc.ReadResponse, error) {
	hookID, err := splitWebhookID(req.Id)
	if err != nil {
		return nil, err
	}
	webhook, err := wh.Client.GetWebhook(context.Background(),
		hookID.organizationName, hookID.projectName, hookID.stackName, hookID.environmentName, hookID.webhookName)
	if err != nil {
		return nil, err
	}

	if webhook == nil {
		return &pulumirpc.ReadResponse{}, nil
	}

	var secret *pulumiapi.SecretValue = nil
	if webhook.HasSecret {
		secret = &pulumiapi.SecretValue{
			Value:  webhook.SecretCiphertext,
			Secret: true,
		}
	}

	retrievedProperties := PulumiServiceWebhookProperties{
		PulumiServiceWebhookInput: PulumiServiceWebhookInput{
			Active:           webhook.Active,
			DisplayName:      webhook.DisplayName,
			PayloadUrl:       webhook.PayloadUrl,
			Secret:           secret,
			Format:           &webhook.Format,
			Filters:          webhook.Filters,
			Groups:           webhook.Groups,
			OrganizationName: hookID.organizationName,
			ProjectName:      hookID.projectName,
			StackName:        hookID.stackName,
			EnvironmentName:  hookID.environmentName,
		},
		Name: webhook.Name,
	}

	var plaintextSecret *pulumiapi.SecretValue
	var ciphertextSecret *pulumiapi.SecretValue
	propertyMap, err := plugin.UnmarshalProperties(
		req.GetProperties(),
		plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true, KeepSecrets: true},
	)
	if err != nil {
		return nil, err
	}
	inputMap, err := plugin.UnmarshalProperties(
		req.GetInputs(),
		plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true, KeepSecrets: true},
	)
	if err != nil {
		return nil, err
	}
	if propertyMap["secret"].HasValue() {
		plaintextSecret = wh.ToPulumiServiceWebhookProperties(inputMap).Secret
		ciphertextSecret = wh.ToPulumiServiceWebhookProperties(propertyMap).Secret
	}

	properties, err := plugin.MarshalProperties(
		retrievedProperties.ToPropertyMap(plaintextSecret, ciphertextSecret, false),
		plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true, KeepSecrets: true},
	)
	if err != nil {
		return nil, err
	}

	inputs, err := plugin.MarshalProperties(
		retrievedProperties.ToPropertyMap(plaintextSecret, ciphertextSecret, true),
		plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true, KeepSecrets: true},
	)
	if err != nil {
		return nil, err
	}

	return &pulumirpc.ReadResponse{
		Id:         req.Id,
		Properties: properties,
		Inputs:     inputs,
	}, nil
}

func generateWebhookID(input PulumiServiceWebhookInput, webhook pulumiapi.Webhook) string {
	if input.ProjectName != nil && input.StackName != nil {
		return fmt.Sprintf("%s/%s/%s/%s", input.OrganizationName, *input.ProjectName, *input.StackName,
			webhook.Name)
	} else if input.ProjectName != nil && input.EnvironmentName != nil {
		// This is not ideal, but inserting "environment" string to distinguish from stack webhooks
		return fmt.Sprintf("%s/environment/%s/%s/%s", input.OrganizationName, *input.ProjectName, *input.EnvironmentName,
			webhook.Name)
	}
	return fmt.Sprintf("%s/%s", input.OrganizationName, webhook.Name)
}

func splitWebhookID(id string) (*webhookID, error) {
	// format:
	// organization/project/stack/webhookName (stack webhook)
	// organization/webhookName (org webhook)
	// organization/environment/projectName/environmentName/webhookName (environment webhook)
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
