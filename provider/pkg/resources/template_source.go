package resources

import (
	"context"
	"fmt"
	"path"
	"strings"

	pbempty "google.golang.org/protobuf/types/known/emptypb"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/config"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
)

type PulumiServiceTemplateSourceResource struct{}

type PulumiServiceTemplateSourceDestination struct {
	Url *string
}

type PulumiServiceTemplateSourceInput struct {
	OrganizationName string
	SourceName       string
	SourceURL        string
	Destination      *PulumiServiceTemplateSourceDestination
}

func (i *PulumiServiceTemplateSourceInput) ToPropertyMap() resource.PropertyMap {
	pm := resource.PropertyMap{}
	pm["organizationName"] = resource.NewPropertyValue(i.OrganizationName)
	pm["sourceName"] = resource.NewPropertyValue(i.SourceName)
	pm["sourceURL"] = resource.NewPropertyValue(i.SourceURL)
	if i.Destination != nil {
		destinationMap := resource.PropertyMap{}
		if i.Destination.Url != nil {
			destinationMap["url"] = resource.NewPropertyValue(i.Destination.Url)
		}
		pm["destination"] = resource.NewObjectProperty(destinationMap)
	}
	return pm
}

func (s *PulumiServiceTemplateSourceResource) ToPulumiServiceTemplateSourceInput(inputMap resource.PropertyMap) (*PulumiServiceTemplateSourceInput, error) {
	input := PulumiServiceTemplateSourceInput{}

	input.OrganizationName = inputMap["organizationName"].StringValue()
	input.SourceName = inputMap["sourceName"].StringValue()
	input.SourceURL = inputMap["sourceURL"].StringValue()

	if inputMap["destination"].HasValue() && inputMap["destination"].IsObject() {
		destinationMap := inputMap["destination"].ObjectValue()
		destination := PulumiServiceTemplateSourceDestination{}
		if destinationMap["url"].HasValue() && destinationMap["url"].IsString() {
			value := destinationMap["url"].StringValue()
			destination.Url = &value
		}
		input.Destination = &destination
	}
	return &input, nil
}

func (s *PulumiServiceTemplateSourceResource) Name() string {
	return "pulumiservice:index:TemplateSource"
}

