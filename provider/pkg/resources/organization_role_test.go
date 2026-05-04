package resources

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/pulumi/pulumi/sdk/v3/go/property"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/apitype"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/config"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
)

type orgRoleClientMock struct {
	config.Client
	create func(ctx context.Context, org string, req apitype.PermissionDescriptorBase) (*apitype.PermissionDescriptorRecord, error)
	get    func(ctx context.Context, org, id string) (*apitype.PermissionDescriptorRecord, error)
	update func(
		ctx context.Context, org, id string, name, desc *string, details apitype.PermissionDescriptor,
	) (*apitype.PermissionDescriptorRecord, error)
	del func(ctx context.Context, org, id string, force bool) error
}

func (c *orgRoleClientMock) CreateRole(
	ctx context.Context, org string, req apitype.PermissionDescriptorBase,
) (*apitype.PermissionDescriptorRecord, error) {
	return c.create(ctx, org, req)
}

func (c *orgRoleClientMock) GetRole(
	ctx context.Context, org, id string,
) (*apitype.PermissionDescriptorRecord, error) {
	return c.get(ctx, org, id)
}

func (c *orgRoleClientMock) UpdateRole(
	ctx context.Context, org, id string, name, desc *string, details apitype.PermissionDescriptor,
) (*apitype.PermissionDescriptorRecord, error) {
	return c.update(ctx, org, id, name, desc, details)
}

func (c *orgRoleClientMock) DeleteRole(ctx context.Context, org, id string, force bool) error {
	return c.del(ctx, org, id, force)
}

// mustParseDescriptor builds a typed PermissionDescriptor from wire-shape JSON
// using the same generated unmarshaller the production code uses.
func mustParseDescriptor(t *testing.T, wireJSON string) apitype.PermissionDescriptor {
	t.Helper()
	var d apitype.PermissionDescriptor
	require.NoError(t, apitype.UnmarshalJSONPermissionDescriptor([]byte(wireJSON), &d))
	require.NotNil(t, d)
	return d
}

var testPermissions = map[string]interface{}{
	"kind":        "PermissionDescriptorAllow",
	"permissions": []interface{}{"stack:read"},
}

