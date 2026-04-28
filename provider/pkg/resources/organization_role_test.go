package resources

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/pulumi/pulumi/sdk/v3/go/property"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/config"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
)

type orgRoleClientMock struct {
	config.Client
	create func(ctx context.Context, org string, req pulumiapi.CreateRoleRequest) (*pulumiapi.RoleDescriptor, error)
	get    func(ctx context.Context, org, id string) (*pulumiapi.RoleDescriptor, error)
	update func(
		ctx context.Context, org, id string, name, desc *string, details json.RawMessage,
	) (*pulumiapi.RoleDescriptor, error)
	del func(ctx context.Context, org, id string, force bool) error
}

func (c *orgRoleClientMock) CreateRole(
	ctx context.Context, org string, req pulumiapi.CreateRoleRequest,
) (*pulumiapi.RoleDescriptor, error) {
	return c.create(ctx, org, req)
}

func (c *orgRoleClientMock) GetRole(
	ctx context.Context, org, id string,
) (*pulumiapi.RoleDescriptor, error) {
	return c.get(ctx, org, id)
}

func (c *orgRoleClientMock) UpdateRole(
	ctx context.Context, org, id string, name, desc *string, details json.RawMessage,
) (*pulumiapi.RoleDescriptor, error) {
	return c.update(ctx, org, id, name, desc, details)
}

func (c *orgRoleClientMock) DeleteRole(ctx context.Context, org, id string, force bool) error {
	return c.del(ctx, org, id, force)
}

var testPermissions = map[string]interface{}{
	"kind":        "allow",
	"permissions": []interface{}{"stack:read"},
}

