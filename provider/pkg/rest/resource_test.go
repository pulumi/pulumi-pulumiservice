// Copyright 2016-2026, Pulumi Corporation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package rest

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi/sdk/v3/go/property"
)

// mockTransport returns canned responses keyed by HTTP method+path. Path
// matching is exact against the request URL's path component. Tests that
// need stateful behavior (e.g., GET returns 404 before PUT, 200 after) can
// set responseFn instead, which is called for every request and overrides
// the responses map.
type mockTransport struct {
	responses  map[string]mockResponse
	responseFn func(req *http.Request) mockResponse
	calls      []string // method + " " + path, in order
}

type mockResponse struct {
	status int
	body   string
}

func (m *mockTransport) Do(_ context.Context, req *http.Request) (*http.Response, error) {
	key := req.Method + " " + req.URL.Path
	m.calls = append(m.calls, key)
	if m.responseFn != nil {
		resp := m.responseFn(req)
		return &http.Response{
			StatusCode: resp.status,
			Body:       io.NopCloser(bytes.NewReader([]byte(resp.body))),
		}, nil
	}
	resp, ok := m.responses[key]
	if !ok {
		return &http.Response{
			StatusCode: 500,
			Body:       io.NopCloser(strings.NewReader("no mock for " + key)),
		}, nil
	}
	return &http.Response{
		StatusCode: resp.status,
		Body:       io.NopCloser(bytes.NewReader([]byte(resp.body))),
	}, nil
}

// loadFixtures reads the embedded spec.json and metadata.json from the
// cloud package fixtures.
func loadFixtures(t *testing.T) (*Spec, *Metadata) {
	t.Helper()
	_, here, _, _ := runtime.Caller(0)
	cloudDir := filepath.Join(filepath.Dir(here), "..", "cloud")
	specBytes, err := os.ReadFile(filepath.Join(cloudDir, "spec.json"))
	if err != nil {
		t.Fatalf("read spec.json: %v", err)
	}
	metaBytes, err := os.ReadFile(filepath.Join(cloudDir, "metadata.json"))
	if err != nil {
		t.Fatalf("read metadata.json: %v", err)
	}
	spec, err := ParseSpec(specBytes)
	if err != nil {
		t.Fatalf("parse spec: %v", err)
	}
	meta, err := ParseMetadata(metaBytes)
	if err != nil {
		t.Fatalf("parse metadata: %v", err)
	}
	return spec, meta
}

func propMap(m map[string]any) property.Map {
	out := make(map[string]property.Value, len(m))
	for k, v := range m {
		out[k] = anyToPropertyValue(v)
	}
	return property.NewMap(out)
}

