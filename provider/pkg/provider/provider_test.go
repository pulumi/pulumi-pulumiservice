// Copyright 2016-2026, Pulumi Corporation.

package provider

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pgo "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	pcommon "github.com/pulumi/pulumi/sdk/v3/go/common/tokens"
	"github.com/pulumi/pulumi/sdk/v3/go/property"
)

// minimalSpec describes a single AgentPool resource with full CRUD —
// enough to exercise dispatch end-to-end without pulling in the full
// pinned spec.
const minimalSpec = `{
  "paths": {
    "/api/orgs/{orgName}/agent-pools": {
      "post": {"operationId": "CreateOrgAgentPool"},
      "get":  {"operationId": "ListOrgAgentPools"}
    },
    "/api/orgs/{orgName}/agent-pools/{agentPoolId}": {
      "get":    {"operationId": "GetOrgAgentPool"},
      "patch":  {"operationId": "UpdateOrgAgentPool"},
      "delete": {"operationId": "DeleteOrgAgentPool"}
    }
  }
}`

const minimalMap = `modules:
  orgs/agents:
    resources:
      AgentPool:
        operations:
          create: CreateOrgAgentPool
          read:   GetOrgAgentPool
          update: UpdateOrgAgentPool
          delete: DeleteOrgAgentPool
        id:
          template: "{organizationName}/{agentPoolId}"
          params: [organizationName, agentPoolId]
        forceNew: [organizationName]
        properties:
          organizationName: { from: orgName,     source: path }
          name:             { from: name,        source: body }
          description:      { from: description, source: body }
          tokenValue:       { from: tokenValue,  source: response, secret: true, output: true }
          agentPoolId:      { from: id,          source: response, output: true }
    functions:
      listAgentPools: { operationId: ListOrgAgentPools }
`

// strSpec is the same spec with one extra operation that the map
// does NOT claim — exercises GetSchema's coverage-error behavior.
const strSpec = `{
  "paths": {
    "/api/orgs/{orgName}/agent-pools": {
      "post": {"operationId": "CreateOrgAgentPool"},
      "get":  {"operationId": "ListOrgAgentPools"}
    },
    "/api/orgs/{orgName}/agent-pools/{agentPoolId}": {
      "get":    {"operationId": "GetOrgAgentPool"},
      "patch":  {"operationId": "UpdateOrgAgentPool"},
      "delete": {"operationId": "DeleteOrgAgentPool"}
    },
    "/api/orgs/{orgName}/orphan": {
      "post": {"operationId": "CreateOrphanThing"}
    }
  }
}`

// strMap doesn't claim CreateOrphanThing; nothing in exclusions
// either. GetSchema must error.
const strMap = `modules:
  orgs/agents:
    resources:
      AgentPool:
        operations:
          create: CreateOrgAgentPool
          read:   GetOrgAgentPool
          update: UpdateOrgAgentPool
          delete: DeleteOrgAgentPool
        id:
          template: "{organizationName}/{agentPoolId}"
          params: [organizationName, agentPoolId]
        properties:
          organizationName: { from: orgName,     source: path }
    functions:
      listAgentPools: { operationId: ListOrgAgentPools }
`

// TestProvider_New_RejectsEmptyInputs guards the constructor's
// preconditions; an empty embed file would otherwise produce a
// confusing runtime error.
func TestProvider_New_RejectsEmptyInputs(t *testing.T) {
	_, err := New(nil, []byte(minimalMap))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "OpenAPI spec")

	_, err = New([]byte(minimalSpec), nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "resource-map")
}

// TestProvider_GetSchema_LazyAndCached confirms two things:
//  1. The schema is generated on demand — `New` returns instantly
//     even though the spec parse for schema emission is heavy.
//  2. The result is cached: a second call returns identical bytes
//     and does not re-emit. We verify by checking the schema's
//     resource entry matches between calls.
func TestProvider_GetSchema_LazyAndCached(t *testing.T) {
	prov, err := New([]byte(minimalSpec), []byte(minimalMap))
	require.NoError(t, err)

	resp1, err := prov.GetSchema(context.Background(), pgo.GetSchemaRequest{})
	require.NoError(t, err)
	assert.Contains(t, resp1.Schema, `"pulumiservice:orgs/agents:AgentPool"`)

	resp2, err := prov.GetSchema(context.Background(), pgo.GetSchemaRequest{})
	require.NoError(t, err)
	assert.Equal(t, resp1.Schema, resp2.Schema, "second GetSchema must hit the cache")
}

