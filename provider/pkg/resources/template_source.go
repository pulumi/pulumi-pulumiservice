// Copyright 2026, Pulumi Corporation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package resources

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/pulumi/pulumi/sdk/v3/go/property"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/config"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
)

type TemplateSource struct{}

var (
	_ infer.CustomCreate[TemplateSourceInput, TemplateSourceState] = &TemplateSource{}
	_ infer.CustomUpdate[TemplateSourceInput, TemplateSourceState] = &TemplateSource{}
	_ infer.CustomDelete[TemplateSourceState]                      = &TemplateSource{}
	_ infer.CustomRead[TemplateSourceInput, TemplateSourceState]   = &TemplateSource{}
	_ infer.CustomStateMigrations[TemplateSourceState]             = &TemplateSource{}
)

func (*TemplateSource) Annotate(a infer.Annotator) {
	a.Describe(&TemplateSource{}, "A source for Pulumi templates.")
	a.SetToken("index", "TemplateSource")
}

type TemplateSourceDestination struct {
	URL *string `pulumi:"url,optional"`
}

func (d *TemplateSourceDestination) Annotate(a infer.Annotator) {
	a.Describe(&d.URL, "Destination URL that gets filled in on new project creation.")
}

type TemplateSourceInput struct {
	OrganizationName string                     `pulumi:"organizationName" provider:"replaceOnChanges"`
	SourceName       string                     `pulumi:"sourceName"`
	SourceURL        string                     `pulumi:"sourceURL"`
	Destination      *TemplateSourceDestination `pulumi:"destination,optional"`
}

func (i *TemplateSourceInput) Annotate(a infer.Annotator) {
	a.Describe(&i.OrganizationName, "Organization name.")
	a.Describe(&i.SourceName, "Source name.")
	a.Describe(&i.SourceURL, "Github URL of the repository from which to grab templates.")
	a.Describe(&i.Destination, "The default destination for projects using templates from this source.")
}

type TemplateSourceState struct {
	TemplateSourceInput
}

func (*TemplateSource) Create(
	ctx context.Context,
	req infer.CreateRequest[TemplateSourceInput],
) (infer.CreateResponse[TemplateSourceState], error) {
	if req.DryRun {
		return infer.CreateResponse[TemplateSourceState]{
			Output: TemplateSourceState{TemplateSourceInput: req.Inputs},
		}, nil
	}
	resp, err := config.GetClient(ctx).CreateTemplateSource(
		ctx, req.Inputs.OrganizationName, req.Inputs.toAPIRequest(),
	)
	if err != nil {
		return infer.CreateResponse[TemplateSourceState]{}, err
	}
	return infer.CreateResponse[TemplateSourceState]{
		ID:     templateSourceID(req.Inputs.OrganizationName, resp.ID),
		Output: TemplateSourceState{TemplateSourceInput: templateSourceInputFromResponse(req.Inputs.OrganizationName, *resp)},
	}, nil
}

func (*TemplateSource) Update(
	ctx context.Context,
	req infer.UpdateRequest[TemplateSourceInput, TemplateSourceState],
) (infer.UpdateResponse[TemplateSourceState], error) {
	if req.DryRun {
		return infer.UpdateResponse[TemplateSourceState]{
			Output: TemplateSourceState{TemplateSourceInput: req.Inputs},
		}, nil
	}
	_, templateID, err := parseTemplateSourceID(req.ID)
	if err != nil {
		return infer.UpdateResponse[TemplateSourceState]{}, err
	}
	resp, err := config.GetClient(ctx).UpdateTemplateSource(
		ctx, req.Inputs.OrganizationName, templateID, req.Inputs.toAPIRequest(),
	)
	if err != nil {
		return infer.UpdateResponse[TemplateSourceState]{}, err
	}
	return infer.UpdateResponse[TemplateSourceState]{
		Output: TemplateSourceState{TemplateSourceInput: templateSourceInputFromResponse(req.Inputs.OrganizationName, *resp)},
	}, nil
}

func (*TemplateSource) Delete(
	ctx context.Context,
	req infer.DeleteRequest[TemplateSourceState],
) (infer.DeleteResponse, error) {
	orgName, templateID, err := parseTemplateSourceID(req.ID)
	if err != nil {
		return infer.DeleteResponse{}, err
	}
	return infer.DeleteResponse{}, config.GetClient(ctx).DeleteTemplateSource(ctx, orgName, templateID)
}

