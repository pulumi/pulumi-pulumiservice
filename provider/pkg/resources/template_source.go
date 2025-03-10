package resources

import (
	"context"
	"fmt"
	"path"
	"strings"

	pbempty "google.golang.org/protobuf/types/known/emptypb"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/util"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
)

type PulumiServiceTemplateSourceResource struct {
	Client *pulumiapi.Client
}

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

func (s *PulumiServiceTemplateSourceResource) Diff(req *pulumirpc.DiffRequest) (*pulumirpc.DiffResponse, error) {
	return util.StandardDiff(req, []string{"organizationName"}, false)
}

func (s *PulumiServiceTemplateSourceResource) Delete(req *pulumirpc.DeleteRequest) (*pbempty.Empty, error) {
	ctx := context.Background()
	orgName, templateId, err := parseTemplateSourceID(req.Id)
	if err != nil {
		return nil, err
	}
	err = s.Client.DeleteTemplateSource(ctx, *orgName, *templateId)
	if err != nil {
		return nil, err
	}
	return &pbempty.Empty{}, nil
}

func (s *PulumiServiceTemplateSourceResource) Create(req *pulumirpc.CreateRequest) (*pulumirpc.CreateResponse, error) {
	ctx := context.Background()
	inputMap, err := plugin.UnmarshalProperties(req.GetProperties(), plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	input, err := s.ToPulumiServiceTemplateSourceInput(inputMap)
	if err != nil {
		return nil, err
	}

	response, err := s.Client.CreateTemplateSource(ctx, input.OrganizationName, input.toRequest())
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

func (s *PulumiServiceTemplateSourceResource) Check(req *pulumirpc.CheckRequest) (*pulumirpc.CheckResponse, error) {
	return &pulumirpc.CheckResponse{Inputs: req.News, Failures: nil}, nil
}

func (s *PulumiServiceTemplateSourceResource) Update(req *pulumirpc.UpdateRequest) (*pulumirpc.UpdateResponse, error) {
	ctx := context.Background()
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

	response, err := s.Client.UpdateTemplateSource(ctx, input.OrganizationName, *templateId, input.toRequest())
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

func (s *PulumiServiceTemplateSourceResource) Read(req *pulumirpc.ReadRequest) (*pulumirpc.ReadResponse, error) {
	ctx := context.Background()
	orgName, templateId, err := parseTemplateSourceID(req.Id)
	if err != nil {
		return nil, err
	}

	response, err := s.Client.GetTemplateSource(ctx, *orgName, *templateId)
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