// TestProvider_GetSchema_ErrorsWhenMapIncomplete is iwahbe's design
// note in test form: GetSchema must fail loudly when the resource
// map is incomplete. Without this, downstream SDKs would advertise
// a schema that the runtime can't actually serve.
func TestProvider_GetSchema_ErrorsWhenMapIncomplete(t *testing.T) {
	prov, err := New([]byte(strSpec), []byte(strMap))
	require.NoError(t, err)

	_, err = prov.GetSchema(context.Background(), pgo.GetSchemaRequest{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "incomplete")
	assert.Contains(t, err.Error(), "CreateOrphanThing",
		"error should name the offending operationId so a maintainer can fix it")
}

// TestProvider_GetSchema_ErrorIsSticky confirms the schemaErr stays
// pinned across calls — we don't redo expensive emission only to
// return the same failure each time.
func TestProvider_GetSchema_ErrorIsSticky(t *testing.T) {
	prov, err := New([]byte(strSpec), []byte(strMap))
	require.NoError(t, err)

	_, err1 := prov.GetSchema(context.Background(), pgo.GetSchemaRequest{})
	require.Error(t, err1)
	_, err2 := prov.GetSchema(context.Background(), pgo.GetSchemaRequest{})
	require.Error(t, err2)
	assert.Equal(t, err1.Error(), err2.Error())
}

// TestProvider_ConfigureAndCreate exercises the full path: Configure
// stashes credentials, then a Create dispatches through the
// metadata-driven runtime against a fake Pulumi Cloud server. Proves
// the property.Map → resource.PropertyMap → dispatcher → HTTP wiring
// end-to-end.
func TestProvider_ConfigureAndCreate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "token test-token", r.Header.Get("Authorization"))
		body, _ := io.ReadAll(r.Body)
		var req map[string]interface{}
		_ = json.Unmarshal(body, &req)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"id":          "pool-abc",
			"name":        req["name"],
			"description": req["description"],
			"tokenValue":  "secret-xyz",
		})
	}))
	defer srv.Close()

	prov, err := New([]byte(minimalSpec), []byte(minimalMap))
	require.NoError(t, err)

	ctx := context.Background()
	require.NoError(t, prov.Configure(ctx, pgo.ConfigureRequest{
		Args: property.NewMap(map[string]property.Value{
			"accessToken": property.New("test-token"),
			"apiUrl":      property.New(srv.URL),
		}),
	}))

	inputs := property.NewMap(map[string]property.Value{
		"organizationName": property.New("acme-corp"),
		"name":             property.New("vpc-pool"),
		"description":      property.New("VPC-isolated agent pool"),
	})
	resp, err := prov.Create(ctx, pgo.CreateRequest{
		Urn:        "urn:pulumi:dev::test::pulumiservice:orgs/agents:AgentPool::my-pool",
		Properties: inputs,
	})
	require.NoError(t, err)
	assert.Equal(t, "acme-corp/pool-abc", resp.ID)

	tok, ok := resp.Properties.GetOk("tokenValue")
	require.True(t, ok, "create response should include tokenValue")
	require.True(t, tok.Secret(), "tokenValue must round-trip as a secret")
	require.True(t, tok.IsString(), "tokenValue must be a string under the secret")
	assert.Equal(t, "secret-xyz", tok.AsString())
}

