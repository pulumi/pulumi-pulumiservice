// Copyright 2016-2026, Pulumi Corporation.
//
// primitives_test.go — unit coverage for each metadata primitive added
// while eliminating the customresources package. Each test spins a
// narrow httptest server that asserts the request the dispatcher sends,
// then asserts the resource.PropertyMap the dispatcher returns.

package runtime

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ─── bodyOverride ───────────────────────────────────────────────────────

// Delete is rerouted through the update op with a tombstone body (permission:0).
// Mirrors TeamStackPermission's deletion semantics.
func TestPrimitive_BodyOverride_DeleteSendsOverrideBody(t *testing.T) {
	var gotBody map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &gotBody)
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	md := &CloudAPIMetadata{Resources: map[string]CloudAPIResource{
		"pulumiservice:x:R": {
			Token: "pulumiservice:x:R",
			Delete: &CloudAPIOperation{
				OperationID: "Patch", Method: "PATCH",
				PathTemplate: "/perms/{id}",
				BodyOverride: map[string]interface{}{"permission": 0},
			},
			ID:         &CloudAPIID{Template: "{id}", Params: []string{"id"}},
			Properties: map[string]CloudAPIProperty{"id": {Source: "path"}},
		},
	}}
	d := &Dispatcher{Client: NewClient(srv.URL, "tok"), Metadata: md}

	err := d.Delete(context.Background(), "pulumiservice:x:R", "abc",
		resource.PropertyMap{"id": resource.NewStringProperty("abc")})
	require.NoError(t, err)
	assert.Equal(t, float64(0), gotBody["permission"])
	assert.Len(t, gotBody, 1, "only the override body should be sent")
}

// ─── createSource + createFrom (per-verb property source rename) ────────

// Stack-style: stackName is body for POST-at-parent, path for read/delete.
func TestPrimitive_CreateSource_ShiftsBodyToPath(t *testing.T) {
	var createBody map[string]interface{}
	var readPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == "POST":
			body, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(body, &createBody)
			_, _ = w.Write([]byte(`{}`))
		case r.Method == "GET":
			readPath = r.URL.Path
			_, _ = w.Write([]byte(`{}`))
		}
	}))
	defer srv.Close()

	md := &CloudAPIMetadata{Resources: map[string]CloudAPIResource{
		"pulumiservice:x:Stack": {
			Token: "pulumiservice:x:Stack",
			Create: &CloudAPIOperation{
				OperationID: "Create", Method: "POST",
				PathTemplate: "/stacks/{org}",
			},
			Read: &CloudAPIOperation{
				OperationID: "Read", Method: "GET",
				PathTemplate: "/stacks/{org}/{stackName}",
			},
			ID: &CloudAPIID{
				Template: "{org}/{stackName}", Params: []string{"org", "stackName"},
			},
			Properties: map[string]CloudAPIProperty{
				"org":       {Source: "path", From: "org"},
				"stackName": {Source: "path", From: "stackName", CreateSource: "body", CreateFrom: "stackName"},
			},
		},
	}}
	d := &Dispatcher{Client: NewClient(srv.URL, "tok"), Metadata: md}

	_, _, err := d.Create(context.Background(), "pulumiservice:x:Stack", resource.PropertyMap{
		"org":       resource.NewStringProperty("acme"),
		"stackName": resource.NewStringProperty("prod"),
	})
	require.NoError(t, err)
	assert.Equal(t, "prod", createBody["stackName"], "stackName should appear in create body")

	_, _, err = d.Read(context.Background(), "pulumiservice:x:Stack", "acme/prod", resource.PropertyMap{})
	require.NoError(t, err)
	assert.Equal(t, "/stacks/acme/prod", readPath, "stackName should be path param on read")
}

// ─── rawBodyFrom / rawBodyTo / contentType ──────────────────────────────

