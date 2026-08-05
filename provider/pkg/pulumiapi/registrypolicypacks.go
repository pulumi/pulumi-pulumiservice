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

package pulumiapi

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
)

// tagQueryParam selects a specific policy pack version tag on registry reads.
const tagQueryParam = "tag"

// RegistryPolicyPackClient reads policy pack provenance from the Pulumi Registry.
//
// The org-scoped list endpoint (GET /api/orgs/{org}/policypacks) carries no
// ownership information, so the registry is the only source for whether a pack
// was published by Pulumi or by the organization itself.
type RegistryPolicyPackClient interface {
	ListRegistryPolicyPacks(ctx context.Context, orgName string) ([]RegistryPolicyPack, error)
	GetRegistryPolicyPack(
		ctx context.Context, orgName, policyPackName, versionTag string,
	) (*RegistryPolicyPack, error)
	ListPolicyPacksWithRegistryMetadata(
		ctx context.Context, orgName string,
	) ([]PolicyPackWithRegistryMetadata, bool, error)
}

// RegistryPolicyPack is the registry's view of a policy pack.
//
// Packs published by Pulumi are Source/Publisher "pulumi" (for example cis-aws);
// packs an organization published itself are Source "private" with the org as
// Publisher.
//
// Version is a plain string rather than a parsed semver: the generated
// apitype.RegistryPolicyPack models it as semver.Version, which fails to
// unmarshal outright if the service ever returns a non-semver tag. Provenance is
// the field we care about here, and it should not be lost to a version parse.
type RegistryPolicyPack struct {
	ID                string   `json:"id,omitempty"`
	Source            string   `json:"source"`
	Publisher         string   `json:"publisher"`
	Name              string   `json:"name"`
	Version           string   `json:"version,omitempty"`
	DisplayName       string   `json:"displayName,omitempty"`
	AccessLevel       string   `json:"accessLevel,omitempty"`
	EnforcementLevels []string `json:"enforcementLevels,omitempty"`
}

// PolicyPackWithRegistryMetadata is an organization's policy pack plus its
// registry provenance.
//
// Source and Publisher are best effort and may be empty: the registry may have
// no entry for the pack's name, may have more than one (see
// indexRegistryPolicyPacksByName), or the backend may not serve the registry
// route at all. Callers must treat empty as "unknown", not as a value.
type PolicyPackWithRegistryMetadata struct {
	PolicyPackWithVersions
	Source    string `json:"source,omitempty"`
	Publisher string `json:"publisher,omitempty"`
}

type listRegistryPolicyPacksRequest struct {
	OrgLogin string `json:"orgLogin,omitempty"`
}

type listRegistryPolicyPacksResponse struct {
	// ContinuationToken is decoded for completeness but is not actionable:
	// neither the GET nor the POST form of this endpoint accepts a continuation
	// token in the request, so there is no reachable next page to fetch.
	ContinuationToken string               `json:"continuationToken,omitempty"`
	PolicyPacks       []RegistryPolicyPack `json:"policyPacks"`
}

type getRegistryPolicyPackResponse struct {
	PolicyPack RegistryPolicyPack `json:"policyPack"`
}

// ListRegistryPolicyPacks lists the registry entries for policy packs owned by or
// available to orgName.
//
// This uses the POST form of /registry/policypacks. The GET form at the same path
// is marked deprecated in the Pulumi Cloud API spec and explicitly superseded by
// this one.
//
// The route is NOT under /preview, even though its registry siblings
// (/preview/registry/packages, /templates, /sources) are — policypacks is served
// only at the unprefixed path. Getting this wrong is quiet rather than loud: the
// router answers an unknown path with a bare 404, which
// ListPolicyPacksWithRegistryMetadata reads as "this backend has no registry" and
// degrades to returning every pack unannotated, with no error to trace back.
//
// The `access` filter is intentionally omitted so the service default applies,
// which is what the Pulumi Cloud console relies on to render its publisher column.
func (c *Client) ListRegistryPolicyPacks(ctx context.Context, orgName string) ([]RegistryPolicyPack, error) {
	if len(orgName) == 0 {
		return nil, errors.New("empty orgName")
	}

	apiPath := path.Join("registry", "policypacks")
	req := listRegistryPolicyPacksRequest{OrgLogin: orgName}

	var response listRegistryPolicyPacksResponse
	if _, err := c.do(ctx, http.MethodPost, apiPath, req, &response); err != nil {
		return nil, fmt.Errorf("failed to list registry policy packs for %q: %w", orgName, err)
	}
	return response.PolicyPacks, nil
}