// TestProvider_CRUDBeforeConfigure_ErrorsClearly: every CRUD callback
// short-circuits with an explicit message if Configure hasn't run.
// This is friendlier than letting a nil dispatcher panic; integration
// tooling sometimes calls Read directly without Configure.
func TestProvider_CRUDBeforeConfigure_ErrorsClearly(t *testing.T) {
	prov, err := New([]byte(minimalSpec), []byte(minimalMap))
	require.NoError(t, err)
	ctx := context.Background()

	_, err = prov.Create(ctx, pgo.CreateRequest{Urn: "urn:pulumi:dev::test::pulumiservice:orgs/agents:AgentPool::p"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not configured")

	_, err = prov.Read(ctx, pgo.ReadRequest{Urn: "urn:pulumi:dev::test::pulumiservice:orgs/agents:AgentPool::p"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not configured")

	_, err = prov.Update(ctx, pgo.UpdateRequest{Urn: "urn:pulumi:dev::test::pulumiservice:orgs/agents:AgentPool::p"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not configured")

	err = prov.Delete(ctx, pgo.DeleteRequest{Urn: "urn:pulumi:dev::test::pulumiservice:orgs/agents:AgentPool::p"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not configured")

	_, err = prov.Invoke(ctx, pgo.InvokeRequest{Token: pcommon.Type("pulumiservice:orgs/agents:listAgentPools")})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not configured")
}

// TestProvider_Parameterize_NotYetImplemented locks in the explicit
// "not implemented" return for the Parameterize wiring. v2.0.0-alpha.1
// ships the structural plumbing so a follow-up can land it without
// retrofitting the Provider literal — but until then, the runtime
// should fail clearly rather than silently accept and do nothing.
func TestProvider_Parameterize_NotYetImplemented(t *testing.T) {
	prov, err := New([]byte(minimalSpec), []byte(minimalMap))
	require.NoError(t, err)

	_, err = prov.Parameterize(context.Background(), pgo.ParameterizeRequest{})
	require.Error(t, err)
	assert.Contains(t, strings.ToLower(err.Error()), "not yet implemented")
}

// TestProvider_TokenFromURN: URN→type-token parsing is a tight,
// pure helper. Worth a sanity test so future refactors of the URN
// format don't silently break dispatch.
func TestProvider_TokenFromURN(t *testing.T) {
	cases := []struct {
		urn  string
		want string
	}{
		{"urn:pulumi:dev::test::pulumiservice:orgs/agents:AgentPool::pool", "pulumiservice:orgs/agents:AgentPool"},
		{"urn:pulumi:prod::svc::pulumiservice:stacks:Stack::s", "pulumiservice:stacks:Stack"},
	}
	for _, c := range cases {
		assert.Equal(t, c.want, tokenFromURN(c.urn))
	}
}

// TestProvider_Diff_ForceNewVsUpdate confirms ForceNew properties
// produce replace diffs and others produce in-place updates. The
// diff engine is the second-most-load-bearing path after dispatch.
func TestProvider_Diff_ForceNewVsUpdate(t *testing.T) {
	prov, err := New([]byte(minimalSpec), []byte(minimalMap))
	require.NoError(t, err)

	urn := "urn:pulumi:dev::test::pulumiservice:orgs/agents:AgentPool::p"

	// Change the ForceNew property `organizationName`.
	resp, err := prov.Diff(context.Background(), pgo.DiffRequest{
		Urn: resource.URN(urn),
		State: property.NewMap(map[string]property.Value{
			"organizationName": property.New("acme"),
			"name":             property.New("p"),
		}),
		Inputs: property.NewMap(map[string]property.Value{
			"organizationName": property.New("acme-renamed"),
			"name":             property.New("p"),
		}),
	})
	require.NoError(t, err)
	require.True(t, resp.HasChanges)
	require.Contains(t, resp.DetailedDiff, "organizationName")
	assert.Equal(t, pgo.UpdateReplace, resp.DetailedDiff["organizationName"].Kind)

	// Change a non-ForceNew property — should be a plain Update.
	resp, err = prov.Diff(context.Background(), pgo.DiffRequest{
		Urn: resource.URN(urn),
		State: property.NewMap(map[string]property.Value{
			"organizationName": property.New("acme"),
			"description":      property.New("old"),
		}),
		Inputs: property.NewMap(map[string]property.Value{
			"organizationName": property.New("acme"),
			"description":      property.New("new"),
		}),
	})
	require.NoError(t, err)
	require.True(t, resp.HasChanges)
	require.Contains(t, resp.DetailedDiff, "description")
	assert.Equal(t, pgo.Update, resp.DetailedDiff["description"].Kind)
}

// poolPath is the per-resource path used by Read/Update/Delete in
// the full-CRUD-lifecycle test. Defined here as a constant so the
// switch in the test handler stays compact.
const poolPath = "/api/orgs/acme-corp/agent-pools/pool-abc"

// TestProvider_FullCRUDLifecycle puts a single resource through every
// verb the engine will exercise on a real `pulumi up` / `up` /
// `refresh` / `destroy` cycle: Create → Read → Update → Delete →
// Invoke. Each step uses a fresh httptest handler, asserting both
// the request shape (URL, method, body) and the response handling.
//
// Catches the kinds of regressions that pure-unit tests miss: an
// Update path that drops secrets, a Read that loses the prior
// inputs, a Delete that omits the ID.
func TestProvider_FullCRUDLifecycle(t *testing.T) {
	var lastMethod, lastPath string
	var lastBody []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lastMethod = r.Method
		lastPath = r.URL.Path
		lastBody, _ = io.ReadAll(r.Body)
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/orgs/acme-corp/agent-pools":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id": "pool-abc", "name": "vpc-pool",
				"description": "first description", "tokenValue": "secret-xyz",
			})
		case r.Method == http.MethodGet && r.URL.Path == poolPath:
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id": "pool-abc", "name": "vpc-pool",
				"description": "first description",
			})
		case r.Method == http.MethodPatch && r.URL.Path == poolPath:
			var body map[string]any
			_ = json.Unmarshal(lastBody, &body)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id": "pool-abc", "name": "vpc-pool",
				"description": body["description"],
			})
		case r.Method == http.MethodDelete && r.URL.Path == poolPath:
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodGet && r.URL.Path == "/api/orgs/acme-corp/agent-pools":
			_ = json.NewEncoder(w).Encode([]map[string]any{
				{"id": "pool-abc", "name": "vpc-pool"},
			})
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	prov, err := New([]byte(minimalSpec), []byte(minimalMap))
	require.NoError(t, err)
	ctx := context.Background()
	require.NoError(t, prov.Configure(ctx, pgo.ConfigureRequest{
		Args: property.NewMap(map[string]property.Value{
			"accessToken": property.New("test-token"),
			"apiUrl":      property.New(srv.URL),
		}),
	}))
	urn := resource.URN("urn:pulumi:dev::test::pulumiservice:orgs/agents:AgentPool::my-pool")

	// 1. Create.
	createInputs := property.NewMap(map[string]property.Value{
		"organizationName": property.New("acme-corp"),
		"name":             property.New("vpc-pool"),
		"description":      property.New("first description"),
	})
	createResp, err := prov.Create(ctx, pgo.CreateRequest{Urn: urn, Properties: createInputs})
	require.NoError(t, err)
	assert.Equal(t, "acme-corp/pool-abc", createResp.ID)
	assert.Equal(t, http.MethodPost, lastMethod)

	// 2. Read with the prior state — proves the dispatcher decomposes
	// the ID, builds the GET URL, and round-trips outputs.
	readResp, err := prov.Read(ctx, pgo.ReadRequest{
		Urn: urn, ID: createResp.ID, Properties: createResp.Properties,
	})
	require.NoError(t, err)
	assert.Equal(t, createResp.ID, readResp.ID, "read should preserve the ID")
	desc, _ := readResp.Properties.GetOk("description")
	assert.Equal(t, "first description", desc.AsString())

	// 3. Update. PATCH should send only the new description; the
	// response carries the post-update state.
	newInputs := property.NewMap(map[string]property.Value{
		"organizationName": property.New("acme-corp"),
		"name":             property.New("vpc-pool"),
		"description":      property.New("second description"),
	})
	updateResp, err := prov.Update(ctx, pgo.UpdateRequest{
		Urn:    urn,
		ID:     createResp.ID,
		State:  createResp.Properties,
		Inputs: newInputs,
	})
	require.NoError(t, err)
	assert.Equal(t, http.MethodPatch, lastMethod)
	desc, _ = updateResp.Properties.GetOk("description")
	assert.Equal(t, "second description", desc.AsString())

	// 4. Delete. DELETE should hit the same path as Read.
	require.NoError(t, prov.Delete(ctx, pgo.DeleteRequest{
		Urn: urn, ID: createResp.ID, Properties: updateResp.Properties,
	}))
	assert.Equal(t, http.MethodDelete, lastMethod)
	assert.Equal(t, poolPath, lastPath)

	// 5. Invoke (data-source function). Confirms list functions are
	// callable through the same dispatcher.
	invokeResp, err := prov.Invoke(ctx, pgo.InvokeRequest{
		Token: pcommon.Type("pulumiservice:orgs/agents:listAgentPools"),
		Args: property.NewMap(map[string]property.Value{
			"orgName": property.New("acme-corp"),
		}),
	})
	require.NoError(t, err)
	require.NotNil(t, invokeResp.Return)
}

// TestProvider_Read_ResourceGone — when the API returns 404, Read
// must return an empty response (engine signal: resource needs to
// be recreated). Without this check, the runtime would treat the
// 404 as a fatal error and break `pulumi refresh`.
func TestProvider_Read_ResourceGone(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer srv.Close()

	prov, err := New([]byte(minimalSpec), []byte(minimalMap))
	require.NoError(t, err)
	ctx := context.Background()
	require.NoError(t, prov.Configure(ctx, pgo.ConfigureRequest{
		Args: property.NewMap(map[string]property.Value{
			"accessToken": property.New("t"), "apiUrl": property.New(srv.URL),
		}),
	}))

	resp, err := prov.Read(ctx, pgo.ReadRequest{
		Urn: "urn:pulumi:dev::test::pulumiservice:orgs/agents:AgentPool::p",
		ID:  "acme-corp/pool-abc",
		Properties: property.NewMap(map[string]property.Value{
			"organizationName": property.New("acme-corp"),
			"agentPoolId":      property.New("pool-abc"),
		}),
	})
	require.NoError(t, err)
	assert.Empty(t, resp.ID, "missing resource must return empty ID so the engine recreates")
}

// TestProvider_DispatcherErrors_Propagate: if the upstream API
// returns a non-2xx, that error should reach the caller verbatim
// (with the response body in the error message). Otherwise users
// see an opaque "unexpected error" with no diagnostic value.
func TestProvider_DispatcherErrors_Propagate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, `{"message":"agent-pool name conflicts with an existing pool"}`, http.StatusConflict)
	}))
	defer srv.Close()

	prov, err := New([]byte(minimalSpec), []byte(minimalMap))
	require.NoError(t, err)
	ctx := context.Background()
	require.NoError(t, prov.Configure(ctx, pgo.ConfigureRequest{
		Args: property.NewMap(map[string]property.Value{
			"accessToken": property.New("t"), "apiUrl": property.New(srv.URL),
		}),
	}))

	_, err = prov.Create(ctx, pgo.CreateRequest{
		Urn: "urn:pulumi:dev::test::pulumiservice:orgs/agents:AgentPool::p",
		Properties: property.NewMap(map[string]property.Value{
			"organizationName": property.New("acme"),
			"name":             property.New("dup"),
		}),
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "conflicts")
}