// Update sends a raw application/x-yaml body pulled from a property and
// reads it back from a raw response, wrapping the result as secret.
// Mirrors ESC Environment's YAML PATCH/GET.
func TestPrimitive_RawBody_RoundTripWithSecretWrapping(t *testing.T) {
	const yamlDoc = "values:\n  hello: world\n"
	var gotCT, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "PATCH":
			gotCT = r.Header.Get("Content-Type")
			b, _ := io.ReadAll(r.Body)
			gotBody = string(b)
			w.WriteHeader(200)
		case "GET":
			w.Header().Set("Content-Type", "application/x-yaml")
			_, _ = w.Write([]byte(yamlDoc))
		}
	}))
	defer srv.Close()

	md := &CloudAPIMetadata{Resources: map[string]CloudAPIResource{
		"pulumiservice:x:Env": {
			Token: "pulumiservice:x:Env",
			Read: &CloudAPIOperation{
				OperationID: "Read", Method: "GET",
				PathTemplate: "/envs/{name}",
				RawBodyTo:    "yaml",
				ContentType:  "application/x-yaml",
			},
			Update: &CloudAPIOperation{
				OperationID: "Update", Method: "PATCH",
				PathTemplate: "/envs/{name}",
				RawBodyFrom:  "yaml",
				ContentType:  "application/x-yaml",
			},
			ID: &CloudAPIID{Template: "{name}", Params: []string{"name"}},
			Properties: map[string]CloudAPIProperty{
				"name": {Source: "path"},
				"yaml": {Source: "rawBody", Secret: true},
			},
		},
	}}
	d := &Dispatcher{Client: NewClient(srv.URL, "tok"), Metadata: md}

	_, err := d.Update(context.Background(), "pulumiservice:x:Env", "prod",
		resource.PropertyMap{"name": resource.NewStringProperty("prod")},
		resource.PropertyMap{
			"name": resource.NewStringProperty("prod"),
			"yaml": resource.MakeSecret(resource.NewStringProperty(yamlDoc)),
		})
	require.NoError(t, err)
	assert.Equal(t, "application/x-yaml", gotCT)
	assert.Equal(t, yamlDoc, gotBody, "rawBodyFrom must bypass JSON encoding")

	_, outs, err := d.Read(context.Background(), "pulumiservice:x:Env", "prod",
		resource.PropertyMap{"name": resource.NewStringProperty("prod")})
	require.NoError(t, err)
	require.Contains(t, outs, resource.PropertyKey("yaml"))
	require.True(t, outs["yaml"].IsSecret(), "rawBodyTo on a secret property must wrap the value")
	assert.Equal(t, yamlDoc, outs["yaml"].SecretValue().Element.StringValue())
}

// ─── iterateOver (delete iterates per map key, in sorted order) ─────────

func TestPrimitive_IterateOver_DeletesEachKeyInSortedOrder(t *testing.T) {
	var mu sync.Mutex
	var deletedInOrder []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		// Path: /tags/{tagName}
		parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/tags/"), "/")
		deletedInOrder = append(deletedInOrder, parts[0])
		w.WriteHeader(204)
	}))
	defer srv.Close()

	md := &CloudAPIMetadata{Resources: map[string]CloudAPIResource{
		"pulumiservice:x:Tags": {
			Token: "pulumiservice:x:Tags",
			Delete: &CloudAPIOperation{
				OperationID:     "Del", Method: "DELETE",
				PathTemplate:    "/tags/{tagName}",
				IterateOver:     "tags",
				IterateKeyParam: "tagName",
			},
			ID:         &CloudAPIID{Template: "x", Params: nil},
			Properties: map[string]CloudAPIProperty{"tags": {Source: "body"}},
		},
	}}
	d := &Dispatcher{Client: NewClient(srv.URL, "tok"), Metadata: md}

	// Seed a tag map with keys deliberately out of alphabetic order; the
	// dispatcher should still iterate them in sorted order.
	state := resource.PropertyMap{
		"tags": resource.NewObjectProperty(resource.PropertyMap{
			"zebra":  resource.NewStringProperty("z"),
			"alpha":  resource.NewStringProperty("a"),
			"mango":  resource.NewStringProperty("m"),
		}),
	}
	err := d.Delete(context.Background(), "pulumiservice:x:Tags", "x", state)
	require.NoError(t, err)
	assert.Equal(t, []string{"alpha", "mango", "zebra"}, deletedInOrder)
}