// TestCreateSynthesizesID exercises Create end-to-end against the real
// spec for representative resources, verifying the synthesized ID matches
// expected composite-from-path-params form. Covers the four interesting
// shapes:
//
//   - server-generated id (AgentPool, Role): create response carries id;
//     ID = {orgName}/{serverID}
//   - 204 no-content (StackTag): all path-param values come from inputs;
//     ID = {orgName}/{projectName}/{stackName}/{tagName}
//   - composite from path (Team): create response includes resource body;
//     ID = {orgName}/{teamName}
//   - singleton (DefaultOrganization, AuditLogExportConfiguration): ID =
//     {orgName}
func TestCreateSynthesizesID(t *testing.T) {
	spec, meta := loadFixtures(t)
	resources := Resources(spec, meta)

	cases := []struct {
		// Lookup token (the metadata.json key).
		token string
		// Mock HTTP responses keyed by "<METHOD> <path>". Mutually exclusive
		// with responseFn.
		responses map[string]mockResponse
		// responseFn, if set, overrides responses for stateful behavior
		// (e.g., GET returns 404 before mutating call, 200 after — needed
		// for resources with requireImport, where the read op fires both
		// as a pre-flight probe and as read-after-create).
		responseFn func() func(req *http.Request) mockResponse
		// Inputs supplied by the user.
		inputs map[string]any
		// Expected resource ID after Create.
		wantID string
	}{
		{
			// AgentPool: server-generated id, response renamed `id`→`poolId`.
			// Old behavior: ID = "<uuid>".
			// New behavior: ID = "<orgName>/<uuid>".
			token: "pulumiservice:v2:AgentPool",
			responses: map[string]mockResponse{
				"POST /api/orgs/test-org/agent-pools": {
					status: 200,
					body:   `{"id":"abc-123","name":"runners","description":"hi"}`,
				},
				"GET /api/orgs/test-org/agent-pools/abc-123": {
					status: 200,
					body:   `{"id":"abc-123","name":"runners","description":"hi"}`,
				},
			},
			inputs: map[string]any{
				"orgName":     "test-org",
				"name":        "runners",
				"description": "hi",
			},
			wantID: "test-org/abc-123",
		},
		{
			// StackTag: 204 No Content, identity entirely from inputs. No
			// read op declared, so no GET fires after create.
			token: "pulumiservice:v2:StackTag",
			responses: map[string]mockResponse{
				"POST /api/stacks/test-org/myproj/mystack/tags": {status: 204, body: ""},
			},
			inputs: map[string]any{
				"orgName":     "test-org",
				"projectName": "myproj",
				"stackName":   "mystack",
				"name":        "owner", // body field; rename name→tagName for path
				"value":       "team-x",
			},
			wantID: "test-org/myproj/mystack/owner",
		},
		{
			// Team: response body has resource fields, composite identity from path.
			token: "pulumiservice:v2:Team",
			responses: map[string]mockResponse{
				"POST /api/orgs/test-org/teams/pulumi": {
					status: 200,
					body:   `{"name":"infra","description":"infra team"}`,
				},
				"GET /api/orgs/test-org/teams/infra": {
					status: 200,
					body:   `{"name":"infra","description":"infra team"}`,
				},
			},
			inputs: map[string]any{
				"orgName":     "test-org",
				"name":        "infra", // body; renames map name→teamName for path
				"description": "infra team",
			},
			wantID: "test-org/infra",
		},
		{
			// DefaultOrganization: requireImport singleton — GET returns 404
			// for the probe (resource doesn't exist), then 200 for the post-
			// create read after the mutating call writes it.
			token: "pulumiservice:v2:DefaultOrganization",
			responseFn: func() func(req *http.Request) mockResponse {
				written := false
				return func(req *http.Request) mockResponse {
					switch {
					case req.Method == "GET" && !written:
						return mockResponse{status: 404, body: `{"error":"not found"}`}
					case req.Method == "GET":
						return mockResponse{status: 200, body: `{"orgName":"test-org"}`}
					case req.URL.Path == "/api/user/organizations/test-org/default":
						written = true
						return mockResponse{status: 200, body: `{}`}
					}
					return mockResponse{status: 500, body: "unexpected"}
				}
			},
			inputs: map[string]any{
				"orgName": "test-org",
			},
			wantID: "test-org",
		},
		{
			// AuditLogExportConfiguration: requireImport singleton, all ops
			// on the same path — same staging pattern as DefaultOrganization.
			token: "pulumiservice:v2:AuditLogExportConfiguration",
			responseFn: func() func(req *http.Request) mockResponse {
				written := false
				return func(req *http.Request) mockResponse {
					switch {
					case req.Method == "GET" && !written:
						return mockResponse{status: 404, body: `{"error":"not found"}`}
					case req.Method == "GET":
						return mockResponse{status: 200, body: `{}`}
					case req.Method == "POST":
						written = true
						return mockResponse{status: 200, body: `{}`}
					}
					return mockResponse{status: 500, body: "unexpected"}
				}
			},
			inputs: map[string]any{
				"orgName": "test-org",
			},
			wantID: "test-org",
		},
		{
			// Role: server-generated id, response renamed `id`→`roleID`.
			token: "pulumiservice:v2:Role",
			responses: map[string]mockResponse{
				"POST /api/orgs/test-org/roles": {
					status: 200,
					body:   `{"id":"role-7","name":"deployer","description":"deploy access"}`,
				},
				"GET /api/orgs/test-org/roles/role-7": {
					status: 200,
					body:   `{"id":"role-7","name":"deployer","description":"deploy access"}`,
				},
			},
			inputs: map[string]any{
				"orgName":     "test-org",
				"name":        "deployer",
				"description": "deploy access",
			},
			wantID: "test-org/role-7",
		},
	}

	for _, tc := range cases {
		t.Run(tc.token, func(t *testing.T) {
			r, ok := resources[tc.token]
			if !ok {
				// Try with the user-facing token as well.
				rm, exists := meta.Resources[tc.token]
				if !exists {
					t.Fatalf("resource %q not in metadata", tc.token)
				}
				if rm.Token != "" {
					r = resources[rm.Token]
				}
			}
			if r == nil {
				t.Fatalf("resource %q not in factory output", tc.token)
			}

			mock := &mockTransport{responses: tc.responses}
			if tc.responseFn != nil {
				mock.responseFn = tc.responseFn()
			}
			SetTransportResolver(func(_ context.Context) (Transport, error) {
				return mock, nil
			})

			req := p.CreateRequest{Properties: propMap(tc.inputs)}
			resp, err := r.Create(context.Background(), req)
			if err != nil {
				t.Fatalf("create: %v\n  calls made: %v", err, mock.calls)
			}
			if resp.ID != tc.wantID {
				t.Errorf("ID:\n  got:  %q\n  want: %q", resp.ID, tc.wantID)
			}
			t.Logf("OK %s -> ID=%q (calls: %v)", tc.token, resp.ID, mock.calls)
		})
	}
}