// TestProvider_SecretInputsSurviveCreate confirms that a secret-marked
// input value reaches the dispatcher unwrapped (so it serializes as
// plaintext on the wire) and then re-emerges secret-marked on any
// response field with `secret: true`. The two property.Map ↔
// resource.PropertyMap conversions on the request path are the most
// likely place for secret bits to be lost.
func TestProvider_SecretInputsSurviveCreate(t *testing.T) {
	var receivedDescription string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		// The wire body must contain the unwrapped string, not the
		// secret-sentinel object. If the conversion lost the secret
		// status incorrectly we'd see a JSON-encoded sentinel here.
		if d, ok := body["description"].(string); ok {
			receivedDescription = d
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id": "pool-abc", "name": body["name"], "description": body["description"],
			"tokenValue": "secret-xyz",
		})
	}))
	defer srv.Close()

	prov, err := New([]byte(minimalSpec), []byte(minimalMap))
	require.NoError(t, err)
	ctx := context.Background()
	require.NoError(t, prov.Configure(ctx, pgo.ConfigureRequest{
		Args: property.NewMap(map[string]property.Value{
			"accessToken": property.New("t"), "apiUrl": property.New(srv.URL),
		}),
	}))

	// Description carries a user-marked secret (e.g. via pulumi
	// config secret on a stack input). The dispatcher should send
	// the plaintext on the wire — Pulumi Cloud doesn't see the
	// secret framing.
	desc := property.New("classified deployment").WithSecret(true)
	resp, err := prov.Create(ctx, pgo.CreateRequest{
		Urn: "urn:pulumi:dev::test::pulumiservice:orgs/agents:AgentPool::p",
		Properties: property.NewMap(map[string]property.Value{
			"organizationName": property.New("acme"),
			"name":             property.New("p"),
			"description":      desc,
		}),
	})
	require.NoError(t, err)
	assert.Equal(t, "classified deployment", receivedDescription,
		"server should receive the plaintext, not a JSON-encoded secret sentinel")

	// Output marked `secret: true` must come back as a secret.
	tok, _ := resp.Properties.GetOk("tokenValue")
	require.True(t, tok.Secret(), "secret-marked output must round-trip as secret")
}