func (*TemplateSource) Read(
	ctx context.Context,
	req infer.ReadRequest[TemplateSourceInput, TemplateSourceState],
) (infer.ReadResponse[TemplateSourceInput, TemplateSourceState], error) {
	orgName, templateID, err := parseTemplateSourceID(req.ID)
	if err != nil {
		return infer.ReadResponse[TemplateSourceInput, TemplateSourceState]{}, err
	}
	resp, err := config.GetClient(ctx).GetTemplateSource(ctx, orgName, templateID)
	if err != nil {
		return infer.ReadResponse[TemplateSourceInput, TemplateSourceState]{}, fmt.Errorf(
			"failed to get template source during Read. org: %s id: %s due to error: %w",
			orgName, templateID, err,
		)
	}
	if resp == nil {
		return infer.ReadResponse[TemplateSourceInput, TemplateSourceState]{}, nil
	}
	inputs := templateSourceInputFromResponse(orgName, *resp)
	return infer.ReadResponse[TemplateSourceInput, TemplateSourceState]{
		ID:     req.ID,
		Inputs: inputs,
		State:  TemplateSourceState{TemplateSourceInput: inputs},
	}, nil
}

// StateMigrations strips the legacy `__inputs` field that the pre-infer
// TemplateSource resource embedded in state. infer rejects unknown fields when
// decoding state, so without this migration a refresh against an existing
// stack errors with "Unrecognized field '__inputs'".
func (*TemplateSource) StateMigrations(context.Context) []infer.StateMigrationFunc[TemplateSourceState] {
	return []infer.StateMigrationFunc[TemplateSourceState]{
		infer.StateMigration(migrateTemplateSourceLegacyInputs),
	}
}

func migrateTemplateSourceLegacyInputs(
	_ context.Context, old property.Map,
) (infer.MigrationResult[TemplateSourceState], error) {
	if _, ok := old.GetOk(gcInputs); !ok {
		return infer.MigrationResult[TemplateSourceState]{}, nil
	}
	state := TemplateSourceState{}
	if v, ok := old.GetOk(gcOrganizationName); ok && v.IsString() {
		state.OrganizationName = v.AsString()
	}
	if v, ok := old.GetOk(gcSourceName); ok && v.IsString() {
		state.SourceName = v.AsString()
	}
	if v, ok := old.GetOk(gcSourceURL); ok && v.IsString() {
		state.SourceURL = v.AsString()
	}
	if v, ok := old.GetOk("destination"); ok && v.IsMap() {
		dest := TemplateSourceDestination{}
		dm := v.AsMap()
		if u, ok := dm.GetOk("url"); ok && u.IsString() {
			s := u.AsString()
			dest.URL = &s
		}
		state.Destination = &dest
	}
	return infer.MigrationResult[TemplateSourceState]{Result: &state}, nil
}

func (i *TemplateSourceInput) toAPIRequest() pulumiapi.CreateTemplateSourceRequest {
	var destination *pulumiapi.CreateTemplateSourceRequestDestination
	if i.Destination != nil {
		destination = &pulumiapi.CreateTemplateSourceRequestDestination{URL: i.Destination.URL}
	}
	return pulumiapi.CreateTemplateSourceRequest{
		Name:        i.SourceName,
		SourceURL:   i.SourceURL,
		Destination: destination,
	}
}

func templateSourceInputFromResponse(
	organization string, response pulumiapi.TemplateSourceResponse,
) TemplateSourceInput {
	var destination *TemplateSourceDestination
	if response.Destination != nil {
		destination = &TemplateSourceDestination{URL: response.Destination.URL}
	}
	return TemplateSourceInput{
		OrganizationName: organization,
		SourceName:       response.Name,
		SourceURL:        response.SourceURL,
		Destination:      destination,
	}
}

func templateSourceID(orgName, templateID string) string {
	return path.Join(orgName, templateID)
}

func parseTemplateSourceID(id string) (organizationName, templateID string, err error) {
	splitID := strings.Split(id, "/")
	if len(splitID) != 2 {
		return "", "", fmt.Errorf("invalid template source id: %s", id)
	}
	return splitID[0], splitID[1], nil
}