// TestCreateReadAfterCreateSourcesFromInputs confirms that the read-after-
// create URL is built from the user inputs (req.Properties), not from the
// (potentially sparse) create response. This is what lets Phase D drop path
// params from state without breaking the read-after-create round-trip: the
// read URL substitution still finds path-param values via inputs, even when
// the create response carried only a server-assigned id.
func TestCreateReadAfterCreateSourcesFromInputs(t *testing.T) {
	const specJSON = `{
	  "openapi": "3.0.0",
	  "components": {"schemas": {
	    "Body":  {"type": "object", "properties": {"name": {"type": "string"}}},
	    "Read":  {"type": "object", "properties": {
	      "id":     {"type": "string"},
	      "name":   {"type": "string"},
	      "status": {"type": "string"}
	    }}
	  }},
	  "paths": {
	    "/things/{org}": {
	      "post": {
	        "operationId": "CreateThing",
	        "parameters": [{"name": "org", "in": "path", "required": true, "schema": {"type": "string"}}],
	        "requestBody": {"content": {"application/json": {"schema": {"$ref": "#/components/schemas/Body"}}}},
	        "responses": {"200": {"content": {"application/json": {"schema": {"type": "object", "properties": {"id": {"type": "string"}}}}}}}
	      }
	    },
	    "/things/{org}/{id}": {
	      "get": {
	        "operationId": "GetThing",
	        "parameters": [
	          {"name": "org", "in": "path", "required": true, "schema": {"type": "string"}},
	          {"name": "id",  "in": "path", "required": true, "schema": {"type": "string"}}
	        ],
	        "responses": {"200": {"content": {"application/json": {"schema": {"$ref": "#/components/schemas/Read"}}}}}
	      }
	    }
	  }
	}`
	spec, _ := ParseSpec([]byte(specJSON))
	r := &Resource{
		spec: spec,
		meta: ResourceMeta{
			Operations: Operations{Create: "CreateThing", Read: "GetThing"},
			IDFormat:   "{org}/{id}",
		},
	}
	mock := &mockTransport{responses: map[string]mockResponse{
		"POST /things/acme":        {status: 200, body: `{"id":"thing-1"}`},
		"GET /things/acme/thing-1": {status: 200, body: `{"id":"thing-1","name":"foo","status":"ready"}`},
	}}
	SetTransportResolver(func(_ context.Context) (Transport, error) { return mock, nil })

	resp, err := r.Create(context.Background(), p.CreateRequest{
		Properties: propMap(map[string]any{"org": "acme", "name": "foo"}),
	})
	if err != nil {
		t.Fatalf("create: %v\n  calls: %v", err, mock.calls)
	}
	if len(mock.calls) != 2 || mock.calls[1] != "GET /things/acme/thing-1" {
		t.Errorf("expected POST then GET /things/acme/thing-1, got: %v", mock.calls)
	}
	if v, ok := resp.Properties.GetOk("status"); !ok || v.AsString() != "ready" {
		t.Errorf("state missing read-only field 'status': ok=%v v=%q", ok, v.AsString())
	}
}

