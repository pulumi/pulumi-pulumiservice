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

// Repeated test-only literals — extracted so goconst stops flagging them.
const (
	teamToken      = "pulumiservice:v2:Team" //nolint:gosec // resource token, not a credential
	getThingPath   = "GET /things/acme/thing-1"
	patchThingPath = "PATCH /things/acme/thing-1"
)

// mockTransport serves canned responses keyed by "<METHOD> <path>". Set
// responseFn for stateful behavior; it overrides responses if non-nil.
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

// loadFixtures reads spec.json and metadata.json from ../cloud.
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

// TestCreateSynthesizesID exercises Create against the real spec, covering
// server-generated id, 204 no-content, composite-from-path, and singleton
// shapes.
func TestCreateSynthesizesID(t *testing.T) {
	spec, meta := loadFixtures(t)
	resources := Resources(spec, meta)

	cases := []struct {
		token      string
		responses  map[string]mockResponse
		responseFn func() func(req *http.Request) mockResponse // stateful, overrides responses
		inputs     map[string]any
		wantID     string
	}{
		{
			// AgentPool: server-generated id, ID = {orgName}/{uuid}.
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
			// StackTag: 204 No Content, identity entirely from inputs.
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
			// Team: composite identity from path.
			token: teamToken,
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
			// DefaultOrganization: requireImport singleton; GET returns 404
			// for the probe, 200 after the mutating call writes it.
			token: "pulumiservice:v2:DefaultOrganization",
			responseFn: func() func(req *http.Request) mockResponse {
				written := false
				return func(req *http.Request) mockResponse {
					switch {
					case req.Method == http.MethodGet && !written:
						return mockResponse{status: 404, body: `{"error":"not found"}`}
					case req.Method == http.MethodGet:
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
			// AuditLogExportConfiguration: requireImport singleton, same
			// staging pattern as DefaultOrganization.
			token: "pulumiservice:v2:AuditLogExportConfiguration",
			responseFn: func() func(req *http.Request) mockResponse {
				written := false
				return func(req *http.Request) mockResponse {
					switch {
					case req.Method == http.MethodGet && !written:
						return mockResponse{status: 404, body: `{"error":"not found"}`}
					case req.Method == http.MethodGet:
						return mockResponse{status: 200, body: `{}`}
					case req.Method == http.MethodPost:
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
			// Role: server-generated id, response rename id→roleID.
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

// TestCreateFusesYamlUpdateAfterJsonCreate: when create is JSON and update
// is yaml, supplying a yaml input fires create then a follow-up yaml update.
func TestCreateFusesYamlUpdateAfterJsonCreate(t *testing.T) {
	const specJSON = `{
	  "openapi": "3.0.0",
	  "components": {"schemas": {
	    "CreateBody": {"type": "object", "properties": {"project": {"type": "string"}, "name": {"type": "string"}}}
	  }},
	  "paths": {
	    "/envs/{org}": {
	      "post": {
	        "operationId": "CreateEnv",
	        "parameters": [{"name": "org", "in": "path", "required": true, "schema": {"type": "string"}}],
	        "requestBody": {"content": {"application/json": {"schema": {"$ref": "#/components/schemas/CreateBody"}}}},
	        "responses": {"200": {"description": "OK"}}
	      }
	    },
	    "/envs/{org}/{project}/{name}": {
	      "patch": {
	        "operationId": "UpdateEnv",
	        "parameters": [
	          {"name": "org",     "in": "path", "required": true, "schema": {"type": "string"}},
	          {"name": "project", "in": "path", "required": true, "schema": {"type": "string"}},
	          {"name": "name",    "in": "path", "required": true, "schema": {"type": "string"}}
	        ],
	        "requestBody": {"content": {"application/x-yaml": {"schema": {"type": "string"}}}},
	        "responses": {"204": {"description": "no content"}}
	      }
	    }
	  }
	}`
	spec, _ := ParseSpec([]byte(specJSON))
	r := &Resource{
		spec: spec,
		meta: ResourceMeta{
			Operations: Operations{Create: "CreateEnv", Update: "UpdateEnv"},
			IDFormat:   "{org}/{project}/{name}",
		},
	}

	type captured struct {
		method      string
		path        string
		contentType string
		body        string
	}
	var seen []captured
	mock := &mockTransport{
		responseFn: func(req *http.Request) mockResponse {
			b, _ := io.ReadAll(req.Body)
			seen = append(seen, captured{
				method:      req.Method,
				path:        req.URL.Path,
				contentType: req.Header.Get("Content-Type"),
				body:        string(b),
			})
			return mockResponse{status: 200, body: ""}
		},
	}
	SetTransportResolver(func(_ context.Context) (Transport, error) { return mock, nil })

	yamlBody := "values:\n  bootstrap:\n    appVersion: 1.0.0\n"
	// Secret-wrapped yaml mirrors what codegen sends for Secret fields.
	yamlVal := property.New(yamlBody).WithSecret(true)
	inputs := property.NewMap(map[string]property.Value{
		"org":     property.New("acme"),
		"project": property.New("default"),
		"name":    property.New("platform-bootstrap"),
		"yaml":    yamlVal,
	})
	_, err := r.Create(context.Background(), p.CreateRequest{Properties: inputs})
	if err != nil {
		t.Fatalf("create: %v\n  calls: %v", err, mock.calls)
	}

	if len(seen) != 2 {
		t.Fatalf("expected 2 HTTP calls (create + yaml apply), got %d: %#v", len(seen), seen)
	}
	if seen[0].method != http.MethodPost || seen[0].path != "/envs/acme" {
		t.Errorf("first call: got %s %s, want POST /envs/acme", seen[0].method, seen[0].path)
	}
	if seen[0].contentType != contentJSON {
		t.Errorf("first call content-type: got %q, want application/json", seen[0].contentType)
	}
	if seen[1].method != http.MethodPatch || seen[1].path != "/envs/acme/default/platform-bootstrap" {
		t.Errorf("second call: got %s %s, want PATCH /envs/acme/default/platform-bootstrap", seen[1].method, seen[1].path)
	}
	if seen[1].contentType != "application/x-yaml" {
		t.Errorf("second call content-type: got %q, want application/x-yaml", seen[1].contentType)
	}
	if seen[1].body != yamlBody {
		t.Errorf("second call body:\ngot:  %q\nwant: %q", seen[1].body, yamlBody)
	}
}

// TestReadDecodesYamlResponseBody: an application/x-yaml response binds
// to state["yaml"] instead of being JSON-unmarshaled.
func TestReadDecodesYamlResponseBody(t *testing.T) {
	const specJSON = `{
	  "openapi": "3.0.0",
	  "paths": {
	    "/envs/{org}/{project}/{name}": {
	      "get": {
	        "operationId": "ReadEnv",
	        "parameters": [
	          {"name": "org",     "in": "path", "required": true, "schema": {"type": "string"}},
	          {"name": "project", "in": "path", "required": true, "schema": {"type": "string"}},
	          {"name": "name",    "in": "path", "required": true, "schema": {"type": "string"}}
	        ],
	        "responses": {"200": {"content": {"application/x-yaml": {"schema": {"type": "string"}}}}}
	      }
	    }
	  }
	}`
	spec, _ := ParseSpec([]byte(specJSON))
	r := &Resource{
		spec: spec,
		meta: ResourceMeta{
			Operations: Operations{Read: "ReadEnv"},
			IDFormat:   "{org}/{project}/{name}",
		},
	}

	yamlBody := "values:\n  bootstrap:\n    appVersion: 1.0.0\n"
	mock := &mockTransport{
		responseFn: func(_ *http.Request) mockResponse {
			return mockResponse{status: 200, body: yamlBody}
		},
	}
	SetTransportResolver(func(_ context.Context) (Transport, error) { return mock, nil })

	resp, err := r.Read(context.Background(), p.ReadRequest{
		ID:         "acme/default/platform-bootstrap",
		Properties: property.NewMap(map[string]property.Value{}),
	})
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	v, ok := resp.Properties.GetOk("yaml")
	if !ok {
		t.Fatalf("read response missing 'yaml': %#v", resp.Properties)
	}
	if v.AsString() != yamlBody {
		t.Errorf("yaml mismatch:\ngot:  %q\nwant: %q", v.AsString(), yamlBody)
	}
}

// TestCreateReadAfterCreateSourcesFromInputs: the read URL is built from
// user inputs, not from the (potentially sparse) create response.
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
	        "responses": {"200": {"content": {"application/json": {
	          "schema": {"type": "object", "properties": {"id": {"type": "string"}}}
	        }}}}
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
		"POST /things/acme": {status: 200, body: `{"id":"thing-1"}`},
		getThingPath:        {status: 200, body: `{"id":"thing-1","name":"foo","status":"ready"}`},
	}}
	SetTransportResolver(func(_ context.Context) (Transport, error) { return mock, nil })

	resp, err := r.Create(context.Background(), p.CreateRequest{
		Properties: propMap(map[string]any{"org": "acme", "name": "foo"}),
	})
	if err != nil {
		t.Fatalf("create: %v\n  calls: %v", err, mock.calls)
	}
	if len(mock.calls) != 2 || mock.calls[1] != getThingPath {
		t.Errorf("expected POST then GET /things/acme/thing-1, got: %v", mock.calls)
	}
	if v, ok := resp.Properties.GetOk("status"); !ok || v.AsString() != "ready" {
		t.Errorf("state missing read-only field 'status': ok=%v v=%q", ok, v.AsString())
	}
}

// TestCreateReadsAfterCreate: Create fires read after the create call and
// returns state from the read response, not the (sparse) create echo.
func TestCreateReadsAfterCreate(t *testing.T) {
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
	        "responses": {"200": {"content": {"application/json": {
	          "schema": {"type": "object", "properties": {"id": {"type": "string"}}}
	        }}}}
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
	readBody := `{"id":"thing-1","name":"foo","lastUpdate":"2026-05-05T00:00:00Z","status":"ready"}`
	mock := &mockTransport{responses: map[string]mockResponse{
		"POST /things/acme": {status: 200, body: `{"id":"thing-1"}`},
		getThingPath:        {status: 200, body: readBody},
	}}
	SetTransportResolver(func(_ context.Context) (Transport, error) { return mock, nil })

	resp, err := r.Create(context.Background(), p.CreateRequest{
		Properties: propMap(map[string]any{"org": "acme", "name": "foo"}),
	})
	if err != nil {
		t.Fatalf("create: %v\n  calls: %v", err, mock.calls)
	}

	if len(mock.calls) != 2 || mock.calls[0] != "POST /things/acme" || mock.calls[1] != getThingPath {
		t.Errorf("expected POST then GET, got: %v", mock.calls)
	}
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
	// Path params are program-owned: they live in inputs and the ID, not state.
	if _, ok := resp.Properties.GetOk("org"); ok {
		t.Errorf("state should not carry path-param `org`")
	}
	if resp.ID != "acme/thing-1" {
		t.Errorf("ID: got %q, want %q", resp.ID, "acme/thing-1")
	}
}

// TestCreateRequireImport_BlocksWhenExists: RequireImport aborts Create
// when the read probe returns 200.
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

// TestCreateRequireImport_ProceedsOn404: a 404 from the probe means proceed.
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
		case http.MethodPut:
			putCalled = true
			return mockResponse{status: 200, body: `{"org":"acme","value":"new"}`}
		case http.MethodGet:
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

// TestCreateRequireImport_NoReadOp_OptsOut: RequireImport is a no-op
// without a read op (the dispatch can't probe what it can't read).
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

// requireImportSpec is a minimal PUT-shaped upsert spec for RequireImport tests.
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

// TestCreateMissingPathParam: synthesizeID returns a clear error when a
// required path param is missing.
func TestCreateMissingPathParam(t *testing.T) {
	spec, meta := loadFixtures(t)
	resources := Resources(spec, meta)
	rm := meta.Resources[teamToken]
	tok := rm.Token
	if tok == "" {
		tok = teamToken
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

	// Omit `name` — the body field that renames to teamName for the path.
	req := p.CreateRequest{Properties: propMap(map[string]any{
		"orgName": "test-org",
	})}
	_, err := r.Create(context.Background(), req)
	if err == nil {
		t.Fatalf("expected error on missing path param, got nil")
	}
	if !strings.Contains(err.Error(), "missing") && !strings.Contains(err.Error(), "teamName") {
		t.Logf("error message: %v", err)
	}
}

// TestUpdateReadsAfterUpdate: PATCH endpoints that echo back only the
// mutated fields would otherwise shrink state to that sparse response and
// leak as `+id` drift on the next refresh. The read-after-update merge
// re-pulls the canonical record so state stays whole.
func TestUpdateReadsAfterUpdate(t *testing.T) {
	const specJSON = `{
	  "openapi": "3.0.0",
	  "components": {"schemas": {
	    "Patch":   {"type": "object", "properties": {"name": {"type": "string"}, "description": {"type": "string"}}},
	    "Sparse":  {"type": "object", "properties": {"name": {"type": "string"}, "description": {"type": "string"}}},
	    "Full":    {"type": "object", "properties": {
	      "id":          {"type": "string"},
	      "name":        {"type": "string"},
	      "description": {"type": "string"},
	      "created":     {"type": "string"}
	    }}
	  }},
	  "paths": {
	    "/things/{org}/{id}": {
	      "get": {
	        "operationId": "GetThing",
	        "parameters": [
	          {"name": "org", "in": "path", "required": true, "schema": {"type": "string"}},
	          {"name": "id",  "in": "path", "required": true, "schema": {"type": "string"}}
	        ],
	        "responses": {"200": {"content": {"application/json": {"schema": {"$ref": "#/components/schemas/Full"}}}}}
	      },
	      "patch": {
	        "operationId": "PatchThing",
	        "parameters": [
	          {"name": "org", "in": "path", "required": true, "schema": {"type": "string"}},
	          {"name": "id",  "in": "path", "required": true, "schema": {"type": "string"}}
	        ],
	        "requestBody": {"content": {"application/json": {"schema": {"$ref": "#/components/schemas/Patch"}}}},
	        "responses": {"200": {"content": {"application/json": {"schema": {"$ref": "#/components/schemas/Sparse"}}}}}
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
			Operations: Operations{Read: "GetThing", Update: "PatchThing"},
			IDFormat:   "{org}/{id}",
		},
	}
	patchBody := `{"name":"foo-renamed","description":"rotated"}`
	readBody := `{"id":"thing-1","name":"foo-renamed","description":"rotated","created":"2026-05-05T00:00:00Z"}`
	priorState := map[string]any{
		"id": "thing-1", "name": "foo", "description": "original",
		"created": "2026-05-05T00:00:00Z",
	}
	mock := &mockTransport{responses: map[string]mockResponse{
		patchThingPath: {status: 200, body: patchBody},
		getThingPath:   {status: 200, body: readBody},
	}}
	SetTransportResolver(func(_ context.Context) (Transport, error) { return mock, nil })

	resp, err := r.Update(context.Background(), p.UpdateRequest{
		Inputs:    propMap(map[string]any{"org": "acme", "name": "foo-renamed", "description": "rotated"}),
		OldInputs: propMap(map[string]any{"org": "acme", "name": "foo", "description": "original"}),
		State:     propMap(priorState),
	})
	if err != nil {
		t.Fatalf("update: %v\n  calls: %v", err, mock.calls)
	}
	if len(mock.calls) != 2 || mock.calls[0] != patchThingPath || mock.calls[1] != getThingPath {
		t.Errorf("expected PATCH then GET, got: %v", mock.calls)
	}
	for _, key := range []string{"id", "created", "name", "description"} {
		if _, ok := resp.Properties.GetOk(key); !ok {
			t.Errorf("state missing %q after read-after-update", key)
		}
	}
}

// TestUpdateMergesPriorStateWithoutReadOp: without a read op, the update
// path still has to preserve fields the update response doesn't echo —
// otherwise refresh sees a synthetic diff.
func TestUpdateMergesPriorStateWithoutReadOp(t *testing.T) {
	const specJSON = `{
	  "openapi": "3.0.0",
	  "components": {"schemas": {
	    "Patch":  {"type": "object", "properties": {"description": {"type": "string"}}},
	    "Sparse": {"type": "object", "properties": {"name": {"type": "string"}, "description": {"type": "string"}}}
	  }},
	  "paths": {
	    "/things/{org}/{id}": {
	      "patch": {
	        "operationId": "PatchThing",
	        "parameters": [
	          {"name": "org", "in": "path", "required": true, "schema": {"type": "string"}},
	          {"name": "id",  "in": "path", "required": true, "schema": {"type": "string"}}
	        ],
	        "requestBody": {"content": {"application/json": {"schema": {"$ref": "#/components/schemas/Patch"}}}},
	        "responses": {"200": {"content": {"application/json": {"schema": {"$ref": "#/components/schemas/Sparse"}}}}}
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
			Operations: Operations{Update: "PatchThing"},
			IDFormat:   "{org}/{id}",
		},
	}
	mock := &mockTransport{responses: map[string]mockResponse{
		patchThingPath: {status: 200, body: `{"name":"foo","description":"rotated"}`},
	}}
	SetTransportResolver(func(_ context.Context) (Transport, error) { return mock, nil })

	priorState := map[string]any{
		"id": "thing-1", "name": "foo", "description": "original",
		"created": "2026-05-05T00:00:00Z",
	}
	resp, err := r.Update(context.Background(), p.UpdateRequest{
		Inputs:    propMap(map[string]any{"org": "acme", "description": "rotated"}),
		OldInputs: propMap(map[string]any{"org": "acme", "description": "original"}),
		State:     propMap(priorState),
	})
	if err != nil {
		t.Fatalf("update: %v\n  calls: %v", err, mock.calls)
	}
	if v, ok := resp.Properties.GetOk("id"); !ok || v.AsString() != "thing-1" {
		t.Errorf("state missing or wrong `id` after update-with-no-read-op: %v (ok=%v)", v, ok)
	}
	if v, ok := resp.Properties.GetOk("created"); !ok || v.AsString() != "2026-05-05T00:00:00Z" {
		t.Errorf("state missing prior `created` after update-with-no-read-op: %v (ok=%v)", v, ok)
	}
	if v, ok := resp.Properties.GetOk("description"); !ok || v.AsString() != "rotated" {
		t.Errorf("state has stale description, want %q got %v (ok=%v)", "rotated", v, ok)
	}
}

// TestBuildRequestBody: the body is driven by the request schema, not by
// the inputs map — fields the schema doesn't declare get filtered out,
// fields the schema does declare get pulled from inputs via the reverse
// rename. This means a path-param wire name that also shows up in the
// body (e.g. OrganizationWebhook's organizationName) survives in the
// body, since the API validates the two copies match.
func TestBuildRequestBody(t *testing.T) {
	const specJSON = `{
	  "openapi": "3.0.0",
	  "components": {"schemas": {
	    "TeamBody":  {"type": "object", "properties": {
	      "name":        {"type": "string"},
	      "description": {"type": "string"}
	    }},
	    "WidgetBody": {"type": "object", "properties": {
	      "widgetID":    {"type": "string"},
	      "description": {"type": "string"}
	    }},
	    "WebhookBody": {"type": "object", "properties": {
	      "organizationName": {"type": "string"},
	      "name":             {"type": "string"},
	      "payloadUrl":       {"type": "string"}
	    }}
	  }},
	  "paths": {
	    "/teams/{org}/pulumi": {
	      "post": {
	        "operationId": "CreateTeam",
	        "parameters": [{"name": "org", "in": "path", "required": true, "schema": {"type": "string"}}],
	        "requestBody": {"content": {"application/json": {"schema": {"$ref": "#/components/schemas/TeamBody"}}}},
	        "responses": {"200": {"description": "OK"}}
	      }
	    },
	    "/widgets/{org}": {
	      "post": {
	        "operationId": "CreateWidget",
	        "parameters": [{"name": "org", "in": "path", "required": true, "schema": {"type": "string"}}],
	        "requestBody": {"content": {"application/json": {"schema": {"$ref": "#/components/schemas/WidgetBody"}}}},
	        "responses": {"200": {"description": "OK"}}
	      }
	    },
	    "/orgs/{orgName}/hooks": {
	      "post": {
	        "operationId": "CreateOrgHook",
	        "parameters": [{"name": "orgName", "in": "path", "required": true, "schema": {"type": "string"}}],
	        "requestBody": {"content": {"application/json": {"schema": {"$ref": "#/components/schemas/WebhookBody"}}}},
	        "responses": {"200": {"description": "OK"}}
	      }
	    }
	  }
	}`
	spec, err := ParseSpec([]byte(specJSON))
	if err != nil {
		t.Fatalf("parse synthetic spec: %v", err)
	}

	cases := []struct {
		name    string
		op      string
		renames map[string]string
		inputs  map[string]any
		want    map[string]any
	}{
		{
			name:    "path param not declared in body schema is dropped",
			op:      "CreateTeam",
			renames: nil,
			inputs:  map[string]any{"org": "acme", "name": "infra", "description": "infra team"},
			want:    map[string]any{"name": "infra", "description": "infra team"},
		},
		{
			name:    "rename for path-param does not rewrite body field of same Pulumi name",
			op:      "CreateTeam",
			renames: map[string]string{"name": "teamName"}, // path-only rename
			inputs:  map[string]any{"org": "acme", "name": "infra", "description": "infra team"},
			want:    map[string]any{"name": "infra", "description": "infra team"},
		},
		{
			name:    "rename that targets a body field is applied",
			op:      "CreateWidget",
			renames: map[string]string{"id": "widgetID"}, // body field rename
			inputs:  map[string]any{"org": "acme", "id": "w-1", "description": "wodget"},
			want:    map[string]any{"widgetID": "w-1", "description": "wodget"},
		},
		{
			name: "path param ALSO declared in body schema is kept (server validates match)",
			op:   "CreateOrgHook",
			// OrganizationWebhook-style: Pulumi-side `organizationName` maps to
			// wire path-param `orgName`, but the body schema also declares a
			// wire-side `organizationName`. The body must carry it.
			renames: map[string]string{"organizationName": "orgName"},
			inputs: map[string]any{
				"organizationName": "acme", "name": "alerts", "payloadUrl": "https://x",
			},
			want: map[string]any{
				"organizationName": "acme", "name": "alerts", "payloadUrl": "https://x",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			op, ok := spec.Op(tc.op)
			if !ok {
				t.Fatalf("op %q not in spec", tc.op)
			}
			r := &Resource{spec: spec, meta: ResourceMeta{Renames: tc.renames}}
			got := r.buildRequestBody(op, propMap(tc.inputs))
			if len(got) != len(tc.want) {
				t.Errorf("len mismatch: got %v, want %v", got, tc.want)
			}
			for k, want := range tc.want {
				if got[k] != want {
					t.Errorf("body[%q]: got %v, want %v\n  full got: %v", k, got[k], want, got)
				}
			}
		})
	}
}

// TestDeleteIsIdempotentOn404 pins the runtime behavior: 404 on Delete is
// treated as success. Centralizing this means every v2 resource is
// uniformly idempotent without per-resource metadata or scaffolder hints.
func TestDeleteIsIdempotentOn404(t *testing.T) {
	const specJSON = `{
	  "openapi": "3.0.0",
	  "paths": {
	    "/things/{org}/{id}": {
	      "delete": {
	        "operationId": "DeleteThing",
	        "parameters": [
	          {"name": "org", "in": "path", "required": true, "schema": {"type": "string"}},
	          {"name": "id",  "in": "path", "required": true, "schema": {"type": "string"}}
	        ],
	        "responses": {"204": {"description": "no content"}}
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
			Operations: Operations{Delete: "DeleteThing"},
			IDFormat:   "{org}/{id}",
		},
	}

	t.Run("204 is success", func(t *testing.T) {
		mock := &mockTransport{responses: map[string]mockResponse{
			"DELETE /things/acme/gone": {status: 204},
		}}
		SetTransportResolver(func(_ context.Context) (Transport, error) { return mock, nil })
		err := r.Delete(context.Background(), p.DeleteRequest{
			ID:         "acme/gone",
			Properties: propMap(map[string]any{"org": "acme", "id": "gone"}),
		})
		if err != nil {
			t.Errorf("204 should be nil, got: %v", err)
		}
	})

	t.Run("404 is also success — already gone is what Delete wants", func(t *testing.T) {
		mock := &mockTransport{responses: map[string]mockResponse{
			"DELETE /things/acme/gone": {status: 404, body: `{"code":404,"message":"not found"}`},
		}}
		SetTransportResolver(func(_ context.Context) (Transport, error) { return mock, nil })
		err := r.Delete(context.Background(), p.DeleteRequest{
			ID:         "acme/gone",
			Properties: propMap(map[string]any{"org": "acme", "id": "gone"}),
		})
		if err != nil {
			t.Errorf("404 should be nil, got: %v", err)
		}
	})

	t.Run("non-404 errors still surface", func(t *testing.T) {
		mock := &mockTransport{responses: map[string]mockResponse{
			"DELETE /things/acme/gone": {status: 500, body: `oops`},
		}}
		SetTransportResolver(func(_ context.Context) (Transport, error) { return mock, nil })
		err := r.Delete(context.Background(), p.DeleteRequest{
			ID:         "acme/gone",
			Properties: propMap(map[string]any{"org": "acme", "id": "gone"}),
		})
		if err == nil {
			t.Error("500 should propagate")
		}
	})
}

// TestPropertyValueToStringNumberFormatting pins decimal-not-scientific
// formatting for numeric values. Path-param substitution and synthesized
// IDs go through this — `1e+18` makes a useless URL segment.
func TestPropertyValueToStringNumberFormatting(t *testing.T) {
	cases := []struct {
		name string
		in   float64
		want string
	}{
		{"small integer", 42, "42"},
		{"large integer (scientific risk)", 1e18, "1000000000000000000"},
		{"larger integer (scientific risk)", 1.234e15, "1234000000000000"},
		{"fractional", 3.14, "3.14"},
		{"zero", 0, "0"},
		{"negative", -1234567, "-1234567"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := propertyValueToString(property.New(tc.in))
			if got != tc.want {
				t.Errorf("propertyValueToString(%v): got %q, want %q", tc.in, got, tc.want)
			}
			if strings.Contains(got, "e+") || strings.Contains(got, "e-") {
				t.Errorf("propertyValueToString(%v) returned scientific notation %q", tc.in, got)
			}
		})
	}
}