func TestOrganizationRoleCreate(t *testing.T) {
	mock := &orgRoleClientMock{
		create: func(_ context.Context, org string, req apitype.PermissionDescriptorBase) (*apitype.PermissionDescriptorRecord, error) {
			assert.Equal(t, "acme", org)
			assert.Equal(t, "read-only", req.Name)
			// Defaulting moved into the resource layer: empty user input
			// becomes "global" before reaching the API.
			assert.Equal(t, "global", req.ResourceType)
			assert.Equal(t, apitype.PermissionDescriptorUXPurposeRole, req.UxPurpose)
			require.NotNil(t, req.Details, "Details must be a typed descriptor")
			// Round-trip the typed descriptor back through JSON to assert the
			// wire shape — the resource layer must emit __type, never `kind`.
			raw, err := json.Marshal(req.Details)
			require.NoError(t, err)
			var parsed map[string]interface{}
			require.NoError(t, json.Unmarshal(raw, &parsed))
			assert.Equal(t, "PermissionDescriptorAllow", parsed["__type"])
			assert.NotContains(t, parsed, "kind", "wire body must not leak `kind` to the API")
			return &apitype.PermissionDescriptorRecord{
				PermissionDescriptorBase: apitype.PermissionDescriptorBase{
					Name:    req.Name,
					Details: req.Details,
				},
				ID:      "role-123",
				Version: 1,
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
			get: func(_ context.Context, _, _ string) (*apitype.PermissionDescriptorRecord, error) { return nil, nil },
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
		details := mustParseDescriptor(t,
			`{"__type":"PermissionDescriptorAllow","permissions":["stack:read"]}`)
		mock := &orgRoleClientMock{
			get: func(_ context.Context, _, _ string) (*apitype.PermissionDescriptorRecord, error) {
				return &apitype.PermissionDescriptorRecord{
					PermissionDescriptorBase: apitype.PermissionDescriptorBase{
						Name:         "read-only",
						Description:  "ro",
						ResourceType: "global",
						UxPurpose:    apitype.PermissionDescriptorUXPurposeRole,
						Details:      details,
					},
					ID:      "role-123",
					Version: 2,
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
		// Read renames the wire `__type` discriminator to `kind`. The
		// PascalCase value is unchanged: SDK kinds are the wire kinds.
		assert.Equal(t, "PermissionDescriptorAllow", resp.State.Permissions["kind"])
		assert.NotContains(t, resp.State.Permissions, "__type", "state must not leak `__type` to the SDK")
	})

	// Pulumi Cloud's permission-descriptor table holds entries for both
	// roles (this resource) and other things (e.g. policies) under the
	// same /orgs/<org>/roles endpoint, distinguished by uxPurpose. A user
	// who points `pulumi import` at a non-role descriptor's id should get
	// a clear error rather than have the descriptor silently round-trip
	// through code that only understands roles. uxPurpose is otherwise
	// hidden from the SDK — Create hardcodes "role".
	t.Run("rejects non-role uxPurpose", func(t *testing.T) {
		mock := &orgRoleClientMock{
			get: func(_ context.Context, _, _ string) (*apitype.PermissionDescriptorRecord, error) {
				return &apitype.PermissionDescriptorRecord{
					PermissionDescriptorBase: apitype.PermissionDescriptorBase{
						Name:      "some-policy",
						UxPurpose: apitype.PermissionDescriptorUXPurposePolicy,
					},
					ID: "policy-123",
				}, nil
			},
		}
		ctx := config.WithMockClient(context.Background(), mock)
		r := &OrganizationRole{}
		_, err := r.Read(ctx, infer.ReadRequest[OrganizationRoleInput, OrganizationRoleState]{
			ID: "acme/policy-123",
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "policy-123")
		assert.Contains(t, err.Error(), "uxPurpose")
		assert.Contains(t, err.Error(), "policy",
			"error must name the actual uxPurpose so the user knows what they pointed at")
	})
}

// TestOrganizationRoleDelete_InUseConflict pins the graceful handling of
// the rejection Pulumi Cloud surfaces when another role's
// PermissionDescriptorCompose still references the role being deleted.
// `force=true` overrides member/team assignments but does *not* override
// structural Compose references, so the delete returns 409 even after
// the force escalation. The provider must wrap that into a message that
// explains *why* and what to do.
func TestOrganizationRoleDelete_InUseConflict(t *testing.T) {
	var calls []bool
	mock := &orgRoleClientMock{
		del: func(_ context.Context, _, _ string, force bool) error {
			calls = append(calls, force)
			return &pulumiapi.ErrorResponse{
				StatusCode: 409,
				Message:    "role is referenced by another role's compose",
			}
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
	// Both attempts (non-force then force) must have been tried before
	// surfacing the wrapped error — the force escalation is what we'd
	// need against a member/team assignment, and only its failure
	// proves the conflict is structural (Compose).
	assert.Equal(t, []bool{false, true}, calls,
		"must try force=false first, then escalate to force=true on 409")
	assert.Error(t, err)
	// Message must name the role, point at PermissionDescriptorCompose
	// (the typical cause), and tell the user how to recover.
	assert.Contains(t, err.Error(), "role-123")
	assert.Contains(t, err.Error(), "PermissionDescriptorCompose")
	assert.Contains(t, err.Error(), "destroy",
		"error should tell the user to destroy the composing role(s) first")
	// The underlying API error must still be wrapped so callers can
	// inspect the status code if they want.
	assert.Equal(t, 409, pulumiapi.GetErrorStatusCode(err),
		"wrapped error must preserve the original 409 status code")
}

// TestOrganizationRoleDelete_EscalatesForceOnConflict pins the
// member/team-assignment escalation path: the unprivileged delete returns
// 409 (role still assigned), the force-true retry succeeds (force clears
// assignments transitively), and Delete returns nil. Without this
// escalation, destroy would fail any time the destroy graph didn't
// happen to clean assignments first (e.g. adopted-member no-op deletes,
// out-of-band assignments).
func TestOrganizationRoleDelete_EscalatesForceOnConflict(t *testing.T) {
	var calls []bool
	mock := &orgRoleClientMock{
		del: func(_ context.Context, _, _ string, force bool) error {
			calls = append(calls, force)
			if !force {
				return &pulumiapi.ErrorResponse{
					StatusCode: 409,
					Message:    "role still assigned to a team",
				}
			}
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
	assert.NoError(t, err, "force-true retry must succeed when assignments are the only blocker")
	assert.Equal(t, []bool{false, true}, calls,
		"must try force=false first, then escalate to force=true on 409")
}

func TestOrganizationRoleDelete(t *testing.T) {
	var calls []bool
	mock := &orgRoleClientMock{
		del: func(_ context.Context, org, id string, force bool) error {
			calls = append(calls, force)
			assert.Equal(t, "acme", org)
			assert.Equal(t, "role-123", id)
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
	assert.Equal(t, []bool{false}, calls,
		"clean delete must succeed on the first (unprivileged) call without escalating to force=true")
}

func TestOrganizationRoleUpdateOmitsDescriptionWhenUnset(t *testing.T) {
	// Regression: Update used to pass `&""` (from util.OrZero) as the description
	// pointer when the user had not set one. With `omitempty` on *string only
	// eliding nil (not empty), the PATCH body included `"description": ""` and
	// cleared any existing description on the server.
	details := mustParseDescriptor(t,
		`{"__type":"PermissionDescriptorAllow","permissions":["stack:read"]}`)
	var gotDesc *string
	gotDesc = new(string) // sentinel so we can distinguish "passed nil" from "test hasn't run"
	*gotDesc = "__sentinel__"
	mock := &orgRoleClientMock{
		update: func(
			_ context.Context, _, _ string, _, desc *string, _ apitype.PermissionDescriptor,
		) (*apitype.PermissionDescriptorRecord, error) {
			gotDesc = desc
			return &apitype.PermissionDescriptorRecord{
				PermissionDescriptorBase: apitype.PermissionDescriptorBase{
					Name:    "read-only",
					Details: details,
				},
				ID:      "role-123",
				Version: 3,
			}, nil
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

	// The translator is structurally agnostic: any `kind` value passes through
	// to the wire format as a `__type` rename. Pulumi Cloud is the source of
	// truth for which descriptor variants exist; the provider does not gate
	// them. This test documents that contract — Check accepts a kind value
	// the provider has no special knowledge of, and the role's validity is
	// the API's call at apply time.
	t.Run("accepts arbitrary kind values (no provider-side allowlist)", func(t *testing.T) {
		resp, err := r.Check(context.Background(), infer.CheckRequest{
			NewInputs: property.NewMap(map[string]property.Value{
				"organizationName": property.New("acme"),
				"name":             property.New("r"),
				"permissions": property.New(property.NewMap(map[string]property.Value{
					"kind": property.New("PermissionDescriptorWhateverFutureCloudVariant"),
				})),
			}),
		})
		assert.NoError(t, err)
		for _, f := range resp.Failures {
			assert.NotEqual(t, "permissions", f.Property,
				"Check must not gate descriptor `kind` values: %s", f.Reason)
		}
	})

	// Pulumi's Python SDK silently strips `__`-prefixed keys from inputs
	// (pulumi/pulumi#22738). A user pasting raw wire format from the REST
	// API docs would see their `__type` discriminator disappear at the
	// language boundary. Check rejects `__type` keys with a clear pointer
	// to the right field, before the malformed body reaches the API.
	t.Run("rejects __type at the SDK boundary", func(t *testing.T) {
		resp, err := r.Check(context.Background(), infer.CheckRequest{
			NewInputs: property.NewMap(map[string]property.Value{
				"organizationName": property.New("acme"),
				"name":             property.New("r"),
				"permissions": property.New(property.NewMap(map[string]property.Value{
					"__type": property.New("PermissionDescriptorAllow"),
					"permissions": property.New(property.NewArray([]property.Value{
						property.New("stack:read"),
					})),
				})),
			}),
		})
		assert.NoError(t, err)
		props := map[string]string{}
		for _, f := range resp.Failures {
			props[f.Property] = f.Reason
		}
		assert.Contains(t, props["permissions"], "__type",
			"Check must name `__type` so the user knows what to fix")
		assert.Contains(t, props["permissions"], "kind",
			"Check must point the user at the correct field name")
	})
}