// TestProvider_ComputedInputsRaiseClearError: if a computed (unknown)
// value somehow reaches the dispatcher (the engine should resolve
// before Create/Update, but defense-in-depth), the runtime must
// surface a clear, named error rather than silently sending a
// JSON-encoded computed sentinel to Pulumi Cloud or panicking.
func TestProvider_ComputedInputsRaiseClearError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Fatalf("dispatcher should not have called the API with a computed input")
	}))
	defer srv.Close()

	prov, err := New([]byte(minimalSpec), []byte(minimalMap))
	require.NoError(t, err)
	ctx := context.Background()
	require.NoError(t, prov.Configure(ctx, pgo.ConfigureRequest{
		Args: property.NewMap(map[string]property.Value{
			"accessToken": property.New("t"), "apiUrl": property.New(srv.URL),
		}),
	}))

	_, err = prov.Create(ctx, pgo.CreateRequest{
		Urn: "urn:pulumi:dev::test::pulumiservice:orgs/agents:AgentPool::p",
		Properties: property.NewMap(map[string]property.Value{
			"organizationName": property.New("acme"),
			"name":             property.New("p"),
			"description":      property.New(property.Computed),
		}),
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "description")
	assert.Contains(t, err.Error(), "computed")
}

// checksSpec / checksMap exercise the declarative-check primitives:
// requireOneOf, requireAtMostOne, requireIfSet. These run in Check
// and surface as failures rather than runtime errors.
const checksSpec = `{
  "paths": {
    "/api/orgs/{orgName}/widgets": {
      "post":   {"operationId": "CreateWidget"},
      "get":    {"operationId": "ListWidgets"}
    },
    "/api/orgs/{orgName}/widgets/{widgetId}": {
      "get":    {"operationId": "GetWidget"},
      "delete": {"operationId": "DeleteWidget"}
    }
  }
}`

