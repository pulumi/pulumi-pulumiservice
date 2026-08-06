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

const (
	// The discriminator has two spellings: `__type` on the Pulumi Cloud wire,
	// `type__` everywhere a user or the SDK can see it. Assertions below pick
	// the one matching the side of the boundary under test.
	gcWireType   = "__type"
	gcSchemaType = "type__"
	gcRoleID     = "role-123"
	gcReadOnly   = "read-only"
	gcRoleURN    = "acme/role-123"
)

type orgRoleClientMock struct {
	config.Client
	create func(
		ctx context.Context, org string, req apitype.PermissionDescriptorBase,
	) (*apitype.PermissionDescriptorRecord, error)
	get    func(ctx context.Context, org, id string) (*apitype.PermissionDescriptorRecord, error)
	update func(
		ctx context.Context, org, id string, req apitype.UpdateRoleRequest,
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
	ctx context.Context, org, id string, req apitype.UpdateRoleRequest,
) (*apitype.PermissionDescriptorRecord, error) {
	return c.update(ctx, org, id, req)
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
	gcSchemaType:  "PermissionDescriptorAllow",
	gcPermissions: []interface{}{"stack:read"},
}

func TestOrganizationRoleCreate(t *testing.T) {
	mock := &orgRoleClientMock{
		create: func(
			_ context.Context, org string, req apitype.PermissionDescriptorBase,
		) (*apitype.PermissionDescriptorRecord, error) {
			assert.Equal(t, gcAcme, org)
			assert.Equal(t, gcReadOnly, req.Name)
			// Defaulting moved into the resource layer: empty user input
			// becomes "global" before reaching the API.
			assert.Equal(t, gcGlobal, req.ResourceType)
			assert.Equal(t, apitype.PermissionDescriptorUXPurposeRole, req.UxPurpose)
			require.NotNil(t, req.Details, "Details must be a typed descriptor")
			// Round-trip the typed descriptor back through JSON to assert the
			// request still carries the wire spelling, even though the user
			// wrote `type__` in testPermissions.
			raw, err := json.Marshal(req.Details)
			require.NoError(t, err)
			var parsed map[string]interface{}
			require.NoError(t, json.Unmarshal(raw, &parsed))
			assert.Equal(t, "PermissionDescriptorAllow", parsed[gcWireType])
			return &apitype.PermissionDescriptorRecord{
				PermissionDescriptorBase: apitype.PermissionDescriptorBase{
					Name:    req.Name,
					Details: req.Details,
				},
				ID:      gcRoleID,
				Version: 1,
			}, nil
		},
	}
	ctx := config.WithMockClient(context.Background(), mock)

	r := &OrganizationRole{}
	resp, err := r.Create(ctx, infer.CreateRequest[OrganizationRoleInput]{
		Inputs: OrganizationRoleInput{
			OrganizationRoleCore: OrganizationRoleCore{
				OrganizationName: gcAcme,
				Name:             gcReadOnly,
				Permissions:      testPermissions,
			},
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, gcRoleURN, resp.ID)
	assert.Equal(t, gcRoleID, resp.Output.RoleId)
}

func TestOrganizationRoleRead(t *testing.T) {
	t.Run("not found", func(t *testing.T) {
		mock := &orgRoleClientMock{
			get: func(_ context.Context, _, _ string) (*apitype.PermissionDescriptorRecord, error) { return nil, nil },
		}
		ctx := config.WithMockClient(context.Background(), mock)
		r := &OrganizationRole{}
		resp, err := r.Read(ctx, infer.ReadRequest[OrganizationRoleInput, OrganizationRoleState]{
			ID: gcRoleURN,
		})
		assert.NoError(t, err)
		assert.Equal(t, "", resp.ID)
	})

	t.Run("found parses details", func(t *testing.T) {
		// The API returns wire format with `__type`; the provider renames it
		// to `type__` before the descriptor reaches state.
		details := mustParseDescriptor(t,
			`{"__type":"PermissionDescriptorAllow","permissions":["stack:read"]}`)
		mock := &orgRoleClientMock{
			get: func(_ context.Context, _, _ string) (*apitype.PermissionDescriptorRecord, error) {
				return &apitype.PermissionDescriptorRecord{
					PermissionDescriptorBase: apitype.PermissionDescriptorBase{
						Name:         gcReadOnly,
						Description:  "ro",
						ResourceType: gcGlobal,
						UxPurpose:    apitype.PermissionDescriptorUXPurposeRole,
						Details:      details,
					},
					ID:      gcRoleID,
					Version: 2,
				}, nil
			},
		}
		ctx := config.WithMockClient(context.Background(), mock)
		r := &OrganizationRole{}
		resp, err := r.Read(ctx, infer.ReadRequest[OrganizationRoleInput, OrganizationRoleState]{
			ID: gcRoleURN,
		})
		assert.NoError(t, err)
		assert.Equal(t, gcRoleURN, resp.ID)
		assert.Equal(t, "PermissionDescriptorAllow", resp.State.Permissions[gcSchemaType])
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

// TestOrganizationRoleDiscriminatorRenameIsRecursive pins the rename at every
// depth of the descriptor tree, in both directions. A top-level-only
// implementation passes the Create and Read cases above but sends `type__`
// verbatim to the API from the second level down, which the service rejects.
func TestOrganizationRoleDiscriminatorRenameIsRecursive(t *testing.T) {
	nested := map[string]interface{}{
		gcSchemaType: "PermissionDescriptorCondition",
		"condition": map[string]interface{}{
			gcSchemaType: "PermissionExpressionEqual",
			"left":       map[string]interface{}{gcSchemaType: "PermissionExpressionEnvironment"},
			"right": map[string]interface{}{
				gcSchemaType: "PermissionLiteralExpressionEnvironment",
				"identity":   "acme/prod",
			},
		},
		"subNode": map[string]interface{}{
			gcSchemaType:  "PermissionDescriptorAllow",
			gcPermissions: []interface{}{"environment:open"},
		},
	}

	var sent map[string]interface{}
	mock := &orgRoleClientMock{
		create: func(
			_ context.Context, _ string, req apitype.PermissionDescriptorBase,
		) (*apitype.PermissionDescriptorRecord, error) {
			raw, err := json.Marshal(req.Details)
			require.NoError(t, err)
			require.NoError(t, json.Unmarshal(raw, &sent))
			return &apitype.PermissionDescriptorRecord{
				PermissionDescriptorBase: apitype.PermissionDescriptorBase{
					Name:      gcReadOnly,
					UxPurpose: apitype.PermissionDescriptorUXPurposeRole,
					Details:   req.Details,
				},
				ID:      gcRoleID,
				Version: 1,
			}, nil
		},
	}
	ctx := config.WithMockClient(context.Background(), mock)
	r := &OrganizationRole{}

	_, err := r.Create(ctx, infer.CreateRequest[OrganizationRoleInput]{
		Inputs: OrganizationRoleInput{
			OrganizationRoleCore: OrganizationRoleCore{
				OrganizationName: gcAcme,
				Name:             gcReadOnly,
				Permissions:      nested,
			},
		},
	})
	require.NoError(t, err)

	// Request side: every level carries the wire spelling and none carries
	// the schema spelling.
	cond, ok := sent["condition"].(map[string]interface{})
	require.True(t, ok, "condition must survive as a map; got %T", sent["condition"])
	left, ok := cond["left"].(map[string]interface{})
	require.True(t, ok, "condition.left must survive as a map; got %T", cond["left"])
	sub, ok := sent["subNode"].(map[string]interface{})
	require.True(t, ok, "subNode must survive as a map; got %T", sent["subNode"])
	for _, level := range []map[string]interface{}{sent, cond, left, sub} {
		assert.NotEmpty(t, level[gcWireType], "every level must carry `__type` on the wire")
		assert.Nil(t, level[gcSchemaType], "no level may leak `type__` to the API")
	}
	assert.Equal(t, "PermissionExpressionEnvironment", left[gcWireType])

	// Response side: reading the same descriptor back restores `type__` at
	// every level, so state matches what the user wrote.
	getMock := &orgRoleClientMock{
		get: func(_ context.Context, _, _ string) (*apitype.PermissionDescriptorRecord, error) {
			details := mustParseDescriptor(t, string(mustMarshal(t, sent)))
			return &apitype.PermissionDescriptorRecord{
				PermissionDescriptorBase: apitype.PermissionDescriptorBase{
					Name:      gcReadOnly,
					UxPurpose: apitype.PermissionDescriptorUXPurposeRole,
					Details:   details,
				},
				ID:      gcRoleID,
				Version: 1,
			}, nil
		},
	}
	resp, err := r.Read(config.WithMockClient(context.Background(), getMock),
		infer.ReadRequest[OrganizationRoleInput, OrganizationRoleState]{ID: gcRoleURN})
	require.NoError(t, err)
	assert.Equal(t, nested, resp.State.Permissions,
		"a descriptor read back must equal the tree the user authored")
}

func mustMarshal(t *testing.T, v any) []byte {
	t.Helper()
	raw, err := json.Marshal(v)
	require.NoError(t, err)
	return raw
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
			OrganizationRoleCore: OrganizationRoleCore{OrganizationName: gcAcme},
			RoleId:               gcRoleID,
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
	assert.Contains(t, err.Error(), gcRoleID)
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
			OrganizationRoleCore: OrganizationRoleCore{OrganizationName: gcAcme},
			RoleId:               gcRoleID,
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
			assert.Equal(t, gcAcme, org)
			assert.Equal(t, gcRoleID, id)
			return nil
		},
	}
	ctx := config.WithMockClient(context.Background(), mock)
	r := &OrganizationRole{}
	_, err := r.Delete(ctx, infer.DeleteRequest[OrganizationRoleState]{
		State: OrganizationRoleState{
			OrganizationRoleCore: OrganizationRoleCore{OrganizationName: gcAcme},
			RoleId:               gcRoleID,
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
			_ context.Context, _, _ string, req apitype.UpdateRoleRequest,
		) (*apitype.PermissionDescriptorRecord, error) {
			gotDesc = req.Description
			return &apitype.PermissionDescriptorRecord{
				PermissionDescriptorBase: apitype.PermissionDescriptorBase{
					Name:    gcReadOnly,
					Details: details,
				},
				ID:      gcRoleID,
				Version: 3,
			}, nil
		},
	}
	ctx := config.WithMockClient(context.Background(), mock)
	r := &OrganizationRole{}
	_, err := r.Update(ctx, infer.UpdateRequest[OrganizationRoleInput, OrganizationRoleState]{
		Inputs: OrganizationRoleInput{
			OrganizationRoleCore: OrganizationRoleCore{
				OrganizationName: gcAcme,
				Name:             gcReadOnly,
				// Description intentionally unset.
				Permissions: testPermissions,
			},
		},
		State: OrganizationRoleState{
			OrganizationRoleCore: OrganizationRoleCore{OrganizationName: gcAcme},
			RoleId:               gcRoleID,
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
				gcOrganizationName: property.New(gcAcme),
				gcName:             property.New("r"),
			}),
		})
		assert.NoError(t, err)
		props := map[string]bool{}
		for _, f := range resp.Failures {
			props[f.Property] = true
		}
		assert.True(t, props[gcPermissions])
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
				gcOrganizationName: property.New(gcAcme),
				gcName:             property.New("r"),
				gcPermissions:      property.New(property.Computed),
			}),
		})
		assert.NoError(t, err)
		for _, f := range resp.Failures {
			assert.NotEqual(t, gcPermissions, f.Property,
				"Check must not reject a computed permissions input: %s", f.Reason)
		}
	})

	t.Run("tolerates computed name at preview", func(t *testing.T) {
		resp, err := r.Check(context.Background(), infer.CheckRequest{
			NewInputs: property.NewMap(map[string]property.Value{
				gcOrganizationName: property.New(gcAcme),
				gcName:             property.New(property.Computed),
				gcPermissions: property.New(property.NewMap(map[string]property.Value{
					gcSchemaType: property.New("allow"),
				})),
			}),
		})
		assert.NoError(t, err)
		for _, f := range resp.Failures {
			assert.NotEqual(t, gcName, f.Property,
				"Check must not reject a computed name input: %s", f.Reason)
		}
	})

	// The provider validates `type__` against the typed
	// `apitype.PermissionDescriptor` discriminator catalogue (generated
	// from the OpenAPI spec the API uses). Unknown variants are rejected
	// at preview with a clear "type 'X' not recognized" message rather
	// than blindly forwarded to the API. New variants Pulumi Cloud adds
	// reach this provider through `apitype` regen.
	t.Run("rejects unknown type__ values", func(t *testing.T) {
		resp, err := r.Check(context.Background(), infer.CheckRequest{
			NewInputs: property.NewMap(map[string]property.Value{
				gcOrganizationName: property.New(gcAcme),
				gcName:             property.New("r"),
				gcPermissions: property.New(property.NewMap(map[string]property.Value{
					gcSchemaType: property.New("PermissionDescriptorWhateverFutureCloudVariant"),
				})),
			}),
		})
		assert.NoError(t, err)
		props := map[string]string{}
		for _, f := range resp.Failures {
			props[f.Property] = f.Reason
		}
		assert.Contains(t, props[gcPermissions], "PermissionDescriptorWhateverFutureCloudVariant",
			"Check must name the unrecognized variant so the user knows what to fix")
		assert.Contains(t, props[gcPermissions], "not recognized",
			"Check must surface the typed unmarshaller's diagnostic")
	})

	// A descriptor missing the top-level `type__` discriminator is
	// rejected at preview by the typed unmarshaller (`type '' not
	// recognized`) rather than reaching the API.
	t.Run("rejects descriptor missing type__", func(t *testing.T) {
		resp, err := r.Check(context.Background(), infer.CheckRequest{
			NewInputs: property.NewMap(map[string]property.Value{
				gcOrganizationName: property.New(gcAcme),
				gcName:             property.New("r"),
				gcPermissions: property.New(property.NewMap(map[string]property.Value{
					gcPermissions: property.New(property.NewArray([]property.Value{
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
		assert.Contains(t, props[gcPermissions], "not recognized",
			"Check must reject a descriptor with no `type__` discriminator")
	})
}