// ─── readVia.extractField + keyBy (read a child from a parent's GET) ────

// Single-entry mode: Tag is stored on the parent stack's response under
// response.tags[name]. Secret valueProperty must wrap the extracted value.
func TestPrimitive_ReadViaExtractField_KeyByWrapsSecretValue(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"tags": map[string]interface{}{
				"env":     "prod",
				"secret":  "s3cr3t-value",
			},
		})
	}))
	defer srv.Close()

	parent := &CloudAPIOperation{
		OperationID: "GetStack", Method: "GET",
		PathTemplate: "/stacks/{stack}",
	}
	md := &CloudAPIMetadata{Resources: map[string]CloudAPIResource{
		"pulumiservice:x:Parent": {
			Token: "pulumiservice:x:Parent",
			Read:  parent,
			ID:    &CloudAPIID{Template: "{stack}", Params: []string{"stack"}},
			Properties: map[string]CloudAPIProperty{"stack": {Source: "path"}},
		},
		"pulumiservice:x:Tag": {
			Token: "pulumiservice:x:Tag",
			ReadVia: &CloudAPIReadVia{
				OperationID:   "GetStack",
				ExtractField:  "tags",
				KeyBy:         "name",
				ValueProperty: "value",
			},
			ID: &CloudAPIID{
				Template: "{stack}/{name}", Params: []string{"stack", "name"},
			},
			Properties: map[string]CloudAPIProperty{
				"stack": {Source: "path"},
				"name":  {Source: "path"},
				"value": {Source: "body", Secret: true},
			},
		},
	}}
	d := &Dispatcher{Client: NewClient(srv.URL, "tok"), Metadata: md}

	_, outs, err := d.Read(context.Background(), "pulumiservice:x:Tag", "prod-stack/secret",
		resource.PropertyMap{
			"stack": resource.NewStringProperty("prod-stack"),
			"name":  resource.NewStringProperty("secret"),
		})
	require.NoError(t, err)
	require.Contains(t, outs, resource.PropertyKey("value"))
	require.True(t, outs["value"].IsSecret(), "secret value property must wrap the extracted value")
	assert.Equal(t, "s3cr3t-value", outs["value"].SecretValue().Element.StringValue())
}

// Whole-map mode: Tags returns the full tag map under the extractField name.
// Missing key → nil outputs (signals recreate).
func TestPrimitive_ReadViaExtractField_WholeMap(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"tags": map[string]interface{}{"env": "prod", "team": "platform"},
		})
	}))
	defer srv.Close()

	parent := &CloudAPIOperation{
		OperationID: "GetStack", Method: "GET",
		PathTemplate: "/stacks/{stack}",
	}
	md := &CloudAPIMetadata{Resources: map[string]CloudAPIResource{
		"pulumiservice:x:Parent": {
			Token: "pulumiservice:x:Parent", Read: parent,
			ID:         &CloudAPIID{Template: "{stack}", Params: []string{"stack"}},
			Properties: map[string]CloudAPIProperty{"stack": {Source: "path"}},
		},
		"pulumiservice:x:Tags": {
			Token: "pulumiservice:x:Tags",
			ReadVia: &CloudAPIReadVia{
				OperationID:  "GetStack",
				ExtractField: "tags",
			},
			ID: &CloudAPIID{Template: "{stack}", Params: []string{"stack"}},
			Properties: map[string]CloudAPIProperty{
				"stack": {Source: "path"},
				"tags":  {Source: "body"},
			},
		},
	}}
	d := &Dispatcher{Client: NewClient(srv.URL, "tok"), Metadata: md}

	_, outs, err := d.Read(context.Background(), "pulumiservice:x:Tags", "prod-stack",
		resource.PropertyMap{"stack": resource.NewStringProperty("prod-stack")})
	require.NoError(t, err)
	require.Contains(t, outs, resource.PropertyKey("tags"))
	tagsMap := outs["tags"].ObjectValue()
	assert.Equal(t, "prod", tagsMap["env"].StringValue())
	assert.Equal(t, "platform", tagsMap["team"].StringValue())
}

