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
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/pulumi/pulumi-cloud-sdk/go/apitype"
	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/config"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/util"
)

const (
	entityTypeStack           = "stack"
	entityTypeEnvironment     = "environment"
	entityTypeInsightsAccount = "insightsAccount"

	permissionProperty = "permission"
)

var permissionPrefixes = map[string]string{
	entityTypeStack:           "stack-",
	entityTypeEnvironment:     "environment-",
	entityTypeInsightsAccount: "insights-account-",
}

type OrganizationMemberEntityAccess struct{}

type (
	entityAccessInput = OrganizationMemberEntityAccessInput
	entityAccessState = OrganizationMemberEntityAccessState
)

var (
	_ infer.CustomCreate[entityAccessInput, entityAccessState] = &OrganizationMemberEntityAccess{}
	_ infer.CustomCheck[entityAccessInput]                     = &OrganizationMemberEntityAccess{}
	_ infer.CustomDelete[entityAccessState]                    = &OrganizationMemberEntityAccess{}
	_ infer.CustomRead[entityAccessInput, entityAccessState]   = &OrganizationMemberEntityAccess{}
	_ infer.CustomUpdate[entityAccessInput, entityAccessState] = &OrganizationMemberEntityAccess{}
)

func (*OrganizationMemberEntityAccess) Annotate(a infer.Annotator) {
	a.Describe(
		&OrganizationMemberEntityAccess{},
		"Grants one organization member a permission set on one stack, environment or Insights account, as in "+
			"the entity access section of the Pulumi Cloud console. See the "+
			"[entity access docs](https://www.pulumi.com/docs/administration/concepts/rbac/entities/). "+
			"Each resource manages the grant of one member on one entity. Other grants of the member, and "+
			"grants of other members on the same entity, are not touched, and destroying the resource "+
			"removes only this grant. Set either `permission`, for a built-in permission set, or "+
			"`permissionSetId`, for a custom one."+
			util.ImportDocs(
				"pulumiservice:index:OrganizationMemberEntityAccess",
				"{organization}/{username}/{entityType}/{entityName}",
				"my-org/alice/stack/my-project/dev",
				"`entityType` is `stack`, `environment` or `insightsAccount`. `entityName` is "+
					"`{project}/{stack}`, `{project}/{environment}` or the account name.",
			),
	)
}

type OrganizationMemberEntityAccessInput struct {
	OrganizationName string  `pulumi:"organizationName"          provider:"replaceOnChanges"`
	Username         string  `pulumi:"username"                  provider:"replaceOnChanges"`
	EntityType       string  `pulumi:"entityType"                provider:"replaceOnChanges"`
	EntityName       string  `pulumi:"entityName"                provider:"replaceOnChanges"`
	Permission       *string `pulumi:"permission,optional"`
	PermissionSetId  *string `pulumi:"permissionSetId,optional"`
}

func (i *OrganizationMemberEntityAccessInput) Annotate(a infer.Annotator) {
	a.Describe(&i.OrganizationName, "The Pulumi Cloud organization name.")
	a.Describe(&i.Username, "The Pulumi Cloud username of the organization member.")
	a.Describe(&i.EntityType, "The kind of entity to grant access to. One of `stack`, `environment` or `insightsAccount`.")
	a.Describe(
		&i.EntityName,
		"The name of the entity: `{project}/{stack}` for a stack, `{project}/{environment}` for an "+
			"environment, or the account name for an Insights account.",
	)
	a.Describe(
		&i.Permission,
		"The built-in permission set to grant, for example `stack-read`, `stack-write`, `stack-admin`, "+
			"`environment-read`, `environment-open`, `environment-write`, `environment-admin`, "+
			"`insights-account-read`, `insights-account-write` or `insights-account-admin`. "+
			"Exactly one of `permission` and `permissionSetId` must be set.",
	)
	a.Describe(
		&i.PermissionSetId,
		"The ID of a custom permission set to grant. Exactly one of `permission` and `permissionSetId` must be set.",
	)
}