// TestCreateReadsAfterCreate verifies that Create fires the read op after
// the create call and returns state populated from the read response, not
// just whatever sparse body the create endpoint echoed. Many Pulumi Cloud
// create endpoints return `{}` or a stripped object; without read-after-
// create, downstream resources referencing read-only outputs would dereference
// missing values until the next refresh.
func TestCreateReadsAfterCreate(t *testing.T) {
	// Synthetic spec: create returns just `{id}`, read returns rich state.
	const specJSON = `{
	  "openapi": "3.0.0",
	  "components": {"schemas": {
	    "Body":  {"type": "object", "properties": {"name": {"type": "string"}}},
	    "Read":  {"type": "object", "properties": {
	      "id":         {"type": "string"},
	      "name":       {"type": "string"},
	      "lastUpdate": {"type": "string"},
	      "status":     {"type": "string"}
	    }}
	  }},
	  "paths": {
	    "/things/{org}": {
	      "post": {
	        "operationId": "CreateThing",
	        "parameters": [{"name": "org", "in": "path", "required": true, "schema": {"type": "string"}}],
	        "requestBody": {"content": {"application/json": {"schema": {"$ref": "#/components/schemas/Body"}}}},
	        "responses": {"200": {"content": {"application/json": {"schema": {"type": "object", "properties": {"id": {"type": "string"}}}}}}}
	      }
	    },
	    "/things/{org}/{id}": {
	      "get": {
	        "operationId": "GetThing",
	        "parameters": [
	          {"name": "org", "in": "path", "required": true, "schema": {"type": "string"}},
	          {"name": "id",  "in": "path", "required": true, "schema": {"type": "string"}}
	        ],
	        "responses": {"200": {"content": {"application/json": {"schema": {"$ref": "#/components/schemas/Read"}}}}}
	      }
	    }
	  }
	}`
	spec, err := ParseSpec([]byte(specJSON))
	if err != nil {
		t.Fatalf("parse synthetic spec: %v", err)
	}
	r := &Resource{
		spec: spec,
		meta: ResourceMeta{
			Operations: Operations{Create: "CreateThing", Read: "GetThing"},
			IDFormat:   "{org}/{id}",
		},
	}
	mock := &mockTransport{responses: map[string]mockResponse{
		"POST /things/acme":           {status: 200, body: `{"id":"thing-1"}`},
		"GET /things/acme/thing-1":    {status: 200, body: `{"id":"thing-1","name":"foo","lastUpdate":"2026-05-05T00:00:00Z","status":"ready"}`},
	}}
	SetTransportResolver(func(_ context.Context) (Transport, error) { return mock, nil })

	resp, err := r.Create(context.Background(), p.CreateRequest{
		Properties: propMap(map[string]any{"org": "acme", "name": "foo"}),
	})
	if err != nil {
		t.Fatalf("create: %v\n  calls: %v", err, mock.calls)
	}

	// The GET must have fired after the POST.
	if len(mock.calls) != 2 || mock.calls[0] != "POST /things/acme" || mock.calls[1] != "GET /things/acme/thing-1" {
		t.Errorf("expected POST then GET, got: %v", mock.calls)
	}
	// State should now carry the read-only fields, not just the create echo.
	for _, key := range []string{"lastUpdate", "status"} {
		v, ok := resp.Properties.GetOk(key)
		if !ok {
			t.Errorf("state missing %q after read-after-create", key)
			continue
		}
		if v.AsString() == "" {
			t.Errorf("state[%q] is empty; want value from read response", key)
		}
	}
	// Path params are program-owned: they belong in inputs and the resource ID,
	// not in cloud-owned state. After read-after-create, state should carry only
	// what the read response returned, plus any emit-on-create preserves.
	if _, ok := resp.Properties.GetOk("org"); ok {
		t.Errorf("state should not carry path-param `org` (program owns inputs, cloud owns outputs)")
	}
	// ID is unchanged by read.
	if resp.ID != "acme/thing-1" {
		t.Errorf("ID: got %q, want %q", resp.ID, "acme/thing-1")
	}
}