// ─── Write-once secret outputs persist across refresh ──────────────────

// An access-token-style property: the Create response carries
// tokenValue, but subsequent List (used as readVia) responses do not.
// ResponseToOutputs must fall back to the prior state for secret
// output properties so the secret doesn't silently vanish on refresh.
func TestPrimitive_WriteOnceSecretOutput_PreservedFromPriorState(t *testing.T) {
	props := map[string]CloudAPIProperty{
		"description": {Source: "body"},
		"tokenId":     {From: "id", Source: "response", Output: true},
		"value":       {From: "tokenValue", Source: "response", Output: true, Secret: true},
	}
	// Server response (list entry) omits tokenValue — typical of list
	// endpoints that don't re-surface secrets.
	response := map[string]interface{}{
		"id":          "tok-abc",
		"description": "CI deploy token",
	}
	// Prior state — from the previous refresh/up — still carries the
	// secret. This is what Pulumi passes as originalInputs on Read.
	prior := resource.PropertyMap{
		"description": resource.NewStringProperty("CI deploy token"),
		"tokenId":     resource.NewStringProperty("tok-abc"),
		"value":       resource.MakeSecret(resource.NewStringProperty("secret-xyz")),
	}

	out := ResponseToOutputs(props, response, prior)

	require.Contains(t, out, resource.PropertyKey("value"))
	require.True(t, out["value"].IsSecret(), "preserved value must remain secret-wrapped")
	assert.Equal(t, "secret-xyz", out["value"].SecretValue().Element.StringValue())
}

// Stable-sort for complex object arrays: two arrays of the same objects
// in different orders must sort to byte-identical results, even though
// map iteration inside fmt.Sprintf is nondeterministic. Guards against
// regressing from json.Marshal back to %v.
func TestPrimitive_SortOnRead_StableForObjectsWithMapFields(t *testing.T) {
	obj := func(a, b string) resource.PropertyValue {
		return resource.NewObjectProperty(resource.PropertyMap{
			"first":  resource.NewStringProperty(a),
			"second": resource.NewStringProperty(b),
		})
	}
	props := map[string]CloudAPIProperty{"policies": {SortOnRead: true}}
	inputs1 := resource.PropertyMap{
		"policies": resource.NewArrayProperty([]resource.PropertyValue{
			obj("y", "1"), obj("a", "2"), obj("m", "3"),
		}),
	}
	inputs2 := resource.PropertyMap{
		"policies": resource.NewArrayProperty([]resource.PropertyValue{
			obj("m", "3"), obj("y", "1"), obj("a", "2"),
		}),
	}

	// Run the sort twice over the different-order inputs; expected
	// canonical form is identical.
	out1 := CanonicalizeSortedInputs(props, inputs1)
	out2 := CanonicalizeSortedInputs(props, inputs2)

	serialize := func(pm resource.PropertyMap) string {
		arr := pm["policies"].ArrayValue()
		keys := make([]string, len(arr))
		for i, v := range arr {
			keys[i] = v.ObjectValue()["first"].StringValue()
		}
		return fmt.Sprintf("%v", keys)
	}
	assert.Equal(t, serialize(out1), serialize(out2),
		"equal object arrays in different input orders must produce the same canonical order")
	assert.Equal(t, "[a m y]", serialize(out1))
}

// ─── sortOnRead canonicalizes inputs too (anti-drift) ───────────────────

// User writes an array in arbitrary order; after Check, inputs are in
// the same canonical order the server will sort to on Read, so diff
// doesn't see spurious reordering.
func TestPrimitive_SortOnRead_CanonicalizesInputs(t *testing.T) {
	props := map[string]CloudAPIProperty{
		"members": {SortOnRead: true},
		"other":   {},
	}
	inputs := resource.PropertyMap{
		"members": resource.NewArrayProperty([]resource.PropertyValue{
			resource.NewStringProperty("zeta"),
			resource.NewStringProperty("alpha"),
			resource.NewStringProperty("mu"),
		}),
		"other": resource.NewStringProperty("unchanged"),
	}

	assert.True(t, HasSortOnRead(props))

	out := CanonicalizeSortedInputs(props, inputs)
	sorted := out["members"].ArrayValue()
	require.Len(t, sorted, 3)
	assert.Equal(t, "alpha", sorted[0].StringValue())
	assert.Equal(t, "mu", sorted[1].StringValue())
	assert.Equal(t, "zeta", sorted[2].StringValue())

	// Non-sorted property untouched.
	assert.Equal(t, "unchanged", out["other"].StringValue())
}