type OrganizationMemberEntityAccessState struct {
	OrganizationMemberEntityAccessInput
	EntityId          string `pulumi:"entityId"`
	PermissionSetName string `pulumi:"permissionSetName"`
}

func (s *OrganizationMemberEntityAccessState) Annotate(a infer.Annotator) {
	a.Describe(&s.EntityId, "The ID of the entity Pulumi Cloud stores the grant against.")
	a.Describe(&s.PermissionSetName, "The display name of the granted permission set.")
}

func (*OrganizationMemberEntityAccess) Check(
	ctx context.Context,
	req infer.CheckRequest,
) (infer.CheckResponse[OrganizationMemberEntityAccessInput], error) {
	in, failures, err := infer.DefaultCheck[OrganizationMemberEntityAccessInput](ctx, req.NewInputs)
	if err != nil {
		return infer.CheckResponse[OrganizationMemberEntityAccessInput]{}, err
	}
	fail := func(property, reason string) {
		failures = append(failures, p.CheckFailure{Property: property, Reason: reason})
	}

	prefix, knownType := permissionPrefixes[in.EntityType]
	switch {
	case in.EntityType == "":
	case !knownType:
		fail("entityType", fmt.Sprintf("entityType must be one of %s, %s or %s, got %q",
			entityTypeStack, entityTypeEnvironment, entityTypeInsightsAccount, in.EntityType))
	case in.EntityType != entityTypeInsightsAccount && in.EntityName != "":
		if _, _, err := splitProjectEntity(in.EntityName); err != nil {
			fail("entityName", err.Error())
		}
	}

	hasPermission := util.OrZero(in.Permission) != ""
	hasSetID := util.OrZero(in.PermissionSetId) != ""
	switch {
	case hasPermission == hasSetID:
		fail(permissionProperty, "exactly one of permission and permissionSetId must be set")
	case hasPermission && knownType && !strings.HasPrefix(*in.Permission, prefix):
		fail(permissionProperty, fmt.Sprintf("permission for a %s must start with %q, got %q",
			in.EntityType, prefix, *in.Permission))
	}
	return infer.CheckResponse[OrganizationMemberEntityAccessInput]{Inputs: in, Failures: failures}, nil
}

func (*OrganizationMemberEntityAccess) Create(
	ctx context.Context,
	req infer.CreateRequest[OrganizationMemberEntityAccessInput],
) (infer.CreateResponse[OrganizationMemberEntityAccessState], error) {
	type response = infer.CreateResponse[OrganizationMemberEntityAccessState]
	id := entityAccessID(req.Inputs)
	if req.DryRun {
		return response{
			ID:     id,
			Output: OrganizationMemberEntityAccessState{OrganizationMemberEntityAccessInput: req.Inputs},
		}, nil
	}
	state, err := applyEntityAccess(ctx, req.Inputs)
	if err != nil {
		return response{}, err
	}
	return response{ID: id, Output: state}, nil
}

func (*OrganizationMemberEntityAccess) Update(
	ctx context.Context,
	req infer.UpdateRequest[OrganizationMemberEntityAccessInput, OrganizationMemberEntityAccessState],
) (infer.UpdateResponse[OrganizationMemberEntityAccessState], error) {
	type response = infer.UpdateResponse[OrganizationMemberEntityAccessState]
	if req.DryRun {
		return response{Output: OrganizationMemberEntityAccessState{
			OrganizationMemberEntityAccessInput: req.Inputs,
			EntityId:                            req.State.EntityId,
		}}, nil
	}
	state, err := applyEntityAccess(ctx, req.Inputs)
	if err != nil {
		return response{}, err
	}
	return response{Output: state}, nil
}

