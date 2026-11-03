// Copyright 2026, Pulumi Corporation.

package resources

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pulumi/pulumi-cloud-sdk/go/apitype"
	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/pulumi/pulumi/sdk/v3/go/property"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/config"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
)

const (
	eaOrg      = "my-org"
	eaUser     = "alice"
	eaStackID  = "stack-uuid"
	eaStackRef = "proj/dev"
	eaID       = "my-org/alice/stack/proj/dev"
)

type entityAccessClientMock struct {
	config.Client
	stack       *apitype.AppStack
	env         *pulumiapi.EnvironmentMetadata
	account     *pulumiapi.InsightsAccount
	listing     *apitype.EntityAssignmentsResponse
	updates     []apitype.UpdateEntityAssignmentRequest
	updateError error
}

func (c *entityAccessClientMock) GetStack(context.Context, pulumiapi.StackIdentifier) (*apitype.AppStack, error) {
	return c.stack, nil
}

func (c *entityAccessClientMock) GetEnvironmentMetadata(
	context.Context, string, string, string,
) (*pulumiapi.EnvironmentMetadata, error) {
	return c.env, nil
}

func (c *entityAccessClientMock) GetInsightsAccount(
	context.Context, string, string,
) (*pulumiapi.InsightsAccount, error) {
	return c.account, nil
}

func (c *entityAccessClientMock) ListEntityAssignments(
	context.Context, string, apitype.EntityInstance,
) (*apitype.EntityAssignmentsResponse, error) {
	return c.listing, nil
}

func (c *entityAccessClientMock) UpdateEntityAssignment(
	_ context.Context, _ string, req apitype.UpdateEntityAssignmentRequest,
) (*apitype.UpdateEntityAssignmentResponse, error) {
	c.updates = append(c.updates, req)
	return &apitype.UpdateEntityAssignmentResponse{}, c.updateError
}

func permissionSet(id, name, identifier string) apitype.PermissionSet {
	set := apitype.PermissionSet{ID: id, Name: name}
	if identifier != "" {
		set.DefaultIdentifier = &identifier
	}
	return set
}

func stackListing(assignments ...apitype.EntityAssignment) *apitype.EntityAssignmentsResponse {
	return &apitype.EntityAssignmentsResponse{
		Assignments: assignments,
		AvailablePermissionSets: []apitype.PermissionSet{
			permissionSet("read-id", "Read", "stack-read"),
			permissionSet("write-id", "Write", "stack-write"),
			permissionSet("custom-id", "Deployer", ""),
		},
	}
}

func userAssignment(login string, set apitype.PermissionSet) apitype.EntityAssignment {
	return apitype.EntityAssignment{
		Principal:     apitype.AssignmentPrincipal{User: &apitype.UserInfo{GitHubLogin: login}},
		PermissionSet: set,
		Source:        apitype.AssignmentSourceAssignment,
	}
}

func stackAccessInput(permission, setID string) OrganizationMemberEntityAccessInput {
	in := OrganizationMemberEntityAccessInput{
		OrganizationName: eaOrg, Username: eaUser, EntityType: "stack", EntityName: eaStackRef,
	}
	if permission != "" {
		in.Permission = &permission
	}
	if setID != "" {
		in.PermissionSetId = &setID
	}
	return in
}