// GetRegistryPolicyPack reads the registry metadata for a single policy pack,
// resolved within the organization's own namespace so there is no name ambiguity
// to resolve client side. Returns (nil, nil) when the pack does not exist.
//
// Pass the pack's actual versionTag. Provenance does not vary by version, so it is
// tempting to omit it — but omitting is not "let the service pick the newest": the
// service substitutes the literal tag `latest`, and packs are tagged with their
// version (`2.6.4`), not with `latest`. Against a real org, every pack resolved
// when given its own tag and every one 404'd without it, so dropping the tag does
// not cost provenance occasionally — it costs it entirely.
//
// Pass "" only when no tag is known, and expect the lookup to miss.
func (c *Client) GetRegistryPolicyPack(
	ctx context.Context,
	orgName string,
	policyPackName string,
	versionTag string,
) (*RegistryPolicyPack, error) {
	if len(orgName) == 0 {
		return nil, errors.New("empty orgName")
	}
	if len(policyPackName) == 0 {
		return nil, errors.New("empty policyPackName")
	}

	apiPath := path.Join("orgs", orgName, "registry", "policypacks", policyPackName)

	var response getRegistryPolicyPackResponse
	var err error
	if versionTag != "" {
		_, err = c.doWithQuery(ctx, http.MethodGet, apiPath, url.Values{tagQueryParam: {versionTag}}, nil, &response)
	} else {
		_, err = c.do(ctx, http.MethodGet, apiPath, nil, &response)
	}
	if err != nil {
		if GetErrorStatusCode(err) == http.StatusNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get registry policy pack: %w", err)
	}

	return &response.PolicyPack, nil
}

// indexRegistryPolicyPacksByName keys registry entries on their bare pack name so
// they can be joined against the org policy pack list, which has no other key.
//
// The registry namespaces packs as {source}/{publisher}/{name}, so a bare name is
// not guaranteed unique. Names claimed by more than one publisher are omitted
// entirely: mis-attributing a pack to the wrong publisher is the same failure this
// metadata exists to prevent, and picking the first match would be
// nondeterministic since registry list order is unspecified.
func indexRegistryPolicyPacksByName(packs []RegistryPolicyPack) map[string]RegistryPolicyPack {
	seen := make(map[string]int, len(packs))
	for _, pack := range packs {
		seen[pack.Name]++
	}

	byName := make(map[string]RegistryPolicyPack, len(packs))
	for _, pack := range packs {
		if seen[pack.Name] == 1 {
			byName[pack.Name] = pack
		}
	}
	return byName
}

// ListPolicyPacksWithRegistryMetadata lists an organization's policy packs and
// annotates each with its registry provenance.
//
// The org list is the source of truth for which packs exist; the registry only
// enriches. Registry entries with no matching org pack are ignored.
//
// registryUnavailable reports that the backend does not serve the registry route
// (an older or self-hosted Pulumi Cloud), in which case every pack is returned
// unannotated and callers should warn rather than fail — getPolicyPacks worked on
// those backends before this feature existed and must keep working.
//
// Every other registry failure is returned as an error rather than degraded. A
// silently unannotated result is worse than no result: a program filtering on
// publisher would match nothing, attach zero policy packs, and still succeed.
func (c *Client) ListPolicyPacksWithRegistryMetadata(
	ctx context.Context,
	orgName string,
) (packs []PolicyPackWithRegistryMetadata, registryUnavailable bool, err error) {
	if len(orgName) == 0 {
		return nil, false, errors.New("empty orgName")
	}

	orgPacks, err := c.ListPolicyPacks(ctx, orgName)
	if err != nil {
		return nil, false, err
	}

	byName := map[string]RegistryPolicyPack{}
	registryPacks, err := c.ListRegistryPolicyPacks(ctx, orgName)
	if err != nil {
		switch GetErrorStatusCode(err) {
		case http.StatusNotFound, http.StatusMethodNotAllowed:
			registryUnavailable = true
		default:
			return nil, false, err
		}
	} else {
		byName = indexRegistryPolicyPacksByName(registryPacks)
	}

	packs = make([]PolicyPackWithRegistryMetadata, len(orgPacks))
	for i, orgPack := range orgPacks {
		packs[i] = PolicyPackWithRegistryMetadata{PolicyPackWithVersions: orgPack}
		if registryPack, ok := byName[orgPack.Name]; ok {
			packs[i].Source = registryPack.Source
			packs[i].Publisher = registryPack.Publisher
		}
	}
	return packs, registryUnavailable, nil
}