func applyEntityAccess(
	ctx context.Context,
	in OrganizationMemberEntityAccessInput,
) (OrganizationMemberEntityAccessState, error) {
	client := config.GetClient(ctx)
	entity, err := resolveAssignmentEntity(ctx, client, in.OrganizationName, in.EntityType, in.EntityName)
	if err != nil {
		return OrganizationMemberEntityAccessState{}, err
	}
	if entity == nil {
		return OrganizationMemberEntityAccessState{}, fmt.Errorf(
			"%s %q not found in organization %q", in.EntityType, in.EntityName, in.OrganizationName)
	}
	listing, err := client.ListEntityAssignments(ctx, in.OrganizationName, *entity)
	if err != nil {
		return OrganizationMemberEntityAccessState{}, err
	}
	if listing == nil {
		return OrganizationMemberEntityAccessState{}, fmt.Errorf(
			"%s %q not found in organization %q", in.EntityType, in.EntityName, in.OrganizationName)
	}
	set, err := findPermissionSet(listing.AvailablePermissionSets, in)
	if err != nil {
		return OrganizationMemberEntityAccessState{}, err
	}

	if _, err := client.UpdateEntityAssignment(ctx, in.OrganizationName, apitype.UpdateEntityAssignmentRequest{
		EntityInstance:  *entity,
		UserID:          in.Username,
		PermissionSetID: set.ID,
	}); err != nil {
		return OrganizationMemberEntityAccessState{}, err
	}
	return OrganizationMemberEntityAccessState{
		OrganizationMemberEntityAccessInput: in,
		EntityId:                            entity.ID,
		PermissionSetName:                   set.Name,
	}, nil
}

func findPermissionSet(
	available []apitype.PermissionSet,
	in OrganizationMemberEntityAccessInput,
) (apitype.PermissionSet, error) {
	if id := util.OrZero(in.PermissionSetId); id != "" {
		for _, s := range available {
			if s.ID == id {
				return s, nil
			}
		}
		return apitype.PermissionSet{}, fmt.Errorf("permission set ID %q cannot be granted on a %s", id, in.EntityType)
	}
	identifiers := []string{}
	for _, s := range available {
		if s.DefaultIdentifier == nil {
			continue
		}
		if *s.DefaultIdentifier == util.OrZero(in.Permission) {
			return s, nil
		}
		identifiers = append(identifiers, *s.DefaultIdentifier)
	}
	return apitype.PermissionSet{}, fmt.Errorf(
		"permission %q is not available on a %s; available built-in permissions: %s",
		util.OrZero(in.Permission), in.EntityType, strings.Join(identifiers, ", "))
}

// Delete revokes only this member's grant. A missing grant or entity counts as already revoked.
func (*OrganizationMemberEntityAccess) Delete(
	ctx context.Context,
	req infer.DeleteRequest[OrganizationMemberEntityAccessState],
) (infer.DeleteResponse, error) {
	if req.State.EntityId == "" {
		return infer.DeleteResponse{}, nil
	}
	entity, err := entityInstance(req.State.EntityType, req.State.EntityId)
	if err != nil {
		return infer.DeleteResponse{}, err
	}
	_, err = config.GetClient(ctx).UpdateEntityAssignment(ctx, req.State.OrganizationName,
		apitype.UpdateEntityAssignmentRequest{EntityInstance: entity, UserID: req.State.Username})
	if err != nil && pulumiapi.GetErrorStatusCode(err) != http.StatusNotFound {
		return infer.DeleteResponse{}, err
	}
	return infer.DeleteResponse{}, nil
}