func TestOrganizationMemberEntityAccessCheck(t *testing.T) {
	check := func(entityType, entityName string, extra map[string]property.Value) []string {
		props := map[string]property.Value{
			gcOrganizationName: property.New(eaOrg),
			"username":         property.New(eaUser),
			"entityType":       property.New(entityType),
			"entityName":       property.New(entityName),
		}
		for k, v := range extra {
			props[k] = v
		}
		resp, err := (&OrganizationMemberEntityAccess{}).Check(context.Background(), infer.CheckRequest{
			NewInputs: property.NewMap(props),
		})
		require.NoError(t, err)
		reasons := []string{}
		for _, f := range resp.Failures {
			reasons = append(reasons, f.Reason)
		}
		return reasons
	}
	perm := func(v string) map[string]property.Value {
		return map[string]property.Value{"permission": property.New(v)}
	}

	assert.Empty(t, check("stack", eaStackRef, perm("stack-write")))
	assert.Empty(t, check("environment", "proj/env", perm("environment-open")))
	assert.Empty(t, check("insightsAccount", "aws-prod", perm("insights-account-read")))
	assert.Empty(t, check("stack", eaStackRef, map[string]property.Value{"permissionSetId": property.New("custom-id")}))

	assert.NotEmpty(t, check("stack", eaStackRef, nil), "a permission is required")
	assert.NotEmpty(t, check("stack", eaStackRef, map[string]property.Value{
		"permission": property.New("stack-read"), "permissionSetId": property.New("custom-id"),
	}), "permission and permissionSetId are exclusive")
	assert.NotEmpty(t, check("stack", eaStackRef, perm("environment-read")), "permission must match the entity type")
	assert.NotEmpty(t, check("team", "x", perm("stack-read")), "unknown entity type")
	assert.NotEmpty(t, check("stack", "dev", perm("stack-read")), "stack needs project/stack")
	assert.NotEmpty(t, check("environment", "a/b/c", perm("environment-read")), "environment needs project/name")
}

func TestOrganizationMemberEntityAccessCreate(t *testing.T) {
	t.Run("resolves the stack and the built-in set to IDs", func(t *testing.T) {
		mock := &entityAccessClientMock{stack: &apitype.AppStack{ID: eaStackID}, listing: stackListing()}
		ctx := config.WithMockClient(context.Background(), mock)

		resp, err := (&OrganizationMemberEntityAccess{}).Create(ctx,
			infer.CreateRequest[entityAccessInput]{Inputs: stackAccessInput("stack-write", "")})
		require.NoError(t, err)
		assert.Equal(t, eaID, resp.ID)
		assert.Equal(t, eaStackID, resp.Output.EntityId)
		assert.Equal(t, "Write", resp.Output.PermissionSetName)
		require.Len(t, mock.updates, 1)
		assert.Equal(t, eaStackID, mock.updates[0].ID)
		assert.Equal(t, eaUser, mock.updates[0].UserID)
		assert.Equal(t, "write-id", mock.updates[0].PermissionSetID)
		assert.Equal(t, "stack", contextKind(mock.updates[0].Context))
	})

	t.Run("environment and Insights account resolve through their own lookups", func(t *testing.T) {
		mock := &entityAccessClientMock{
			env:     &pulumiapi.EnvironmentMetadata{ID: "env-uuid"},
			account: &pulumiapi.InsightsAccount{ID: "acct-id"},
			listing: &apitype.EntityAssignmentsResponse{AvailablePermissionSets: []apitype.PermissionSet{
				permissionSet("set-id", "Read", "environment-read"),
				permissionSet("acct-set-id", "Read", "insights-account-read"),
			}},
		}
		ctx := config.WithMockClient(context.Background(), mock)
		r := &OrganizationMemberEntityAccess{}

		_, err := r.Create(ctx, infer.CreateRequest[entityAccessInput]{Inputs: entityAccessInput{
			OrganizationName: eaOrg, Username: eaUser, EntityType: "environment", EntityName: "proj/env",
			Permission: eaPtr("environment-read"),
		}})
		require.NoError(t, err)
		_, err = r.Create(ctx, infer.CreateRequest[entityAccessInput]{Inputs: entityAccessInput{
			OrganizationName: eaOrg, Username: eaUser, EntityType: "insightsAccount", EntityName: "aws-prod",
			Permission: eaPtr("insights-account-read"),
		}})
		require.NoError(t, err)
		require.Len(t, mock.updates, 2)
		assert.Equal(t, "env-uuid", mock.updates[0].ID)
		assert.Equal(t, "environment", contextKind(mock.updates[0].Context))
		assert.Equal(t, "acct-id", mock.updates[1].ID)
		assert.Equal(t, "insightsAccount", contextKind(mock.updates[1].Context))
	})

	t.Run("custom permission set ID", func(t *testing.T) {
		mock := &entityAccessClientMock{stack: &apitype.AppStack{ID: eaStackID}, listing: stackListing()}
		ctx := config.WithMockClient(context.Background(), mock)
		_, err := (&OrganizationMemberEntityAccess{}).Create(ctx,
			infer.CreateRequest[entityAccessInput]{Inputs: stackAccessInput("", "custom-id")})
		require.NoError(t, err)
		assert.Equal(t, "custom-id", mock.updates[0].PermissionSetID)
	})

	t.Run("unknown permission and unknown set ID write nothing", func(t *testing.T) {
		mock := &entityAccessClientMock{stack: &apitype.AppStack{ID: eaStackID}, listing: stackListing()}
		ctx := config.WithMockClient(context.Background(), mock)
		r := &OrganizationMemberEntityAccess{}

		_, err := r.Create(ctx, infer.CreateRequest[entityAccessInput]{Inputs: stackAccessInput("stack-admin", "")})
		assert.ErrorContains(t, err, "stack-read, stack-write")
		_, err = r.Create(ctx, infer.CreateRequest[entityAccessInput]{Inputs: stackAccessInput("", "nope")})
		assert.Error(t, err)
		assert.Empty(t, mock.updates)
	})

	t.Run("missing stack is an error", func(t *testing.T) {
		ctx := config.WithMockClient(context.Background(), &entityAccessClientMock{})
		_, err := (&OrganizationMemberEntityAccess{}).Create(ctx,
			infer.CreateRequest[entityAccessInput]{Inputs: stackAccessInput("stack-read", "")})
		assert.ErrorContains(t, err, "not found")
	})

	t.Run("preview makes no call", func(t *testing.T) {
		mock := &entityAccessClientMock{}
		ctx := config.WithMockClient(context.Background(), mock)
		resp, err := (&OrganizationMemberEntityAccess{}).Create(ctx,
			infer.CreateRequest[entityAccessInput]{DryRun: true, Inputs: stackAccessInput("stack-read", "")})
		require.NoError(t, err)
		assert.Equal(t, eaID, resp.ID)
		assert.Empty(t, mock.updates)
	})
}

