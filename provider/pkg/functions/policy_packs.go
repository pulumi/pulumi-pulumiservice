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

package functions

import (
	"context"
	"encoding/json"
	"fmt"

	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/pulumi/pulumi/sdk/v3/go/common/apitype"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/config"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
)

const (
	policyPackNameDescription = "The name of the policy pack."
	policyPackDisplayNameDesc = "The display name of the policy pack."
	policyPackSourceDesc      = "Where the policy pack is hosted in the Pulumi Registry: `pulumi` for packs " +
		"published by Pulumi (for example `cis-aws`), `private` for packs published by an organization. " +
		"Omitted when the provider could not determine registry metadata for this pack."
	policyPackPublisherDesc = "The organization or user that published the policy pack. `pulumi` for " +
		"Pulumi-published packs, otherwise the publishing organization's name. Omitted when the provider " +
		"could not determine registry metadata for this pack."
)

// GetPolicyPacksFunction lists the policy packs available to an organization,
// annotated with their Pulumi Registry provenance.
type GetPolicyPacksFunction struct{}

type GetPolicyPacksInput struct {
	OrganizationName string `pulumi:"organizationName"`
}

// PolicyPackSummary is one entry in the getPolicyPacks result.
type PolicyPackSummary struct {
	Name        string   `pulumi:"name"`
	DisplayName string   `pulumi:"displayName"`
	Versions    []int    `pulumi:"versions"`
	VersionTags []string `pulumi:"versionTags"`
	// Source and Publisher are pointers so that unknown provenance is omitted
	// rather than sent as "", which would read as a real value in typed SDKs and
	// compare equal to nothing.
	Source    *string `pulumi:"source,optional"`
	Publisher *string `pulumi:"publisher,optional"`
}

type GetPolicyPacksOutput struct {
	PolicyPacks []PolicyPackSummary `pulumi:"policyPacks"`
}

func (GetPolicyPacksFunction) Annotate(a infer.Annotator) {
	a.Describe(&GetPolicyPacksFunction{}, "Get a list of all policy packs for an organization.")
	a.SetToken("index", "getPolicyPacks")
}

func (i *GetPolicyPacksInput) Annotate(a infer.Annotator) {
	a.Describe(&i.OrganizationName, "The name of the Pulumi organization.")
}

func (o *GetPolicyPacksOutput) Annotate(a infer.Annotator) {
	a.Describe(&o.PolicyPacks, "List of policy packs in the organization.")
}

func (s *PolicyPackSummary) Annotate(a infer.Annotator) {
	a.Describe(s, "Summary metadata for a policy pack available to an organization, including its "+
		"Pulumi Registry provenance.")
	a.Describe(&s.Name, policyPackNameDescription)
	a.Describe(&s.DisplayName, policyPackDisplayNameDesc)
	a.Describe(&s.Versions, "List of version numbers for this policy pack.")
	a.Describe(&s.VersionTags, "List of version tags for this policy pack.")
	a.Describe(&s.Source, policyPackSourceDesc)
	a.Describe(&s.Publisher, policyPackPublisherDesc)
}

func (GetPolicyPacksFunction) Invoke(
	ctx context.Context,
	req infer.FunctionRequest[GetPolicyPacksInput],
) (infer.FunctionResponse[GetPolicyPacksOutput], error) {
	orgName := req.Input.OrganizationName
	if orgName == "" {
		return infer.FunctionResponse[GetPolicyPacksOutput]{}, fmt.Errorf("organizationName is required")
	}

	// Registry provenance (source/publisher) comes from a second endpoint joined
	// on pack name; any failure other than the route being absent is fatal, so a
	// program filtering on publisher can't silently match nothing and attach zero
	// policy packs.
	packs, registryUnavailable, err := config.GetClient(ctx).ListPolicyPacksWithRegistryMetadata(ctx, orgName)
	if err != nil {
		return infer.FunctionResponse[GetPolicyPacksOutput]{}, fmt.Errorf("failed to list policy packs: %w", err)
	}
	if registryUnavailable {
		p.GetLogger(ctx).Warning("this Pulumi backend does not serve the policy pack registry; " +
			"`source` and `publisher` will be absent from getPolicyPacks results")
	}

	out := make([]PolicyPackSummary, len(packs))
	for i, pack := range packs {
		out[i] = policyPackSummaryFromAPI(pack)
	}

	return infer.FunctionResponse[GetPolicyPacksOutput]{
		Output: GetPolicyPacksOutput{PolicyPacks: out},
	}, nil
}

