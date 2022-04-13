package provider

import (
	"fmt"
	"strings"

	pbempty "github.com/golang/protobuf/ptypes/empty"
	pulumiapi "github.com/pierskarsenbarg/pulumi-apiclient"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
)

type PulumiServiceWebhookResource struct {
	config PulumiServiceConfig
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

	outputStore := resource.PropertyMap{}
	outputStore["__inputs"] = resource.NewObjectProperty(inputs)

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
	token, err := wh.config.getPulumiAccessToken()
	if err != nil {
		return nil, err
	}

	url, err := wh.config.getPulumiServiceUrl()
	if err != nil {
		return nil, err
	}

	c := pulumiapi.NewClient(*token, *url)
	webhook, err := c.CreateWebhook(input.OrganizationName, input.DisplayName, input.PayloadUrl, input.Secret, input.Active)
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
	if diffs.Changed("active") ||
		diffs.Changed("displayName") ||
		diffs.Changed("payloadUrl") ||
		diffs.Changed("secret") ||
		diffs.Changed("organizationName") ||
		diffs.Changed("name") {
		changes = pulumirpc.DiffResponse_DIFF_SOME
	}

	return &pulumirpc.DiffResponse{
		Changes:             changes,
		Replaces:            []string{},
		Stables:             []string{},
		DeleteBeforeReplace: false,
	}, nil
}

func (wh *PulumiServiceWebhookResource) Update(req *pulumirpc.UpdateRequest) (*pulumirpc.UpdateResponse, error) {
	inputsOld, err := plugin.UnmarshalProperties(req.GetOlds(), plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
	if err != nil {
		return nil, err
	}
	inputsNew, err := plugin.UnmarshalProperties(req.GetNews(), plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	webhookOld := wh.ToPulumiServiceWebhookInput(inputsOld["__inputs"].ObjectValue())
	webhookNew := wh.ToPulumiServiceWebhookInput(inputsNew)

	if webhookOld.Active != webhookNew.Active ||
		webhookOld.DisplayName != webhookNew.DisplayName ||
		webhookOld.PayloadUrl != webhookNew.PayloadUrl ||
		webhookOld.Secret != webhookNew.Secret {
		err = wh.updateWebhook(webhookNew)
		if err != nil {
			return nil, err
		}
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

func (wh *PulumiServiceWebhookResource) updateWebhook(input PulumiServiceWebhookInput) error {
	token, err := wh.config.getPulumiAccessToken()
	if err != nil {
		return err
	}

	url, err := wh.config.getPulumiServiceUrl()
	if err != nil {
		return err
	}

	c := pulumiapi.NewClient(*token, *url)
	_, err = c.UpdateWebhook(input.Name, input.OrganizationName, input.DisplayName, input.PayloadUrl, input.Secret, input.Active)
	if err != nil {
		return err
	}
	return nil
}

func (wh *PulumiServiceWebhookResource) Delete(req *pulumirpc.DeleteRequest) (*pbempty.Empty, error) {
	err := wh.deleteWebhook(req.Id)
	if err != nil {
		return &pbempty.Empty{}, err
	}
	return &pbempty.Empty{}, nil
}

func (wh *PulumiServiceWebhookResource) deleteWebhook(id string) error {
	token, err := wh.config.getPulumiAccessToken()
	if err != nil {
		return err
	}

	url, err := wh.config.getPulumiServiceUrl()
	if err != nil {
		return err
	}

	s := strings.Split(id, "/")

	c := pulumiapi.NewClient(*token, *url)
	err = c.DeleteTeam(s[0], s[1])
	if err != nil {
		return err
	}
	return nil
}