func TestOrganizationMemberEntityAccessUpdate(t *testing.T) {
	mock := &entityAccessClientMock{stack: &apitype.AppStack{ID: eaStackID}, listing: stackListing()}
	ctx := config.WithMockClient(context.Background(), mock)
	resp, err := (&OrganizationMemberEntityAccess{}).Update(ctx,
		infer.UpdateRequest[entityAccessInput, entityAccessState]{
			ID:     eaID,
			Inputs: stackAccessInput("stack-write", ""),
			State:  entityAccessState{OrganizationMemberEntityAccessInput: stackAccessInput("stack-read", "")},
		})
	require.NoError(t, err)
	assert.Equal(t, "Write", resp.Output.PermissionSetName)
	require.Len(t, mock.updates, 1)
	assert.Equal(t, "write-id", mock.updates[0].PermissionSetID)
}

func TestOrganizationMemberEntityAccessRead(t *testing.T) {
	read := func(mock *entityAccessClientMock, inputs OrganizationMemberEntityAccessInput) (
		infer.ReadResponse[entityAccessInput, entityAccessState], error,
	) {
		return (&OrganizationMemberEntityAccess{}).Read(config.WithMockClient(context.Background(), mock),
			infer.ReadRequest[entityAccessInput, entityAccessState]{
				ID: eaID, Inputs: inputs,
			})
	}
	stack := &apitype.AppStack{ID: eaStackID}

	t.Run("reports this member's grant and ignores other members and teams", func(t *testing.T) {
		team := apitype.EntityAssignment{
			Principal:     apitype.AssignmentPrincipal{Team: &apitype.AssignmentTeam{Name: eaUser}},
			PermissionSet: permissionSet("read-id", "Read", "stack-read"),
			Source:        apitype.AssignmentSourceTeam,
		}
		mock := &entityAccessClientMock{stack: stack, listing: stackListing(
			team,
			userAssignment("bob", permissionSet("read-id", "Read", "stack-read")),
			userAssignment("Alice", permissionSet("write-id", "Write", "stack-write")),
		)}
		resp, err := read(mock, entityAccessInput{})
		require.NoError(t, err)
		assert.Equal(t, eaID, resp.ID)
		assert.Equal(t, "stack-write", *resp.Inputs.Permission)
		assert.Nil(t, resp.Inputs.PermissionSetId)
		assert.Equal(t, "proj/dev", resp.Inputs.EntityName)
		assert.Equal(t, "Write", resp.State.PermissionSetName)
	})

	t.Run("custom set is reported by ID", func(t *testing.T) {
		mock := &entityAccessClientMock{stack: stack, listing: stackListing(
			userAssignment(eaUser, permissionSet("custom-id", "Deployer", "")))}
		resp, err := read(mock, entityAccessInput{})
		require.NoError(t, err)
		assert.Equal(t, "custom-id", *resp.Inputs.PermissionSetId)
		assert.Nil(t, resp.Inputs.Permission)
	})

	t.Run("set declared by ID keeps permissionSetId", func(t *testing.T) {
		mock := &entityAccessClientMock{stack: stack, listing: stackListing(
			userAssignment(eaUser, permissionSet("read-id", "Read", "stack-read")))}
		resp, err := read(mock, stackAccessInput("", "read-id"))
		require.NoError(t, err)
		assert.Equal(t, "read-id", *resp.Inputs.PermissionSetId)
		assert.Nil(t, resp.Inputs.Permission)
	})

	t.Run("removed grant and missing entity report the resource gone", func(t *testing.T) {
		resp, err := read(&entityAccessClientMock{stack: stack, listing: stackListing()}, entityAccessInput{})
		require.NoError(t, err)
		assert.Empty(t, resp.ID)
		resp, err = read(&entityAccessClientMock{}, entityAccessInput{})
		require.NoError(t, err)
		assert.Empty(t, resp.ID)
	})

	t.Run("malformed import ID", func(t *testing.T) {
		ctx := config.WithMockClient(context.Background(), &entityAccessClientMock{})
		_, err := (&OrganizationMemberEntityAccess{}).Read(ctx,
			infer.ReadRequest[entityAccessInput, entityAccessState]{ID: "my-org/alice"})
		assert.ErrorContains(t, err, "invalid ID")
	})
}