func policyPackSummaryFromAPI(pack pulumiapi.PolicyPackWithRegistryMetadata) PolicyPackSummary {
	versions := pack.Versions
	if versions == nil {
		versions = []int{}
	}
	versionTags := pack.VersionTags
	if versionTags == nil {
		versionTags = []string{}
	}
	return PolicyPackSummary{
		Name:        pack.Name,
		DisplayName: pack.DisplayName,
		Versions:    versions,
		VersionTags: versionTags,
		Source:      nonEmpty(pack.Source),
		Publisher:   nonEmpty(pack.Publisher),
	}
}

// GetPolicyPackFunction reads one version of a policy pack.
type GetPolicyPackFunction struct{}

type GetPolicyPackInput struct {
	OrganizationName string  `pulumi:"organizationName"`
	PolicyPackName   string  `pulumi:"policyPackName"`
	VersionTag       *string `pulumi:"versionTag,optional"`
}

type GetPolicyPackOutput struct {
	Name        string           `pulumi:"name"`
	DisplayName string           `pulumi:"displayName"`
	Version     int              `pulumi:"version"`
	VersionTag  *string          `pulumi:"versionTag,optional"`
	Source      *string          `pulumi:"source,optional"`
	Publisher   *string          `pulumi:"publisher,optional"`
	Config      map[string]any   `pulumi:"config,optional"`
	Policies    []map[string]any `pulumi:"policies,optional"`
}

func (GetPolicyPackFunction) Annotate(a infer.Annotator) {
	a.Describe(&GetPolicyPackFunction{}, "Get details about a specific version of a policy pack.")
	a.SetToken("index", "getPolicyPack")
}

func (i *GetPolicyPackInput) Annotate(a infer.Annotator) {
	a.Describe(&i.OrganizationName, "The name of the Pulumi organization.")
	a.Describe(&i.PolicyPackName, policyPackNameDescription)
	a.Describe(&i.VersionTag, "The version tag of the policy pack, for example `1.4.1`, as shown in "+
		"Pulumi Cloud. If not specified, returns the latest version.")
}

func (o *GetPolicyPackOutput) Annotate(a infer.Annotator) {
	a.Describe(&o.Name, policyPackNameDescription)
	a.Describe(&o.DisplayName, policyPackDisplayNameDesc)
	a.Describe(&o.Version, "The version number.")
	a.Describe(&o.VersionTag, "The version tag (if any).")
	a.Describe(&o.Source, policyPackSourceDesc)
	a.Describe(&o.Publisher, policyPackPublisherDesc)
	a.Describe(&o.Config, "Configuration for the policy pack.")
	a.Describe(&o.Policies, "List of policies in this pack.")
}