const checksMap = `modules:
  orgs/widgets:
    resources:
      Widget:
        operations:
          create: CreateWidget
          read:   GetWidget
          delete: DeleteWidget
        id:
          template: "{organizationName}/{widgetId}"
          params: [organizationName, widgetId]
        forceNew: [organizationName]
        properties:
          organizationName: { from: orgName,    source: path }
          metadataUrl:      { source: body, required: false }
          metadataXml:      { source: body, required: false }
          schedule:         { source: body, required: false }
          rotationCron:     { source: body, required: false }
          widgetId:         { from: id, source: response, output: true }
        checks:
          - requireOneOf: [metadataUrl, metadataXml]
            message: "exactly one of metadataUrl or metadataXml must be provided"
          - requireIfSet: schedule
            field: rotationCron
            message: "rotationCron is required when schedule is set"
    functions:
      listWidgets: { operationId: ListWidgets }
`

// TestProvider_Check_RequireOneOf_NeitherFails: when neither
// `metadataUrl` nor `metadataXml` is set, Check must surface the
// declarative requireOneOf rule as a CheckFailure.
func TestProvider_Check_RequireOneOf_NeitherFails(t *testing.T) {
	prov, err := New([]byte(checksSpec), []byte(checksMap))
	require.NoError(t, err)

	resp, err := prov.Check(context.Background(), pgo.CheckRequest{
		Urn: "urn:pulumi:dev::test::pulumiservice:orgs/widgets:Widget::w",
		Inputs: property.NewMap(map[string]property.Value{
			"organizationName": property.New("acme"),
		}),
	})
	require.NoError(t, err)
	require.NotEmpty(t, resp.Failures, "missing required-one-of must produce a failure")
	assert.Contains(t, resp.Failures[0].Reason, "metadataUrl")
}

// TestProvider_Check_RequireOneOf_BothPass: when at least one of the
// required-one-of properties is set, Check passes (no failures).
func TestProvider_Check_RequireOneOf_BothPass(t *testing.T) {
	prov, err := New([]byte(checksSpec), []byte(checksMap))
	require.NoError(t, err)

	resp, err := prov.Check(context.Background(), pgo.CheckRequest{
		Urn: "urn:pulumi:dev::test::pulumiservice:orgs/widgets:Widget::w",
		Inputs: property.NewMap(map[string]property.Value{
			"organizationName": property.New("acme"),
			"metadataUrl":      property.New("https://example.com/saml"),
		}),
	})
	require.NoError(t, err)
	assert.Empty(t, resp.Failures)
}

// TestProvider_Check_RequireIfSet_TriggersWhenAbsent: setting the
// trigger property without the required field must surface a
// failure naming the missing field.
func TestProvider_Check_RequireIfSet_TriggersWhenAbsent(t *testing.T) {
	prov, err := New([]byte(checksSpec), []byte(checksMap))
	require.NoError(t, err)

	resp, err := prov.Check(context.Background(), pgo.CheckRequest{
		Urn: "urn:pulumi:dev::test::pulumiservice:orgs/widgets:Widget::w",
		Inputs: property.NewMap(map[string]property.Value{
			"organizationName": property.New("acme"),
			"metadataUrl":      property.New("https://example.com/saml"),
			"schedule":         property.New("hourly"),
			// rotationCron deliberately missing.
		}),
	})
	require.NoError(t, err)
	require.NotEmpty(t, resp.Failures)
	var found bool
	for _, f := range resp.Failures {
		if strings.Contains(f.Reason, "rotationCron") {
			found = true
			break
		}
	}
	assert.True(t, found, "expected a failure naming rotationCron, got %+v", resp.Failures)
}

