package provider

import (
	"context"
	"fmt"
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
	Secret           string
	Name             string
	OrganizationName string
}

func (i *PulumiServiceWebhookInput) ToPropertyMap() resource.PropertyMap {
	pm := resource.PropertyMap{}
	pm["active"] = resource.NewPropertyValue(i.Active)
	pm["displayName"] = resource.NewPropertyValue(i.DisplayName)
	pm["payloadUrl"] = resource.NewPropertyValue(i.PayloadUrl)
	pm["secret"] = resource.NewPropertyValue(i.Secret)
	pm["organizationName"] = resource.NewPropertyValue(i.OrganizationName)
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

	if inputMap["secret"].HasValue() && inputMap["secret"].IsString() {
		input.Secret = inputMap["secret"].StringValue()
	}

	if inputMap["organizationName"].HasValue() && inputMap["organizationName"].IsString() {
		input.OrganizationName = inputMap["organizationName"].StringValue()
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

	s := strings.Split(*webhookId, "/")

	outputStore := resource.PropertyMap{}
	outputStore["__inputs"] = resource.NewObjectProperty(inputs)
	outputStore["name"] = resource.NewPropertyValue(s[1])

	outputProperties, err := plugin.MarshalProperties(
		outputStore,
		plugin.MarshalOptions{},
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

	news, err := plugin.UnmarshalProperties(req.GetNews(), plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: false})
	if err != nil {
		return nil, err
	}

	diffs := olds["__inputs"].ObjectValue().Diff(news)
	if diffs == nil {
		return &pulumirpc.DiffResponse{
			Changes:             pulumirpc.DiffResponse_DIFF_NONE,
			Replaces:            []string{},
			Stables:             []string{},
			DeleteBeforeReplace: false,
		}, nil
	}

	changes := pulumirpc.DiffResponse_DIFF_NONE
	replaceProperties := []string{"organizationName", "name"}
	var replaces []string
	for _, prop := range replaceProperties {
		if diffs.Changed(resource.PropertyKey(prop)) {
			replaces = append(replaces, prop)
			changes = pulumirpc.DiffResponse_DIFF_SOME
		}
	}

	if diffs.Changed("active") ||
		diffs.Changed("displayName") ||
		diffs.Changed("payloadUrl") ||
		diffs.Changed("secret") {
		changes = pulumirpc.DiffResponse_DIFF_SOME
	}

	return &pulumirpc.DiffResponse{
		Changes:             changes,
		Replaces:            replaces,
		Stables:             []string{},
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

	outputStore := resource.PropertyMap{}
	outputStore["__inputs"] = resource.NewObjectProperty(webhookNew.ToPropertyMap())

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
		return &pulumirpc.ReadResponse{}, err
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
		return "", "", fmt.Errorf("id %q is invalid, must contain a single slash ('/')", id)
	}
	return s[0], s[1], nil
}