func (GetPolicyPackFunction) Invoke(
	ctx context.Context,
	req infer.FunctionRequest[GetPolicyPackInput],
) (infer.FunctionResponse[GetPolicyPackOutput], error) {
	orgName := req.Input.OrganizationName
	if orgName == "" {
		return infer.FunctionResponse[GetPolicyPackOutput]{}, fmt.Errorf("organizationName is required")
	}
	packName := req.Input.PolicyPackName
	if packName == "" {
		return infer.FunctionResponse[GetPolicyPackOutput]{}, fmt.Errorf("policyPackName is required")
	}

	client := config.GetClient(ctx)

	var (
		pack *pulumiapi.PolicyPackDetail
		err  error
	)
	if tag := req.Input.VersionTag; tag != nil && *tag != "" {
		pack, err = client.GetPolicyPack(ctx, orgName, packName, *tag)
	} else {
		pack, err = client.GetLatestPolicyPack(ctx, orgName, packName)
	}
	if err != nil {
		return infer.FunctionResponse[GetPolicyPackOutput]{}, fmt.Errorf("failed to get policy pack: %w", err)
	}
	if pack == nil {
		return infer.FunctionResponse[GetPolicyPackOutput]{}, fmt.Errorf("policy pack not found")
	}

	// Look the pack up by the version tag we just resolved. Provenance is version
	// independent, but the tag is not optional in practice: omitting it makes the
	// service look for the literal tag `latest`, which packs do not carry, so
	// source and publisher would come back empty for effectively every pack.
	//
	// Note the deliberate asymmetry with getPolicyPacks, which treats a registry
	// failure as fatal: there, a failed lookup silently empties a publisher filter
	// and the program attaches nothing. Here the caller already named the pack, so
	// that hazard doesn't exist and failing would only deny them a result the
	// provider could already produce in full before this feature. Don't "fix" one
	// to match the other.
	registryPack, registryErr := client.GetRegistryPolicyPack(ctx, orgName, packName, pack.VersionTag)
	if registryErr != nil {
		p.GetLogger(ctx).Warningf("could not read registry metadata for policy pack %q: %v", packName, registryErr)
		registryPack = nil
	}

	return infer.FunctionResponse[GetPolicyPackOutput]{
		Output: policyPackOutputFromAPI(pack, registryPack),
	}, nil
}

// policyPackOutputFromAPI renders a policy pack for getPolicyPack. registry
// carries the pack's Pulumi Registry provenance and may be nil when that lookup
// was unavailable; source and publisher are then omitted.
func policyPackOutputFromAPI(
	pack *pulumiapi.PolicyPackDetail,
	registry *pulumiapi.RegistryPolicyPack,
) GetPolicyPackOutput {
	out := GetPolicyPackOutput{
		Name:        pack.Name,
		DisplayName: pack.DisplayName,
		Version:     pack.Version,
		VersionTag:  nonEmpty(pack.VersionTag),
		Config:      pack.Config,
	}

	if registry != nil {
		out.Source = nonEmpty(registry.Source)
		out.Publisher = nonEmpty(registry.Publisher)
	}

	if len(pack.Policies) > 0 {
		out.Policies = make([]map[string]any, len(pack.Policies))
		for i, policy := range pack.Policies {
			out.Policies[i] = policyToMap(policy)
		}
	}

	return out
}

func policyToMap(policy apitype.Policy) map[string]any {
	m := map[string]any{"name": policy.Name}
	if policy.DisplayName != "" {
		m["displayName"] = policy.DisplayName
	}
	if policy.Description != "" {
		m["description"] = policy.Description
	}
	if policy.EnforcementLevel != "" {
		m["enforcementLevel"] = string(policy.EnforcementLevel)
	}
	if policy.Message != "" {
		m["message"] = policy.Message
	}
	if policy.ConfigSchema != nil {
		m["configSchema"] = configSchemaToMap(policy.ConfigSchema)
	}
	return m
}

func configSchemaToMap(cs *apitype.PolicyConfigSchema) map[string]any {
	m := map[string]any{}
	if cs.Type != "" {
		m["type"] = string(cs.Type)
	}
	if len(cs.Required) > 0 {
		required := make([]any, len(cs.Required))
		for i, r := range cs.Required {
			required[i] = r
		}
		m["required"] = required
	}
	if len(cs.Properties) > 0 {
		properties := map[string]any{}
		for name, raw := range cs.Properties {
			if raw == nil {
				continue
			}
			var decoded any
			if err := json.Unmarshal(*raw, &decoded); err != nil {
				continue
			}
			properties[name] = decoded
		}
		m["properties"] = properties
	}
	return m
}

func nonEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