// TestCreateRequireImport_BlocksWhenExists verifies that Create with
// RequireImport=true issues the read op first and aborts when the resource
// already exists upstream, instead of silently upserting.
func TestCreateRequireImport_BlocksWhenExists(t *testing.T) {
	spec := requireImportSpec(t)
	r := &Resource{
		spec: spec,
		meta: ResourceMeta{
			Operations:    Operations{Create: "PutThing", Read: "GetThing", Update: "PutThing"},
			IDFormat:      "{org}",
			RequireImport: true,
		},
	}
	mock := &mockTransport{responses: map[string]mockResponse{
		"GET /things/acme": {status: 200, body: `{"org":"acme","value":"existing"}`},
	}}
	SetTransportResolver(func(_ context.Context) (Transport, error) { return mock, nil })

	_, err := r.Create(context.Background(), p.CreateRequest{
		Properties: propMap(map[string]any{"org": "acme", "value": "new"}),
	})
	if err == nil {
		t.Fatalf("expected error when resource already exists, got nil")
	}
	if !strings.Contains(err.Error(), "already exists") || !strings.Contains(err.Error(), "import") {
		t.Errorf("error message should mention `already exists` and `import`: %v", err)
	}
	if len(mock.calls) != 1 || mock.calls[0] != "GET /things/acme" {
		t.Errorf("expected exactly one GET, got: %v", mock.calls)
	}
}

// TestCreateRequireImport_ProceedsOn404 verifies that a 404 from the probe
// is treated as "not yet exists" and Create proceeds with the upsert call.
func TestCreateRequireImport_ProceedsOn404(t *testing.T) {
	spec := requireImportSpec(t)
	r := &Resource{
		spec: spec,
		meta: ResourceMeta{
			Operations:    Operations{Create: "PutThing", Read: "GetThing", Update: "PutThing"},
			IDFormat:      "{org}",
			RequireImport: true,
		},
	}
	var putCalled bool
	mock := &mockTransport{responseFn: func(req *http.Request) mockResponse {
		switch req.Method {
		case "PUT":
			putCalled = true
			return mockResponse{status: 200, body: `{"org":"acme","value":"new"}`}
		case "GET":
			if putCalled {
				return mockResponse{status: 200, body: `{"org":"acme","value":"new"}`}
			}
			return mockResponse{status: 404, body: `{"error":"not found"}`}
		}
		return mockResponse{status: 500, body: "unexpected"}
	}}
	SetTransportResolver(func(_ context.Context) (Transport, error) { return mock, nil })

	resp, err := r.Create(context.Background(), p.CreateRequest{
		Properties: propMap(map[string]any{"org": "acme", "value": "new"}),
	})
	if err != nil {
		t.Fatalf("create should proceed on 404, got: %v\n  calls: %v", err, mock.calls)
	}
	if resp.ID != "acme" {
		t.Errorf("ID: got %q, want %q", resp.ID, "acme")
	}
	// Probe (GET, 404) → create (PUT) → read-after-create (GET, now 200).
	wantCalls := []string{"GET /things/acme", "PUT /things/acme", "GET /things/acme"}
	if len(mock.calls) != len(wantCalls) {
		t.Fatalf("expected %d calls, got %d: %v", len(wantCalls), len(mock.calls), mock.calls)
	}
	for i, want := range wantCalls {
		if mock.calls[i] != want {
			t.Errorf("call %d: got %q, want %q", i, mock.calls[i], want)
		}
	}
}