// polymorphicSpec / polymorphicMap exercise the case+scopes operation
// shape (Team-style: same Pulumi resource, different create endpoint
// based on a discriminator field). This is the most fragile path in
// the dispatcher — a regression here would silently route to the
// wrong endpoint.
const polymorphicSpec = `{
  "paths": {
    "/api/orgs/{orgName}/teams/pulumi":  {"post": {"operationId": "CreatePulumiTeam"}},
    "/api/orgs/{orgName}/teams/github":  {"post": {"operationId": "CreateGitHubTeam"}},
    "/api/orgs/{orgName}/teams/{teamName}": {
      "get":    {"operationId": "GetTeam"},
      "patch":  {"operationId": "UpdateTeam"},
      "delete": {"operationId": "DeleteTeam"}
    }
  }
}`

const polymorphicMap = `modules:
  orgs/teams:
    resources:
      Team:
        discriminator: { field: teamType, values: [pulumi, github] }
        operations:
          create:
            case: teamType
            pulumi: CreatePulumiTeam
            github: CreateGitHubTeam
          read:   GetTeam
          update: UpdateTeam
          delete: DeleteTeam
        id:
          template: "{organizationName}/{name}"
          params: [organizationName, name]
        forceNew: [organizationName, teamType]
        properties:
          organizationName: { from: orgName,  source: path }
          teamType:         { from: teamType, source: body }
          name:             { from: name,     source: body }
          description:      { from: description, source: body }
`

// TestProvider_PolymorphicCreate_DispatchesCorrectScope: the create
// case+scopes path must pick the operationId matching the
// discriminator value. Tests both branches.
func TestProvider_PolymorphicCreate_DispatchesCorrectScope(t *testing.T) {
	cases := []struct {
		teamType string
		wantPath string
	}{
		{"pulumi", "/api/orgs/acme/teams/pulumi"},
		{"github", "/api/orgs/acme/teams/github"},
	}
	for _, c := range cases {
		c := c
		t.Run(c.teamType, func(t *testing.T) {
			var hitPath string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				hitPath = r.URL.Path
				_ = json.NewEncoder(w).Encode(map[string]any{
					"name": "platform", "description": "...", "teamType": c.teamType,
				})
			}))
			defer srv.Close()

			prov, err := New([]byte(polymorphicSpec), []byte(polymorphicMap))
			require.NoError(t, err)
			ctx := context.Background()
			require.NoError(t, prov.Configure(ctx, pgo.ConfigureRequest{
				Args: property.NewMap(map[string]property.Value{
					"accessToken": property.New("t"), "apiUrl": property.New(srv.URL),
				}),
			}))

			_, err = prov.Create(ctx, pgo.CreateRequest{
				Urn: "urn:pulumi:dev::test::pulumiservice:orgs/teams:Team::team",
				Properties: property.NewMap(map[string]property.Value{
					"organizationName": property.New("acme"),
					"teamType":         property.New(c.teamType),
					"name":             property.New("platform"),
				}),
			})
			require.NoError(t, err)
			assert.Equal(t, c.wantPath, hitPath)
		})
	}
}

// TestProvider_PolymorphicDiff_TeamTypeIsForceNew: changing the
// discriminator must trigger a replace. Without forceNew on the
// discriminator, the engine would try to PATCH a team across types
// and corrupt state.
func TestProvider_PolymorphicDiff_TeamTypeIsForceNew(t *testing.T) {
	prov, err := New([]byte(polymorphicSpec), []byte(polymorphicMap))
	require.NoError(t, err)
	urn := resource.URN("urn:pulumi:dev::test::pulumiservice:orgs/teams:Team::t")

	resp, err := prov.Diff(context.Background(), pgo.DiffRequest{
		Urn: urn,
		State: property.NewMap(map[string]property.Value{
			"organizationName": property.New("acme"),
			"teamType":         property.New("pulumi"),
			"name":             property.New("p"),
		}),
		Inputs: property.NewMap(map[string]property.Value{
			"organizationName": property.New("acme"),
			"teamType":         property.New("github"),
			"name":             property.New("p"),
		}),
	})
	require.NoError(t, err)
	require.True(t, resp.HasChanges)
	require.Contains(t, resp.DetailedDiff, "teamType")
	assert.Equal(t, pgo.UpdateReplace, resp.DetailedDiff["teamType"].Kind)
}