func (s *PulumiServiceTemplateSourceResource) Diff(ctx context.Context, req *pulumirpc.DiffRequest) (*pulumirpc.DiffResponse, error) {
	olds, err := plugin.UnmarshalProperties(req.GetOldInputs(), plugin.MarshalOptions{KeepUnknowns: false, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	news, err := plugin.UnmarshalProperties(req.GetNews(), plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	diffs := olds.Diff(news)
	if diffs == nil {
		return &pulumirpc.DiffResponse{
			Changes: pulumirpc.DiffResponse_DIFF_NONE,
		}, nil
	}

	dd := plugin.NewDetailedDiffFromObjectDiff(diffs, false)

	detailedDiffs := map[string]*pulumirpc.PropertyDiff{}
	replaces := []string(nil)
	replaceProperties := map[string]bool{
		"organizationName": true,
	}
	for k, v := range dd {
		if _, ok := replaceProperties[k]; ok {
			v.Kind = v.Kind.AsReplace()
			replaces = append(replaces, k)
		}
		detailedDiffs[k] = &pulumirpc.PropertyDiff{
			Kind:      pulumirpc.PropertyDiff_Kind(v.Kind),
			InputDiff: v.InputDiff,
		}
	}

	return &pulumirpc.DiffResponse{
		Changes:             pulumirpc.DiffResponse_DIFF_SOME,
		Replaces:            replaces,
		DetailedDiff:        detailedDiffs,
		HasDetailedDiff:     true,
		DeleteBeforeReplace: true,
	}, nil
}

func (s *PulumiServiceTemplateSourceResource) Delete(ctx context.Context, req *pulumirpc.DeleteRequest) (*pbempty.Empty, error) {
	client := config.GetClient[*pulumiapi.Client](ctx)
	orgName, templateId, err := parseTemplateSourceID(req.Id)
	if err != nil {
		return nil, err
	}
	err = client.DeleteTemplateSource(ctx, *orgName, *templateId)
	if err != nil {
		return nil, err
	}
	return &pbempty.Empty{}, nil
}

func (s *PulumiServiceTemplateSourceResource) Create(ctx context.Context, req *pulumirpc.CreateRequest) (*pulumirpc.CreateResponse, error) {
	client := config.GetClient[*pulumiapi.Client](ctx)
	inputMap, err := plugin.UnmarshalProperties(req.GetProperties(), plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	input, err := s.ToPulumiServiceTemplateSourceInput(inputMap)
	if err != nil {
		return nil, err
	}

	response, err := client.CreateTemplateSource(ctx, input.OrganizationName, input.toRequest())
	if err != nil {
		return nil, err
	}

	outputProperties, err := plugin.MarshalProperties(
		toProperties(input.OrganizationName, *response).ToPropertyMap(),
		plugin.MarshalOptions{
			KeepUnknowns: true,
			SkipNulls:    true,
			KeepSecrets:  true,
		},
	)
	if err != nil {
		return nil, err
	}

	return &pulumirpc.CreateResponse{
		Id:         path.Join(input.OrganizationName, response.Id),
		Properties: outputProperties,
	}, nil
}

func (s *PulumiServiceTemplateSourceResource) Check(ctx context.Context, req *pulumirpc.CheckRequest) (*pulumirpc.CheckResponse, error) {
	return &pulumirpc.CheckResponse{Inputs: req.News, Failures: nil}, nil
}

func (s *PulumiServiceTemplateSourceResource) Update(ctx context.Context, req *pulumirpc.UpdateRequest) (*pulumirpc.UpdateResponse, error) {
	client := config.GetClient[*pulumiapi.Client](ctx)
	inputMap, err := plugin.UnmarshalProperties(req.GetNews(), plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	input, err := s.ToPulumiServiceTemplateSourceInput(inputMap)
	if err != nil {
		return nil, err
	}

	_, templateId, err := parseTemplateSourceID(req.Id)
	if err != nil {
		return nil, err
	}

	response, err := client.UpdateTemplateSource(ctx, input.OrganizationName, *templateId, input.toRequest())
	if err != nil {
		return nil, err
	}

	outputProperties, err := plugin.MarshalProperties(
		toProperties(input.OrganizationName, *response).ToPropertyMap(),
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
		Properties: outputProperties,
	}, nil
}

func (s *PulumiServiceTemplateSourceResource) Read(ctx context.Context, req *pulumirpc.ReadRequest) (*pulumirpc.ReadResponse, error) {
	client := config.GetClient[*pulumiapi.Client](ctx)
	orgName, templateId, err := parseTemplateSourceID(req.Id)
	if err != nil {
		return nil, err
	}

	response, err := client.GetTemplateSource(ctx, *orgName, *templateId)
	if err != nil {
		return nil, fmt.Errorf("failed to get template source during Read. org: %s id: %s due to error: %w", *orgName, *templateId, err)
	}
	if response == nil {
		return &pulumirpc.ReadResponse{}, nil
	}

	properties, err := plugin.MarshalProperties(
		toProperties(*orgName, *response).ToPropertyMap(),
		plugin.MarshalOptions{},
	)
	if err != nil {
		return nil, err
	}

	return &pulumirpc.ReadResponse{
		Id:         path.Join(*orgName, *templateId),
		Properties: properties,
		Inputs:     properties,
	}, nil
}

func parseTemplateSourceID(id string) (organizationName *string, templateId *string, err error) {
	splitID := strings.Split(id, "/")
	if len(splitID) != 2 {
		return nil, nil, fmt.Errorf("invalid template source id: %s", id)
	}
	return &splitID[0], &splitID[1], nil
}

func (input *PulumiServiceTemplateSourceInput) toRequest() pulumiapi.CreateTemplateSourceRequest {
	var destination *pulumiapi.CreateTemplateSourceRequestDestination
	if input.Destination != nil {
		destination = &pulumiapi.CreateTemplateSourceRequestDestination{
			URL: input.Destination.Url,
		}
	} else {
		destination = nil
	}

	return pulumiapi.CreateTemplateSourceRequest{
		Name:        input.SourceName,
		SourceURL:   input.SourceURL,
		Destination: destination,
	}
}

func toProperties(organization string, response pulumiapi.TemplateSourceResponse) *PulumiServiceTemplateSourceInput {
	var destination *PulumiServiceTemplateSourceDestination
	if response.Destination != nil {
		destination = &PulumiServiceTemplateSourceDestination{
			Url: response.Destination.URL,
		}
	} else {
		destination = nil
	}

	return &PulumiServiceTemplateSourceInput{
		OrganizationName: organization,
		SourceName:       response.Name,
		SourceURL:        response.SourceURL,
		Destination:      destination,
	}
}
