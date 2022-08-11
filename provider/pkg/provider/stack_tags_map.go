package provider

import (
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/internal/pulumiapi"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/internal/serde"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
)

type PulumiServiceStackTagMapResource struct {
	client *pulumiapi.Client
}

type PulumiServiceStackTagMapInput struct {
	Organization string            `pulumi:"organization"`
	Project      string            `pulumi:"project"`
	Stack        string            `pulumi:"stack"`
	Tags         map[string]string `pulumi:"tags"`
}

func (i *PulumiServiceStackTagMapInput) ToPropertyMap() resource.PropertyMap {
  properties := resource.NewPropertyMapFromMap(map[string]interface{}{
  })
  return properties
}

func (st *PulumiServiceStackTagMapResource) ToPulumiServiceStackTagInput(inputMap resource.PropertyMap) PulumiServiceStackTagMapInput {
	input := PulumiServiceStackTagMapInput{}
	serde.FromPropertyMap(inputMap, structTagKey, &input)
	return input
}