// Secrets wrapping a sorted array: sort occurs inside the secret wrap,
// so the property remains secret after canonicalization.
func TestPrimitive_SortOnRead_PreservesSecretWrap(t *testing.T) {
	props := map[string]CloudAPIProperty{
		"members": {SortOnRead: true, Secret: true},
	}
	inputs := resource.PropertyMap{
		"members": resource.MakeSecret(resource.NewArrayProperty([]resource.PropertyValue{
			resource.NewStringProperty("zeta"),
			resource.NewStringProperty("alpha"),
		})),
	}

	out := CanonicalizeSortedInputs(props, inputs)
	require.True(t, out["members"].IsSecret())
	inner := out["members"].SecretValue().Element.ArrayValue()
	assert.Equal(t, "alpha", inner[0].StringValue())
	assert.Equal(t, "zeta", inner[1].StringValue())
}

// ─── Invoke (data-source function) ──────────────────────────────────────

func TestPrimitive_Invoke_ListResponseWrappedInItems(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/orgs/acme/agent-pools", r.URL.Path)
		_, _ = w.Write([]byte(`[{"id":"pool-1"},{"id":"pool-2"}]`))
	}))
	defer srv.Close()

	md := &CloudAPIMetadata{Functions: map[string]CloudAPIFunction{
		"pulumiservice:orgs/agents:listAgentPools": {
			Token: "pulumiservice:orgs/agents:listAgentPools",
			Operation: CloudAPIOperation{
				OperationID: "ListOrgAgentPool", Method: "GET",
				PathTemplate: "/api/orgs/{orgName}/agent-pools",
			},
		},
	}}
	d := &Dispatcher{Client: NewClient(srv.URL, "tok"), Metadata: md}

	out, err := d.Invoke(context.Background(), "pulumiservice:orgs/agents:listAgentPools",
		resource.PropertyMap{"orgName": resource.NewStringProperty("acme")})
	require.NoError(t, err)
	require.Contains(t, out, resource.PropertyKey("items"))
	items := out["items"].ArrayValue()
	require.Len(t, items, 2)
	assert.Equal(t, "pool-1", items[0].ObjectValue()["id"].StringValue())
}

