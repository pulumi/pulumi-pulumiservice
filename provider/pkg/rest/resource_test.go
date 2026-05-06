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
// matching is exact against the request URL's path component.
type mockTransport struct {
	responses map[string]mockResponse
	calls     []string // method + " " + path, in order
}

type mockResponse struct {
	status int
	body   string
}

func (m *mockTransport) Do(_ context.Context, req *http.Request) (*http.Response, error) {
	key := req.Method + " " + req.URL.Path
	m.calls = append(m.calls, key)
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
		// Mock HTTP responses keyed by "<METHOD> <path>".
		responses map[string]mockResponse
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
			// DefaultOrganization: singleton at /api/user/organizations/default;
			// create path is POST /api/user/organizations/{orgName}/default.
			token: "pulumiservice:v2:DefaultOrganization",
			responses: map[string]mockResponse{
				"POST /api/user/organizations/test-org/default": {
					status: 200,
					body:   `{}`,
				},
				"GET /api/user/organizations/default": {
					status: 200,
					body:   `{"orgName":"test-org"}`,
				},
			},
			inputs: map[string]any{
				"orgName": "test-org",
			},
			wantID: "test-org",
		},
		{
			// AuditLogExportConfiguration: singleton, all ops on the same path.
			token: "pulumiservice:v2:AuditLogExportConfiguration",
			responses: map[string]mockResponse{
				"POST /api/orgs/test-org/auditlogs/export/config": {
					status: 200,
					body:   `{}`,
				},
				"GET /api/orgs/test-org/auditlogs/export/config": {
					status: 200,
					body:   `{}`,
				},
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
	// Path params should still be present (populatePathParams runs after fetchState).
	if v, ok := resp.Properties.GetOk("org"); !ok || v.AsString() != "acme" {
		t.Errorf("state lost path-param `org` after read-after-create: ok=%v v=%q", ok, v.AsString())
	}
	// ID is unchanged by read.
	if resp.ID != "acme/thing-1" {
		t.Errorf("ID: got %q, want %q", resp.ID, "acme/thing-1")
	}
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

