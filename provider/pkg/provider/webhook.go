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
	return &pulumirpc.CheckResponse{Inputs: req.News, Failures: nil}, nil
}

func (wh *PulumiServiceWebhookResource) Create(req *pulumirpc.CreateRequest) (*pulumirpc.CreateResponse, error) {
	inputs, err := plugin.UnmarshalProperties(req.GetProperties(), plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	inputsWebhook := wh.ToPulumiServiceWebhookInput(inputs)

	webhookId, err := wh.createWebhook(inputsWebhook)
	if err != nil {
		return nil, err
	}

	_, webhookName, err := splitSingleSlashString(*webhookId) 
	if err != nil {
		return nil, err
	}

	inputsWebhook.Name = webhookName

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
		Id:         *webhookId,
		Properties: outputProperties,
	}, nil
}

func (wh *PulumiServiceWebhookResource) createWebhook(input PulumiServiceWebhookInput) (*string, error) {
	ctx := context.Background()
	webhook, err := wh.client.CreateWebhook(ctx, input.OrganizationName, input.DisplayName, input.PayloadUrl, input.Secret, input.Active)
	if err != nil {
		return nil, err
	}

	webhookId := fmt.Sprintf("%s/%s", input.OrganizationName, webhook.Name)
	return &webhookId, nil
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
		Diffs:    diffs,
		Changes:  changes,
		Replaces: replaces,
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
	_, webhookName, err := splitSingleSlashString(req.GetId())
	if err != nil {
		return nil, fmt.Errorf("invalid resource id: %v", err)
	}
	webhookNew.Name = webhookName

	err = wh.client.UpdateWebhook(
		context.Background(),
		webhookNew.Name,
		webhookNew.OrganizationName,
		webhookNew.DisplayName,
		webhookNew.PayloadUrl,
		webhookNew.Secret,
		webhookNew.Active,
	)
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
	orgName, webhookName, err := splitSingleSlashString(id)
	if err != nil {
		return err
	}
	return wh.client.DeleteWebhook(context.Background(), orgName, webhookName)
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

	orgName, webhookName, err := splitSingleSlashString(req.Id)
	if err != nil {
		return nil, err
	}

	webhookInput := PulumiServiceWebhookInput{
		Active:           webhook.Active,
		DisplayName:      webhook.DisplayName,
		PayloadUrl:       webhook.PayloadUrl,
		Secret:           webhook.Secret,
		OrganizationName: orgName,
		Name:             webhookName,
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
	org, webhookName, err := splitSingleSlashString(id)
	if err != nil {
		return nil, err
	}
	webhook, err := wh.client.GetWebhook(context.Background(), org, webhookName)
	if err != nil {
		return nil, err
	}
	return webhook, nil
}

func splitSingleSlashString(id string) (string, string, error) {
	// format: organization/webhookName
	s := strings.Split(id, "/")
	if len(s) != 2 {
		return "", "", fmt.Errorf("%q is invalid, must contain a single slash ('/')", id)
	}
	return s[0], s[1], nil
}