func TestOrganizationRoleCreate(t *testing.T) {
	mock := &orgRoleClientMock{
		create: func(_ context.Context, org string, req pulumiapi.CreateRoleRequest) (*pulumiapi.RoleDescriptor, error) {
			assert.Equal(t, "acme", org)
			assert.Equal(t, "read-only", req.Name)
			// Defaults are applied by the API client layer, so at this seam
			// ResourceType stays empty when the user didn't set it.
			assert.Equal(t, "", req.ResourceType)
			// details should be the JSON-encoded permissions map.
			var parsed map[string]interface{}
			assert.NoError(t, json.Unmarshal(req.Details, &parsed))
			assert.Equal(t, "PermissionDescriptorAllow", parsed["__type"])
			assert.NotContains(t, parsed, "kind", "wire body must not leak `kind` to the API")
			return &pulumiapi.RoleDescriptor{
				ID:      "role-123",
				Name:    req.Name,
				Version: 1,
				Details: req.Details,
			}, nil
		},
	}
	ctx := config.WithMockClient(context.Background(), mock)

	r := &OrganizationRole{}
	resp, err := r.Create(ctx, infer.CreateRequest[OrganizationRoleInput]{
		Inputs: OrganizationRoleInput{
			OrganizationRoleCore: OrganizationRoleCore{
				OrganizationName: "acme",
				Name:             "read-only",
				Permissions:      testPermissions,
			},
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, "acme/role-123", resp.ID)
	assert.Equal(t, "role-123", resp.Output.RoleId)
}

func TestOrganizationRoleRead(t *testing.T) {
	t.Run("not found", func(t *testing.T) {
		mock := &orgRoleClientMock{
			get: func(_ context.Context, _, _ string) (*pulumiapi.RoleDescriptor, error) { return nil, nil },
		}
		ctx := config.WithMockClient(context.Background(), mock)
		r := &OrganizationRole{}
		resp, err := r.Read(ctx, infer.ReadRequest[OrganizationRoleInput, OrganizationRoleState]{
			ID: "acme/role-123",
		})
		assert.NoError(t, err)
		assert.Equal(t, "", resp.ID)
	})

	t.Run("found parses details", func(t *testing.T) {
		// The API returns wire format with __type; the provider must translate to kind.
		wirePermissions := map[string]interface{}{
			"__type":      "PermissionDescriptorAllow",
			"permissions": []interface{}{"stack:read"},
		}
		raw, _ := json.Marshal(wirePermissions)
		mock := &orgRoleClientMock{
			get: func(_ context.Context, _, _ string) (*pulumiapi.RoleDescriptor, error) {
				return &pulumiapi.RoleDescriptor{
					ID:           "role-123",
					Name:         "read-only",
					Description:  "ro",
					ResourceType: "global",
					Version:      2,
					Details:      raw,
				}, nil
			},
		}
		ctx := config.WithMockClient(context.Background(), mock)
		r := &OrganizationRole{}
		resp, err := r.Read(ctx, infer.ReadRequest[OrganizationRoleInput, OrganizationRoleState]{
			ID: "acme/role-123",
		})
		assert.NoError(t, err)
		assert.Equal(t, "acme/role-123", resp.ID)
		// Read translates the wire __type back into the user-facing kind.
		assert.Equal(t, "allow", resp.State.Permissions["kind"])
		assert.NotContains(t, resp.State.Permissions, "__type", "state must not leak `__type` to the SDK")
	})
}

func TestOrganizationRoleDelete(t *testing.T) {
	called := false
	mock := &orgRoleClientMock{
		del: func(_ context.Context, org, id string, force bool) error {
			called = true
			assert.Equal(t, "acme", org)
			assert.Equal(t, "role-123", id)
			assert.True(t, force)
			return nil
		},
	}
	ctx := config.WithMockClient(context.Background(), mock)
	r := &OrganizationRole{}
	_, err := r.Delete(ctx, infer.DeleteRequest[OrganizationRoleState]{
		State: OrganizationRoleState{
			OrganizationRoleCore: OrganizationRoleCore{OrganizationName: "acme"},
			RoleId:               "role-123",
		},
	})
	assert.NoError(t, err)
	assert.True(t, called)
}

func TestOrganizationRoleUpdateOmitsDescriptionWhenUnset(t *testing.T) {
	// Regression: Update used to pass `&""` (from util.OrZero) as the description
	// pointer when the user had not set one. With `omitempty` on *string only
	// eliding nil (not empty), the PATCH body included `"description": ""` and
	// cleared any existing description on the server.
	raw, _ := json.Marshal(testPermissions)
	var gotDesc *string
	gotDesc = new(string) // sentinel so we can distinguish "passed nil" from "test hasn't run"
	*gotDesc = "__sentinel__"
	mock := &orgRoleClientMock{
		update: func(
			_ context.Context, _, _ string, _, desc *string, _ json.RawMessage,
		) (*pulumiapi.RoleDescriptor, error) {
			gotDesc = desc
			return &pulumiapi.RoleDescriptor{ID: "role-123", Name: "read-only", Version: 3, Details: raw}, nil
		},
	}
	ctx := config.WithMockClient(context.Background(), mock)
	r := &OrganizationRole{}
	_, err := r.Update(ctx, infer.UpdateRequest[OrganizationRoleInput, OrganizationRoleState]{
		Inputs: OrganizationRoleInput{
			OrganizationRoleCore: OrganizationRoleCore{
				OrganizationName: "acme",
				Name:             "read-only",
				// Description intentionally unset.
				Permissions: testPermissions,
			},
		},
		State: OrganizationRoleState{
			OrganizationRoleCore: OrganizationRoleCore{OrganizationName: "acme"},
			RoleId:               "role-123",
		},
	})
	assert.NoError(t, err)
	assert.Nil(t, gotDesc, "Update must pass nil description when the user did not set one")
}

func TestOrganizationRoleCheck(t *testing.T) {
	r := &OrganizationRole{}

	t.Run("rejects empty permissions", func(t *testing.T) {
		resp, err := r.Check(context.Background(), infer.CheckRequest{
			NewInputs: property.NewMap(map[string]property.Value{
				"organizationName": property.New("acme"),
				"name":             property.New("r"),
			}),
		})
		assert.NoError(t, err)
		props := map[string]bool{}
		for _, f := range resp.Failures {
			props[f.Property] = true
		}
		assert.True(t, props["permissions"])
	})

	// At preview, `permissions` (or `name`) may arrive as Computed when wired
	// to another resource's output — e.g. `permissions =
	// buildEnvironmentScopedPermissionsOutput({...}).permissions`. The empty
	// check would otherwise fire on the zero-value decoded Go field and break
	// every fresh `pulumi preview`. The same emptiness checks belong in
	// Create/Update, where Pulumi guarantees concrete values.
	t.Run("tolerates computed permissions at preview", func(t *testing.T) {
		resp, err := r.Check(context.Background(), infer.CheckRequest{
			NewInputs: property.NewMap(map[string]property.Value{
				"organizationName": property.New("acme"),
				"name":             property.New("r"),
				"permissions":      property.New(property.Computed),
			}),
		})
		assert.NoError(t, err)
		for _, f := range resp.Failures {
			assert.NotEqual(t, "permissions", f.Property,
				"Check must not reject a computed permissions input: %s", f.Reason)
		}
	})

	t.Run("tolerates computed name at preview", func(t *testing.T) {
		resp, err := r.Check(context.Background(), infer.CheckRequest{
			NewInputs: property.NewMap(map[string]property.Value{
				"organizationName": property.New("acme"),
				"name":             property.New(property.Computed),
				"permissions": property.New(property.NewMap(map[string]property.Value{
					"kind": property.New("allow"),
				})),
			}),
		})
		assert.NoError(t, err)
		for _, f := range resp.Failures {
			assert.NotEqual(t, "name", f.Property,
				"Check must not reject a computed name input: %s", f.Reason)
		}
	})

	t.Run("rejects unknown kind", func(t *testing.T) {
		resp, err := r.Check(context.Background(), infer.CheckRequest{
			NewInputs: property.NewMap(map[string]property.Value{
				"organizationName": property.New("acme"),
				"name":             property.New("r"),
				"permissions": property.New(property.NewMap(map[string]property.Value{
					"kind": property.New("totallyMadeUp"),
				})),
			}),
		})
		assert.NoError(t, err)
		props := map[string]string{}
		for _, f := range resp.Failures {
			props[f.Property] = f.Reason
		}
		assert.Contains(t, props["permissions"], "totallyMadeUp",
			"Check must reject unknown kind values upfront")
	})

	t.Run("rejects empty on map", func(t *testing.T) {
		resp, err := r.Check(context.Background(), infer.CheckRequest{
			NewInputs: property.NewMap(map[string]property.Value{
				"organizationName": property.New("acme"),
				"name":             property.New("r"),
				"permissions": property.New(property.NewMap(map[string]property.Value{
					"kind":        property.New("allow"),
					"on":          property.New(property.NewMap(map[string]property.Value{})),
					"permissions": property.New(property.NewArray([]property.Value{property.New("stack:read")})),
				})),
			}),
		})
		assert.NoError(t, err)
		props := map[string]string{}
		for _, f := range resp.Failures {
			props[f.Property] = f.Reason
		}
		assert.Contains(t, props["permissions"], "on",
			"empty on map should produce a permissions failure mentioning on")
	})

	t.Run("rejects multi-key on map", func(t *testing.T) {
		resp, err := r.Check(context.Background(), infer.CheckRequest{
			NewInputs: property.NewMap(map[string]property.Value{
				"organizationName": property.New("acme"),
				"name":             property.New("r"),
				"permissions": property.New(property.NewMap(map[string]property.Value{
					"kind": property.New("allow"),
					"on": property.New(property.NewMap(map[string]property.Value{
						"environment": property.New("e"),
						"stack":       property.New("s"),
					})),
					"permissions": property.New(property.NewArray([]property.Value{property.New("stack:read")})),
				})),
			}),
		})
		assert.NoError(t, err)
		props := map[string]string{}
		for _, f := range resp.Failures {
			props[f.Property] = f.Reason
		}
		assert.Contains(t, props["permissions"], "on",
			"multi-key on should produce a permissions failure mentioning on")
	})

	t.Run("rejects unknown on entity type", func(t *testing.T) {
		resp, err := r.Check(context.Background(), infer.CheckRequest{
			NewInputs: property.NewMap(map[string]property.Value{
				"organizationName": property.New("acme"),
				"name":             property.New("r"),
				"permissions": property.New(property.NewMap(map[string]property.Value{
					"kind": property.New("allow"),
					"on": property.New(property.NewMap(map[string]property.Value{
						"unknownEntity": property.New("x"),
					})),
					"permissions": property.New(property.NewArray([]property.Value{property.New("stack:read")})),
				})),
			}),
		})
		assert.NoError(t, err)
		props := map[string]string{}
		for _, f := range resp.Failures {
			props[f.Property] = f.Reason
		}
		assert.Contains(t, props["permissions"], "unknownEntity",
			"unknown on entity should produce a permissions failure naming the bad key")
	})
}