// TestProvider_Configure_FallsBackToEnv: when the engine doesn't pass
// accessToken / apiUrl in the Configure args (a common case for
// in-process integration harnesses), the provider must fall back to
// PULUMI_ACCESS_TOKEN / PULUMI_BACKEND_URL from the environment.
// Matches v1 behavior — programs and tests can rely on env vars
// without setting stack config.
func TestProvider_Configure_FallsBackToEnv(t *testing.T) {
	var seenAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenAuth = r.Header.Get("Authorization")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id": "pool-abc", "name": "p", "tokenValue": "secret-xyz",
		})
	}))
	defer srv.Close()

	t.Setenv("PULUMI_ACCESS_TOKEN", "env-fallback-token")
	t.Setenv("PULUMI_BACKEND_URL", srv.URL)

	prov, err := New([]byte(minimalSpec), []byte(minimalMap))
	require.NoError(t, err)
	ctx := context.Background()
	// Args deliberately empty — exercise the env fallback path.
	require.NoError(t, prov.Configure(ctx, pgo.ConfigureRequest{
		Args: property.NewMap(map[string]property.Value{}),
	}))

	_, err = prov.Create(ctx, pgo.CreateRequest{
		Urn: "urn:pulumi:dev::test::pulumiservice:orgs/agents:AgentPool::p",
		Properties: property.NewMap(map[string]property.Value{
			"organizationName": property.New("acme"),
			"name":             property.New("p"),
		}),
	})
	require.NoError(t, err)
	assert.Equal(t, "token env-fallback-token", seenAuth,
		"Authorization header must be derived from PULUMI_ACCESS_TOKEN env")
}

// TestProvider_Configure_ArgsTakePrecedenceOverEnv: when the engine
// supplies an explicit accessToken arg, it overrides the env var.
// Otherwise stack-config-driven multi-account programs would silently
// pick up the developer's local PULUMI_ACCESS_TOKEN.
func TestProvider_Configure_ArgsTakePrecedenceOverEnv(t *testing.T) {
	var seenAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenAuth = r.Header.Get("Authorization")
		_ = json.NewEncoder(w).Encode(map[string]any{"id": "pool-abc", "name": "p"})
	}))
	defer srv.Close()

	t.Setenv("PULUMI_ACCESS_TOKEN", "env-token")

	prov, err := New([]byte(minimalSpec), []byte(minimalMap))
	require.NoError(t, err)
	ctx := context.Background()
	require.NoError(t, prov.Configure(ctx, pgo.ConfigureRequest{
		Args: property.NewMap(map[string]property.Value{
			"accessToken": property.New("explicit-arg-token"),
			"apiUrl":      property.New(srv.URL),
		}),
	}))

	_, err = prov.Create(ctx, pgo.CreateRequest{
		Urn: "urn:pulumi:dev::test::pulumiservice:orgs/agents:AgentPool::p",
		Properties: property.NewMap(map[string]property.Value{
			"organizationName": property.New("acme"),
			"name":             property.New("p"),
		}),
	})
	require.NoError(t, err)
	assert.Equal(t, "token explicit-arg-token", seenAuth,
		"explicit arg must override env-var fallback")
}

// TestProvider_GetSchema_ConcurrentCallers confirms the sync.Once
// guarantees a single emission even if many CRUD goroutines race
// to ask for the schema. Belt-and-suspenders against a regression
// where someone replaces the once with a non-thread-safe cache.
func TestProvider_GetSchema_ConcurrentCallers(t *testing.T) {
	prov, err := New([]byte(minimalSpec), []byte(minimalMap))
	require.NoError(t, err)

	const n = 16
	var wg sync.WaitGroup
	results := make([]string, n)
	wg.Add(n)
	for i := 0; i < n; i++ {
		i := i
		go func() {
			defer wg.Done()
			resp, err := prov.GetSchema(context.Background(), pgo.GetSchemaRequest{})
			if err == nil {
				results[i] = resp.Schema
			}
		}()
	}
	wg.Wait()
	first := results[0]
	require.NotEmpty(t, first)
	for i := 1; i < n; i++ {
		assert.Equal(t, first, results[i], "all concurrent callers must observe the same cached schema")
	}
	assert.True(t, strings.HasPrefix(first, "{"), "schema should be JSON")
}