func (*OrganizationMemberEntityAccess) Read(
	ctx context.Context,
	req infer.ReadRequest[OrganizationMemberEntityAccessInput, OrganizationMemberEntityAccessState],
) (infer.ReadResponse[OrganizationMemberEntityAccessInput, OrganizationMemberEntityAccessState], error) {
	type response = infer.ReadResponse[OrganizationMemberEntityAccessInput, OrganizationMemberEntityAccessState]
	org, username, entityType, entityName, err := parseEntityAccessID(req.ID)
	if err != nil {
		return response{}, err
	}

	client := config.GetClient(ctx)
	entity, err := resolveAssignmentEntity(ctx, client, org, entityType, entityName)
	if err != nil {
		return response{}, err
	}
	if entity == nil {
		return response{}, nil
	}
	listing, err := client.ListEntityAssignments(ctx, org, *entity)
	if err != nil {
		return response{}, err
	}
	if listing == nil {
		return response{}, nil
	}
	idx := slices.IndexFunc(listing.Assignments, func(a apitype.EntityAssignment) bool {
		return a.Source == apitype.AssignmentSourceAssignment &&
			a.Principal.User != nil &&
			strings.EqualFold(a.Principal.User.GitHubLogin, username)
	})
	if idx < 0 {
		return response{}, nil
	}
	set := listing.Assignments[idx].PermissionSet

	in := OrganizationMemberEntityAccessInput{
		OrganizationName: org,
		Username:         username,
		EntityType:       entityType,
		EntityName:       entityName,
	}
	declaredByID := util.OrZero(req.Inputs.PermissionSetId) == set.ID
	if set.DefaultIdentifier != nil && !declaredByID {
		in.Permission = set.DefaultIdentifier
	} else {
		in.PermissionSetId = &set.ID
	}
	return response{
		ID:     req.ID,
		Inputs: in,
		State: OrganizationMemberEntityAccessState{
			OrganizationMemberEntityAccessInput: in,
			EntityId:                            entity.ID,
			PermissionSetName:                   set.Name,
		},
	}, nil
}

func entityAccessID(in OrganizationMemberEntityAccessInput) string {
	return strings.Join([]string{in.OrganizationName, in.Username, in.EntityType, in.EntityName}, "/")
}

func parseEntityAccessID(id string) (org, username, entityType, entityName string, err error) {
	parts := strings.SplitN(id, "/", 4)
	if len(parts) != 4 || slices.Contains(parts, "") {
		return "", "", "", "", fmt.Errorf(
			"invalid ID %q: expected {organization}/{username}/{entityType}/{entityName}", id)
	}
	return parts[0], parts[1], parts[2], parts[3], nil
}

func splitProjectEntity(name string) (project, entity string, err error) {
	parts := strings.Split(name, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("entityName must have the form {project}/{name}, got %q", name)
	}
	return parts[0], parts[1], nil
}

func entityInstance(entityType, id string) (apitype.EntityInstance, error) {
	switch entityType {
	case entityTypeStack:
		return apitype.EntityInstance{Context: apitype.PermissionExpressionStackBuilder{}.Build(), ID: id}, nil
	case entityTypeEnvironment:
		return apitype.EntityInstance{Context: apitype.PermissionExpressionEnvironmentBuilder{}.Build(), ID: id}, nil
	case entityTypeInsightsAccount:
		return apitype.EntityInstance{Context: apitype.PermissionExpressionInsightsAccountBuilder{}.Build(), ID: id}, nil
	default:
		return apitype.EntityInstance{}, fmt.Errorf("unsupported entity type %q", entityType)
	}
}

func resolveAssignmentEntity(
	ctx context.Context,
	client config.Client,
	org, entityType, entityName string,
) (*apitype.EntityInstance, error) {
	var id string
	switch entityType {
	case entityTypeStack:
		project, stack, err := splitProjectEntity(entityName)
		if err != nil {
			return nil, err
		}
		s, err := client.GetStack(ctx, pulumiapi.StackIdentifier{OrgName: org, ProjectName: project, StackName: stack})
		if err != nil || s == nil {
			return nil, err
		}
		id = s.ID
	case entityTypeEnvironment:
		project, env, err := splitProjectEntity(entityName)
		if err != nil {
			return nil, err
		}
		meta, err := client.GetEnvironmentMetadata(ctx, org, project, env)
		if err != nil || meta == nil {
			return nil, err
		}
		id = meta.ID
	case entityTypeInsightsAccount:
		account, err := client.GetInsightsAccount(ctx, org, entityName)
		if err != nil || account == nil {
			return nil, err
		}
		id = account.ID
	default:
		return nil, fmt.Errorf("unsupported entity type %q", entityType)
	}
	if id == "" {
		return nil, errors.New("Pulumi Cloud returned no ID for " + entityType + " " + entityName)
	}
	entity, err := entityInstance(entityType, id)
	if err != nil {
		return nil, err
	}
	return &entity, nil
}