func TestOrganizationMemberEntityAccessDelete(t *testing.T) {
	state := OrganizationMemberEntityAccessState{
		OrganizationMemberEntityAccessInput: stackAccessInput("stack-read", ""),
		EntityId:                            eaStackID,
	}
	del := func(mock *entityAccessClientMock) error {
		_, err := (&OrganizationMemberEntityAccess{}).Delete(config.WithMockClient(context.Background(), mock),
			infer.DeleteRequest[entityAccessState]{State: state})
		return err
	}

	t.Run("revokes only this member's grant on this entity", func(t *testing.T) {
		mock := &entityAccessClientMock{}
		require.NoError(t, del(mock))
		require.Len(t, mock.updates, 1)
		assert.Equal(t, eaStackID, mock.updates[0].ID)
		assert.Equal(t, eaUser, mock.updates[0].UserID)
		assert.Empty(t, mock.updates[0].PermissionSetID)
	})

	t.Run("already revoked is not an error", func(t *testing.T) {
		mock := &entityAccessClientMock{updateError: &pulumiapi.ErrorResponse{StatusCode: http.StatusNotFound}}
		assert.NoError(t, del(mock))
	})
}

func eaPtr(s string) *string { return &s }

func contextKind(ctx apitype.PermissionContextExpression) string {
	switch ctx.(type) {
	case apitype.PermissionExpressionStack:
		return "stack"
	case apitype.PermissionExpressionEnvironment:
		return "environment"
	case apitype.PermissionExpressionInsightsAccount:
		return "insightsAccount"
	default:
		return "unknown"
	}
}