// TestCreateRequireImport_NoReadOp_OptsOut verifies that RequireImport is a
// no-op for resources without a read op declared. The dispatch can't probe
// what it can't read.
func TestCreateRequireImport_NoReadOp_OptsOut(t *testing.T) {
	spec := requireImportSpec(t)
	r := &Resource{
		spec: spec,
		meta: ResourceMeta{
			Operations:    Operations{Create: "PutThing", Update: "PutThing"},
			IDFormat:      "{org}",
			RequireImport: true,
		},
	}
	mock := &mockTransport{responses: map[string]mockResponse{
		"PUT /things/acme": {status: 200, body: `{"org":"acme","value":"new"}`},
	}}
	SetTransportResolver(func(_ context.Context) (Transport, error) { return mock, nil })

	_, err := r.Create(context.Background(), p.CreateRequest{
		Properties: propMap(map[string]any{"org": "acme", "value": "new"}),
	})
	if err != nil {
		t.Fatalf("create should proceed without probe, got: %v\n  calls: %v", err, mock.calls)
	}
	if len(mock.calls) != 1 || mock.calls[0] != "PUT /things/acme" {
		t.Errorf("expected only the PUT call, got: %v", mock.calls)
	}
}

// requireImportSpec builds a minimal synthetic spec with a single PUT-shaped
// upsert resource at /things/{org}, used by the RequireImport tests.
func requireImportSpec(t *testing.T) *Spec {
	t.Helper()
	const specJSON = `{
	  "openapi": "3.0.0",
	  "components": {"schemas": {
	    "Body":  {"type": "object", "properties": {"value": {"type": "string"}}},
	    "Read":  {"type": "object", "properties": {"org": {"type": "string"}, "value": {"type": "string"}}}
	  }},
	  "paths": {
	    "/things/{org}": {
	      "put": {
	        "operationId": "PutThing",
	        "parameters": [{"name": "org", "in": "path", "required": true, "schema": {"type": "string"}}],
	        "requestBody": {"content": {"application/json": {"schema": {"$ref": "#/components/schemas/Body"}}}},
	        "responses": {"200": {"content": {"application/json": {"schema": {"$ref": "#/components/schemas/Read"}}}}}
	      },
	      "get": {
	        "operationId": "GetThing",
	        "parameters": [{"name": "org", "in": "path", "required": true, "schema": {"type": "string"}}],
	        "responses": {"200": {"content": {"application/json": {"schema": {"$ref": "#/components/schemas/Read"}}}}}
	      }
	    }
	  }
	}`
	spec, err := ParseSpec([]byte(specJSON))
	if err != nil {
		t.Fatalf("parse synthetic spec: %v", err)
	}
	return spec
}

// TestCreateMissingPathParam verifies that synthesizeID returns a clear
// error when a required path-param value can't be found in state. This
// catches bugs where a rename mismatch or input omission would otherwise
// produce a malformed ID.
func TestCreateMissingPathParam(t *testing.T) {
	spec, meta := loadFixtures(t)
	resources := Resources(spec, meta)
	rm := meta.Resources["pulumiservice:v2:Team"]
	tok := rm.Token
	if tok == "" {
		tok = "pulumiservice:v2:Team"
	}
	r := resources[tok]
	if r == nil {
		t.Fatalf("Team not in factory output")
	}

	mock := &mockTransport{responses: map[string]mockResponse{
		"POST /api/orgs/test-org/teams/pulumi": {status: 200, body: `{}`},
	}}
	SetTransportResolver(func(_ context.Context) (Transport, error) {
		return mock, nil
	})

	// Omit `name` — the body field that maps via rename to teamName for the path.
	req := p.CreateRequest{Properties: propMap(map[string]any{
		"orgName": "test-org",
	})}
	_, err := r.Create(context.Background(), req)
	if err == nil {
		t.Fatalf("expected error on missing path param, got nil")
	}
	// Surface the error so we can see what shape it has.
	if !strings.Contains(err.Error(), "missing") && !strings.Contains(err.Error(), "teamName") {
		t.Logf("error message: %v", err)
	}
}