func TestPrimitive_Invoke_ObjectResponsePassedThrough(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"decision":"allow","reason":"trusted sub"}`))
	}))
	defer srv.Close()

	md := &CloudAPIMetadata{Functions: map[string]CloudAPIFunction{
		"pulumiservice:orgs/oidc:getAuthPolicy": {
			Token: "pulumiservice:orgs/oidc:getAuthPolicy",
			Operation: CloudAPIOperation{
				OperationID: "GetAuthPolicy", Method: "GET",
				PathTemplate: "/api/orgs/{orgName}/oidc/issuers/{issuerId}/policy",
			},
		},
	}}
	d := &Dispatcher{Client: NewClient(srv.URL, "tok"), Metadata: md}

	out, err := d.Invoke(context.Background(), "pulumiservice:orgs/oidc:getAuthPolicy",
		resource.PropertyMap{
			"orgName":  resource.NewStringProperty("acme"),
			"issuerId": resource.NewStringProperty("iss-1"),
		})
	require.NoError(t, err)
	assert.Equal(t, "allow", out["decision"].StringValue())
	assert.Equal(t, "trusted sub", out["reason"].StringValue())
}

// ─── postCreate (two-step create) ───────────────────────────────────────

// Environment-style: POST creates the empty resource, a follow-on PATCH
// with a raw body fills in the content. Both calls must fire, in order.
func TestPrimitive_PostCreate_RunsFollowOnOp(t *testing.T) {
	var mu sync.Mutex
	var calls []string
	var postCreateBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		calls = append(calls, r.Method+" "+r.URL.Path)
		switch r.Method {
		case "POST":
			_, _ = w.Write([]byte(`{}`))
		case "PATCH":
			b, _ := io.ReadAll(r.Body)
			postCreateBody = string(b)
			w.WriteHeader(200)
		}
	}))
	defer srv.Close()

	md := &CloudAPIMetadata{Resources: map[string]CloudAPIResource{
		"pulumiservice:x:Env": {
			Token: "pulumiservice:x:Env",
			Create: &CloudAPIOperation{
				OperationID: "Create", Method: "POST", PathTemplate: "/envs/{org}",
			},
			PostCreate: &CloudAPIOperation{
				OperationID: "Update", Method: "PATCH", PathTemplate: "/envs/{org}/{name}",
				RawBodyFrom: "yaml", ContentType: "application/x-yaml",
			},
			ID: &CloudAPIID{Template: "{org}/{name}", Params: []string{"org", "name"}},
			Properties: map[string]CloudAPIProperty{
				"org":  {Source: "path", From: "org"},
				"name": {Source: "path", CreateSource: "body", CreateFrom: "name"},
				"yaml": {Source: "rawBody"},
			},
		},
	}}
	d := &Dispatcher{Client: NewClient(srv.URL, "tok"), Metadata: md}

	id, _, err := d.Create(context.Background(), "pulumiservice:x:Env", resource.PropertyMap{
		"org":  resource.NewStringProperty("acme"),
		"name": resource.NewStringProperty("prod"),
		"yaml": resource.NewStringProperty("values: {}\n"),
	})
	require.NoError(t, err)
	assert.Equal(t, "acme/prod", id)
	require.Equal(t, []string{
		"POST /envs/acme",             // initial create
		"PATCH /envs/acme/prod",       // post-create follow-on
	}, calls)
	assert.Equal(t, "values: {}\n", postCreateBody)
}

// ─── requireIfSet (Check) ───────────────────────────────────────────────

// A required field gated on a trigger field being set: when the trigger
// is absent the rule is silent, when the trigger is present the gated
// field must also be present. Mirrors Webhook's projectName-when-
// stackName check — stack-scoped webhooks need a project path; ESC-
// scoped webhooks need it too but for a different scope selector; org-
// scoped webhooks need neither.
func TestPrimitive_RequireIfSet_TriggerAbsent_Silent(t *testing.T) {
	res := &CloudAPIResource{
		Checks: []CheckRule{{
			RequireIfSet: "stackName",
			Field:        "projectName",
		}},
	}
	failures := EvaluateChecks(res, resource.PropertyMap{
		"organizationName": resource.NewStringProperty("acme"),
	})
	assert.Empty(t, failures, "rule should be silent when trigger field is unset")
}

func TestPrimitive_RequireIfSet_TriggerPresent_FieldMissing_Fails(t *testing.T) {
	res := &CloudAPIResource{
		Checks: []CheckRule{{
			RequireIfSet: "stackName",
			Field:        "projectName",
		}},
	}
	failures := EvaluateChecks(res, resource.PropertyMap{
		"organizationName": resource.NewStringProperty("acme"),
		"stackName":        resource.NewStringProperty("prod"),
	})
	require.Len(t, failures, 1)
	assert.Equal(t, "projectName", failures[0].Property)
}

func TestPrimitive_RequireIfSet_TriggerPresent_FieldPresent_Passes(t *testing.T) {
	res := &CloudAPIResource{
		Checks: []CheckRule{{
			RequireIfSet: "stackName",
			Field:        "projectName",
		}},
	}
	failures := EvaluateChecks(res, resource.PropertyMap{
		"organizationName": resource.NewStringProperty("acme"),
		"stackName":        resource.NewStringProperty("prod"),
		"projectName":      resource.NewStringProperty("infra"),
	})
	assert.Empty(t, failures)
}
