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
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi/sdk/v3/go/property"
)

// Repeated test-only literals — extracted so goconst stops flagging them.
const (
	teamToken            = "pulumiservice:api:Team" //nolint:gosec // resource token, not a credential
	getThingPath         = "GET /things/acme/thing-1"
	patchThingPath       = "PATCH /things/acme/thing-1"
	nameKey              = "name"
	orgKey               = "org"
	orgNameKey           = "orgName"
	valueKey             = "value"
	organizationNameVal  = "organizationName"
	widgetIDVal          = "widgetID"
	acmeVal              = "acme"
	infraVal             = "infra"
	infraTeamVal         = "infra team"
	newVal               = "new"
	originalVal          = "original"
	rotatedVal           = "rotated"
	goneVal              = "gone"
	acmeGoneID           = "acme/gone"
	createThingOp        = "CreateThing"
	putThingOp           = "PutThing"
	orgIDFormat          = "{org}/{id}"
	orgFormat            = "{org}"
	postThingsAcme       = "POST /things/acme"
	getThingsAcme        = "GET /things/acme"
	putThingsAcme        = "PUT /things/acme"
	deleteThingsAcmeGone = "DELETE /things/acme/gone"
	unexpectedBody       = "unexpected"
	notFoundBody         = `{"error":"not found"}`
	orgAcmeNewBody       = `{"org":"acme","value":"new"}`
	createdTimestamp     = "2026-05-05T00:00:00Z"
	testOrgName          = "test-org"
	orgProjectNameFormat = "{org}/{project}/{name}"
	getThingOp           = "GetThing"
	fooVal               = "foo"
	thing1ID             = "thing-1"
	createdKey           = "created"
	statusKey            = "status"
	passwordKey          = "password" //nolint:gosec // property name, not a credential
	detailsKey           = "details"
	tagsKey              = "tags"
	newDescVal           = "new desc"
	newTagKey            = "newTag"
	envNameKey           = "envName"
	projectNameKey       = "projectName"
	myprojVal            = "myproj"
	myenvVal             = "myenv"
	ownerVal             = "owner"
	teamXVal             = "team-x"
	teamYVal             = "team-y"
	typeKey              = "type"
	cloudCountKey        = "cloudCount"
	stackVal             = "stack"
	testOrgInfraID       = "test-org/infra"
	itemsKey             = "items"
	envVarsKey           = "environmentVariables"
	newDescriptionKey    = "newDescription"
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
		{ //nolint:gosec // G101: test fixture, not a real credential.
			// AgentPool: server-generated id, ID = {orgName}/{uuid}.
			token: "pulumiservice:api:AgentPool",
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
				orgNameKey:     testOrgName,
				nameKey:        "runners",
				descriptionKey: "hi",
			},
			wantID: "test-org/abc-123",
		},
		{ //nolint:gosec // G101: test fixture, not a real credential.
			// StackTag: 204 No Content, identity entirely from inputs.
			token: "pulumiservice:api:StackTag",
			responses: map[string]mockResponse{
				"POST /api/stacks/test-org/myproj/mystack/tags": {status: 204, body: ""},
			},
			inputs: map[string]any{
				orgNameKey:     testOrgName,
				projectNameKey: myprojVal,
				"stackName":    "mystack",
				nameKey:        ownerVal, // body field; rename name→tagName for path
				valueKey:       teamXVal,
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
				orgNameKey:     testOrgName,
				nameKey:        infraVal, // body; renames map name→teamName for path
				descriptionKey: infraTeamVal,
			},
			wantID: testOrgInfraID,
		},
		{ //nolint:gosec // G101: test fixture, not a real credential.
			// DefaultOrganization: requireImport singleton; GET returns 404
			// for the probe, 200 after the mutating call writes it.
			token: "pulumiservice:api:DefaultOrganization",
			responseFn: func() func(req *http.Request) mockResponse {
				written := false
				return func(req *http.Request) mockResponse {
					switch {
					case req.Method == http.MethodGet && !written:
						return mockResponse{status: 404, body: notFoundBody}
					case req.Method == http.MethodGet:
						return mockResponse{status: 200, body: `{"orgName":"test-org"}`}
					case req.URL.Path == "/api/user/organizations/test-org/default":
						written = true
						return mockResponse{status: 200, body: `{}`}
					}
					return mockResponse{status: 500, body: unexpectedBody}
				}
			},
			inputs: map[string]any{
				orgNameKey: testOrgName,
			},
			wantID: testOrgName,
		},
		{ //nolint:gosec // G101: test fixture, not a real credential.
			// AuditLogExportConfiguration: requireImport singleton, same
			// staging pattern as DefaultOrganization.
			token: "pulumiservice:api:AuditLogExportConfiguration",
			responseFn: func() func(req *http.Request) mockResponse {
				written := false
				return func(req *http.Request) mockResponse {
					switch {
					case req.Method == http.MethodGet && !written:
						return mockResponse{status: 404, body: notFoundBody}
					case req.Method == http.MethodGet:
						return mockResponse{status: 200, body: `{}`}
					case req.Method == http.MethodPost:
						written = true
						return mockResponse{status: 200, body: `{}`}
					}
					return mockResponse{status: 500, body: unexpectedBody}
				}
			},
			inputs: map[string]any{
				orgNameKey: testOrgName,
			},
			wantID: testOrgName,
		},
		{ //nolint:gosec // G101: test fixture, not a real credential.
			// Role: server-generated id, response rename id→roleID.
			token: "pulumiservice:api:Role",
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
				orgNameKey:     testOrgName,
				nameKey:        "deployer",
				descriptionKey: "deploy access",
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
			resp, err := r.Create(t.Context(), req)
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
			IDFormat:   orgProjectNameFormat,
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
		orgKey:    property.New(acmeVal),
		"project": property.New("default"),
		nameKey:   property.New("platform-bootstrap"),
		"yaml":    yamlVal,
	})
	_, err := r.Create(t.Context(), p.CreateRequest{Properties: inputs})
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
			IDFormat:   orgProjectNameFormat,
		},
	}

	yamlBody := "values:\n  bootstrap:\n    appVersion: 1.0.0\n"
	mock := &mockTransport{
		responseFn: func(_ *http.Request) mockResponse {
			return mockResponse{status: 200, body: yamlBody}
		},
	}
	SetTransportResolver(func(_ context.Context) (Transport, error) { return mock, nil })

	resp, err := r.Read(t.Context(), p.ReadRequest{
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

// TestReadRefreshProjectsRemoteDriftIntoInputs: on refresh (non-empty
// req.Inputs), values the server reports differently are projected into
// the returned Inputs so the engine's input-driven refresh diff surfaces
// drift (#938). Keys absent from the response (write-only fields) keep
// their prior input values; response-only keys are never added.
func TestReadRefreshProjectsRemoteDriftIntoInputs(t *testing.T) {
	const specJSON = `{
	  "openapi": "3.0.0",
	  "components": {"schemas": {
	    "Body": {"type": "object", "properties": {
	      "name":     {"type": "string"},
	      "value":    {"type": "string"},
	      "password": {"type": "string"},
	      "tags":     {"type": "array", "items": {"type": "string"}}
	    }},
	    "Thing": {"type": "object", "properties": {
	      "id":     {"type": "string"},
	      "name":   {"type": "string"},
	      "value":  {"type": "string"},
	      "tags":   {"type": "array", "items": {"type": "string"}},
	      "status": {"type": "string"}
	    }}
	  }},
	  "paths": {
	    "/things/{org}": {
	      "post": {
	        "operationId": "CreateThing",
	        "parameters": [{"name": "org", "in": "path", "required": true, "schema": {"type": "string"}}],
	        "requestBody": {"content": {"application/json": {"schema": {"$ref": "#/components/schemas/Body"}}}},
	        "responses": {"200": {"content": {"application/json": {"schema": {"$ref": "#/components/schemas/Thing"}}}}}
	      }
	    },
	    "/things/{org}/{id}": {
	      "get": {
	        "operationId": "GetThing",
	        "parameters": [
	          {"name": "org", "in": "path", "required": true, "schema": {"type": "string"}},
	          {"name": "id",  "in": "path", "required": true, "schema": {"type": "string"}}
	        ],
	        "responses": {"200": {"content": {"application/json": {"schema": {"$ref": "#/components/schemas/Thing"}}}}}
	      }
	    }
	  }
	}`
	spec, err := ParseSpec([]byte(specJSON))
	if err != nil {
		t.Fatalf("parse spec: %v", err)
	}
	r := &Resource{
		spec: spec,
		meta: ResourceMeta{
			Operations: Operations{Create: createThingOp, Read: getThingOp},
			IDFormat:   orgIDFormat,
			Fields:     map[string]FieldMeta{tagsKey: {Unordered: true}},
		},
	}

	mock := &mockTransport{responses: map[string]mockResponse{
		getThingPath: {
			status: 200,
			body:   `{"id":"thing-1","name":"n1","value":"remote","tags":["z","a"],"status":"ok"}`,
		},
	}}
	SetTransportResolver(func(_ context.Context) (Transport, error) { return mock, nil })

	oldInputs := property.NewMap(map[string]property.Value{
		orgKey:      property.New(acmeVal),
		nameKey:     property.New("n1"),
		valueKey:    property.New("local").WithSecret(true),
		passwordKey: property.New("hunter2"),
		tagsKey: property.New(property.NewArray([]property.Value{
			property.New("a"), property.New("b"),
		})),
	})
	resp, err := r.Read(t.Context(), p.ReadRequest{
		ID:     "acme/thing-1",
		Inputs: oldInputs,
		Properties: propMap(map[string]any{
			"id": thing1ID, nameKey: "n1", valueKey: "local", statusKey: "ok",
		}),
	})
	if err != nil {
		t.Fatalf("read: %v\n  calls: %v", err, mock.calls)
	}

	if v, ok := resp.Inputs.GetOk(valueKey); !ok || v.AsString() != "remote" {
		t.Errorf("inputs.value: got %#v, want remote drift projected", v)
	} else if !v.Secret() {
		t.Errorf("inputs.value lost its secret marking")
	}
	if v, ok := resp.Inputs.GetOk(passwordKey); !ok || v.AsString() != "hunter2" {
		t.Errorf("inputs.password: got %#v, want prior value preserved", v)
	}
	if v, ok := resp.Inputs.GetOk(orgKey); !ok || v.AsString() != acmeVal {
		t.Errorf("inputs.org: got %#v, want %q", v, acmeVal)
	}
	for _, k := range []string{statusKey, "id"} {
		if _, ok := resp.Inputs.GetOk(k); ok {
			t.Errorf("inputs gained response-only key %q", k)
		}
	}
	tags, ok := resp.Inputs.GetOk(tagsKey)
	if !ok {
		t.Fatalf("inputs.tags missing")
	}
	var gotTags []string
	for _, v := range tags.AsArray().AsSlice() {
		gotTags = append(gotTags, v.AsString())
	}
	if !reflect.DeepEqual(gotTags, []string{"a", "z"}) {
		t.Errorf("inputs.tags: got %v, want [a z] projected and unordered-sorted", gotTags)
	}
}

// TestRefreshChainSurfacesRoleDetailsDrift replays the engine's refresh
// sequence for api:Role against the real spec: Read, then Diff with the
// Read-returned inputs as OldInputs (mirroring diffResource in
// pulumi/pulumi's RefreshStep). Out-of-band permission edits must surface
// as a diff on `details` (#938), and an unchanged server response must
// stay quiet.
func TestRefreshChainSurfacesRoleDetailsDrift(t *testing.T) {
	spec, meta := loadFixtures(t)
	resources := Resources(spec, meta)
	r, ok := resources["pulumiservice:api:Role"]
	if !ok {
		t.Fatalf("api:Role resource not found")
	}

	oldInputs := propMap(map[string]any{
		orgNameKey:     testOrgName,
		nameKey:        "perm-set",
		"uxPurpose":    "set",
		"resourceType": "global",
		detailsKey: map[string]any{
			"__type":      "PermissionDescriptorAllow",
			"permissions": []any{"organization:read_usage"},
		},
	})
	oldState := propMap(map[string]any{
		"roleID":       "role-123",
		nameKey:        "perm-set",
		"uxPurpose":    "set",
		"resourceType": "global",
		detailsKey: map[string]any{
			"__type":      "PermissionDescriptorAllow",
			"permissions": []any{"organization:read_usage"},
		},
	})

	cases := []struct {
		name       string
		details    string
		wantDrift  bool
		driftedKey string
	}{
		{
			name:       "console edit drifts details",
			details:    `{"__type":"PermissionDescriptorAllow","permissions":["organization:read_usage","organization:admin"]}`,
			wantDrift:  true,
			driftedKey: detailsKey,
		},
		{
			name:      "unchanged server state stays quiet",
			details:   `{"__type":"PermissionDescriptorAllow","permissions":["organization:read_usage"]}`,
			wantDrift: false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mock := &mockTransport{responses: map[string]mockResponse{
				"GET /api/orgs/test-org/roles/role-123": {
					status: 200,
					body: `{"id":"role-123","name":"perm-set","uxPurpose":"set","resourceType":"global",` +
						`"details":` + tc.details + `,"version":2,"orgId":"o1"}`,
				},
			}}
			SetTransportResolver(func(_ context.Context) (Transport, error) { return mock, nil })

			readResp, err := r.Read(t.Context(), p.ReadRequest{
				ID:         "test-org/role-123",
				Inputs:     oldInputs,
				Properties: oldState,
			})
			if err != nil {
				t.Fatalf("read: %v\n  calls: %v", err, mock.calls)
			}

			// RefreshStep passes the freshly-read inputs as OldInputs and the
			// pre-refresh inputs as Inputs, then inverts the result for display.
			diffResp, err := r.Diff(t.Context(), p.DiffRequest{
				ID:        "test-org/role-123",
				OldInputs: readResp.Inputs,
				State:     readResp.Properties,
				Inputs:    oldInputs,
			})
			if err != nil {
				t.Fatalf("diff: %v", err)
			}
			if diffResp.HasChanges != tc.wantDrift {
				t.Fatalf("HasChanges: got %v, want %v (detailed: %#v)",
					diffResp.HasChanges, tc.wantDrift, diffResp.DetailedDiff)
			}
			if tc.wantDrift {
				if _, ok := diffResp.DetailedDiff[tc.driftedKey]; !ok {
					t.Errorf("detailed diff missing %q: %#v", tc.driftedKey, diffResp.DetailedDiff)
				}
			}
		})
	}
}

// TestRefreshChainIgnoresServerAddedStackTags: GetStack returns system
// tags (pulumi:project etc.) alongside the program's own; key-scoped map
// projection keeps them out of inputs so refresh stays quiet, while drift
// on a program-written tag still surfaces.
func TestRefreshChainIgnoresServerAddedStackTags(t *testing.T) {
	spec, meta := loadFixtures(t)
	resources := Resources(spec, meta)
	r, ok := resources["pulumiservice:api/stacks:Stack"]
	if !ok {
		t.Fatalf("api/stacks:Stack resource not found")
	}

	oldInputs := propMap(map[string]any{
		orgNameKey:     testOrgName,
		projectNameKey: myprojVal,
		"stackName":    "dev",
		tagsKey:        map[string]any{ownerVal: "platform", "purpose": "demo"},
	})

	cases := []struct {
		name      string
		tags      string
		wantDrift bool
	}{
		{
			name:      "server-added system tags stay quiet",
			tags:      `{"owner":"platform","purpose":"demo","pulumi:project":"myproj","pulumi:runtime":"yaml"}`,
			wantDrift: false,
		},
		{
			name:      "drift on a program-written tag surfaces",
			tags:      `{"owner":"someone-else","purpose":"demo","pulumi:project":"myproj"}`,
			wantDrift: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mock := &mockTransport{responses: map[string]mockResponse{
				"GET /api/stacks/test-org/myproj/dev": {
					status: 200,
					body: `{"orgName":"test-org","projectName":"myproj","stackName":"dev",` +
						`"tags":` + tc.tags + `}`,
				},
			}}
			SetTransportResolver(func(_ context.Context) (Transport, error) { return mock, nil })

			readResp, err := r.Read(t.Context(), p.ReadRequest{
				ID:         "test-org/myproj/dev",
				Inputs:     oldInputs,
				Properties: oldInputs,
			})
			if err != nil {
				t.Fatalf("read: %v\n  calls: %v", err, mock.calls)
			}
			diffResp, err := r.Diff(t.Context(), p.DiffRequest{
				ID:        "test-org/myproj/dev",
				OldInputs: readResp.Inputs,
				State:     readResp.Properties,
				Inputs:    oldInputs,
			})
			if err != nil {
				t.Fatalf("diff: %v", err)
			}
			if diffResp.HasChanges != tc.wantDrift {
				t.Fatalf("HasChanges: got %v, want %v (detailed: %#v)",
					diffResp.HasChanges, tc.wantDrift, diffResp.DetailedDiff)
			}
			if tc.wantDrift {
				if _, ok := diffResp.DetailedDiff[tagsKey]; !ok {
					t.Errorf("detailed diff missing tags: %#v", diffResp.DetailedDiff)
				}
			}
		})
	}
}

// TestUpdateAliasesFlatNewFields: UpdateTeamRequest renames the fields it
// mutates (newDescription/newDisplayName); the update body must source
// them from the base-named inputs or the PATCH is a silent server-side
// no-op.
func TestUpdateAliasesFlatNewFields(t *testing.T) {
	spec, meta := loadFixtures(t)
	resources := Resources(spec, meta)
	r, ok := resources["pulumiservice:api/teams:Team"]
	if !ok {
		t.Fatalf("api/teams:Team resource not found")
	}

	var patchBody map[string]any
	mock := &mockTransport{
		responseFn: func(req *http.Request) mockResponse {
			switch req.Method {
			case http.MethodPatch:
				b, _ := io.ReadAll(req.Body)
				if err := json.Unmarshal(b, &patchBody); err != nil {
					t.Fatalf("unmarshal patch body %q: %v", b, err)
				}
				return mockResponse{status: 204, body: ""}
			case http.MethodGet:
				return mockResponse{status: 200, body: `{"name":"infra","displayName":"New Display","description":"new desc"}`}
			}
			return mockResponse{status: 500, body: unexpectedBody}
		},
	}
	SetTransportResolver(func(_ context.Context) (Transport, error) { return mock, nil })

	prior := propMap(map[string]any{
		orgNameKey:     testOrgName,
		nameKey:        infraVal,
		"displayName":  "Old Display",
		descriptionKey: "old desc",
	})
	newInputs := propMap(map[string]any{
		orgNameKey:     testOrgName,
		nameKey:        infraVal,
		"displayName":  "New Display",
		descriptionKey: newDescVal,
	})
	resp, err := r.Update(t.Context(), p.UpdateRequest{
		ID:        testOrgInfraID,
		OldInputs: prior,
		State:     prior,
		Inputs:    newInputs,
	})
	if err != nil {
		t.Fatalf("update: %v\n  calls: %v", err, mock.calls)
	}

	if got := patchBody[newDescriptionKey]; got != newDescVal {
		t.Errorf("patch newDescription: got %#v, want new desc", got)
	}
	if got := patchBody["newDisplayName"]; got != "New Display" {
		t.Errorf("patch newDisplayName: got %#v, want New Display", got)
	}
	if _, ok := patchBody[descriptionKey]; ok {
		t.Errorf("patch body carries off-schema %q key: %#v", descriptionKey, patchBody)
	}
	if v, ok := resp.Properties.GetOk(descriptionKey); !ok || v.AsString() != newDescVal {
		t.Errorf("state description: got %#v, want new desc", v)
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
			Operations: Operations{Create: createThingOp, Read: getThingOp},
			IDFormat:   orgIDFormat,
		},
	}
	mock := &mockTransport{responses: map[string]mockResponse{
		postThingsAcme: {status: 200, body: `{"id":"thing-1"}`},
		getThingPath:   {status: 200, body: `{"id":"thing-1","name":"foo","status":"ready"}`},
	}}
	SetTransportResolver(func(_ context.Context) (Transport, error) { return mock, nil })

	resp, err := r.Create(t.Context(), p.CreateRequest{
		Properties: propMap(map[string]any{orgKey: acmeVal, nameKey: fooVal}),
	})
	if err != nil {
		t.Fatalf("create: %v\n  calls: %v", err, mock.calls)
	}
	if len(mock.calls) != 2 || mock.calls[1] != getThingPath {
		t.Errorf("expected POST then GET /things/acme/thing-1, got: %v", mock.calls)
	}
	if v, ok := resp.Properties.GetOk(statusKey); !ok || v.AsString() != "ready" {
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
			Operations: Operations{Create: createThingOp, Read: getThingOp},
			IDFormat:   orgIDFormat,
		},
	}
	readBody := `{"id":"thing-1","name":"foo","lastUpdate":"2026-05-05T00:00:00Z","status":"ready"}`
	mock := &mockTransport{responses: map[string]mockResponse{
		postThingsAcme: {status: 200, body: `{"id":"thing-1"}`},
		getThingPath:   {status: 200, body: readBody},
	}}
	SetTransportResolver(func(_ context.Context) (Transport, error) { return mock, nil })

	resp, err := r.Create(t.Context(), p.CreateRequest{
		Properties: propMap(map[string]any{orgKey: acmeVal, nameKey: fooVal}),
	})
	if err != nil {
		t.Fatalf("create: %v\n  calls: %v", err, mock.calls)
	}

	if len(mock.calls) != 2 || mock.calls[0] != postThingsAcme || mock.calls[1] != getThingPath {
		t.Errorf("expected POST then GET, got: %v", mock.calls)
	}
	for _, key := range []string{"lastUpdate", statusKey} {
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
	if _, ok := resp.Properties.GetOk(orgKey); ok {
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
			Operations:    Operations{Create: putThingOp, Read: getThingOp, Update: putThingOp},
			IDFormat:      orgFormat,
			RequireImport: true,
		},
	}
	mock := &mockTransport{responses: map[string]mockResponse{
		getThingsAcme: {status: 200, body: `{"org":"acme","value":"existing"}`},
	}}
	SetTransportResolver(func(_ context.Context) (Transport, error) { return mock, nil })

	_, err := r.Create(t.Context(), p.CreateRequest{
		Properties: propMap(map[string]any{orgKey: acmeVal, valueKey: newVal}),
	})
	if err == nil {
		t.Fatalf("expected error when resource already exists, got nil")
	}
	if !strings.Contains(err.Error(), "already exists") || !strings.Contains(err.Error(), "import") {
		t.Errorf("error message should mention `already exists` and `import`: %v", err)
	}
	if len(mock.calls) != 1 || mock.calls[0] != getThingsAcme {
		t.Errorf("expected exactly one GET, got: %v", mock.calls)
	}
}

// TestCreateRequireImport_ProceedsOn404: a 404 from the probe means proceed.
func TestCreateRequireImport_ProceedsOn404(t *testing.T) {
	spec := requireImportSpec(t)
	r := &Resource{
		spec: spec,
		meta: ResourceMeta{
			Operations:    Operations{Create: putThingOp, Read: getThingOp, Update: putThingOp},
			IDFormat:      orgFormat,
			RequireImport: true,
		},
	}
	var putCalled bool
	mock := &mockTransport{responseFn: func(req *http.Request) mockResponse {
		switch req.Method {
		case http.MethodPut:
			putCalled = true
			return mockResponse{status: 200, body: orgAcmeNewBody}
		case http.MethodGet:
			if putCalled {
				return mockResponse{status: 200, body: orgAcmeNewBody}
			}
			return mockResponse{status: 404, body: notFoundBody}
		}
		return mockResponse{status: 500, body: unexpectedBody}
	}}
	SetTransportResolver(func(_ context.Context) (Transport, error) { return mock, nil })

	resp, err := r.Create(t.Context(), p.CreateRequest{
		Properties: propMap(map[string]any{orgKey: acmeVal, valueKey: newVal}),
	})
	if err != nil {
		t.Fatalf("create should proceed on 404, got: %v\n  calls: %v", err, mock.calls)
	}
	if resp.ID != acmeVal {
		t.Errorf("ID: got %q, want %q", resp.ID, acmeVal)
	}
	wantCalls := []string{getThingsAcme, putThingsAcme, getThingsAcme}
	if len(mock.calls) != len(wantCalls) {
		t.Fatalf("expected %d calls, got %d: %v", len(wantCalls), len(mock.calls), mock.calls)
	}
	for i, want := range wantCalls {
		if mock.calls[i] != want {
			t.Errorf("call %d: got %q, want %q", i, mock.calls[i], want)
		}
	}
}

// TestCreateRequireImport_ServerIDReadURL_MetadataError: if requireImport is
// enabled but the read URL needs a path param the user can't supply pre-create
// (typically a server-generated ID), surface a metadata-error message instead
// of the generic "requireImport probe" wrap so the operator knows what to fix.
func TestCreateRequireImport_ServerIDReadURL_MetadataError(t *testing.T) {
	const specJSON = `{
	  "openapi": "3.0.0",
	  "components": {"schemas": {
	    "Body": {"type": "object", "properties": {"value": {"type": "string"}}},
	    "Read": {"type": "object", "properties": {"id": {"type": "string"}, "value": {"type": "string"}}}
	  }},
	  "paths": {
	    "/things": {
	      "post": {
	        "operationId": "CreateThing",
	        "requestBody": {"content": {"application/json": {"schema": {"$ref": "#/components/schemas/Body"}}}},
	        "responses": {"200": {"content": {"application/json": {"schema": {"$ref": "#/components/schemas/Read"}}}}}
	      }
	    },
	    "/things/{id}": {
	      "get": {
	        "operationId": "ReadThing",
	        "parameters": [{"name": "id", "in": "path", "required": true, "schema": {"type": "string"}}],
	        "responses": {"200": {"content": {"application/json": {"schema": {"$ref": "#/components/schemas/Read"}}}}}
	      }
	    }
	  }
	}`
	spec, err := ParseSpec([]byte(specJSON))
	if err != nil {
		t.Fatalf("ParseSpec: %v", err)
	}
	r := &Resource{
		spec: spec,
		meta: ResourceMeta{ //nolint:gosec // G101: test fixture, not a real credential.
			Operations:    Operations{Create: createThingOp, Read: "ReadThing"},
			Token:         "pulumiservice:api:Thing",
			RequireImport: true,
		},
	}
	mock := &mockTransport{}
	SetTransportResolver(func(_ context.Context) (Transport, error) { return mock, nil })

	_, err = r.Create(t.Context(), p.CreateRequest{
		Properties: propMap(map[string]any{valueKey: newVal}),
	})
	if err == nil {
		t.Fatalf("expected metadata-error, got nil")
	}
	if !strings.Contains(err.Error(), "metadata error") {
		t.Errorf("error must signal a metadata issue: %v", err)
	}
	if !strings.Contains(err.Error(), "pulumiservice:api:Thing") {
		t.Errorf("error must name the resource token: %v", err)
	}
	if len(mock.calls) != 0 {
		t.Errorf("no HTTP calls should fire when path param missing, got: %v", mock.calls)
	}
}

// TestCreateRequireImport_NoReadOp_OptsOut: RequireImport is a no-op
// without a read op (the dispatch can't probe what it can't read).
func TestCreateRequireImport_NoReadOp_OptsOut(t *testing.T) {
	spec := requireImportSpec(t)
	r := &Resource{
		spec: spec,
		meta: ResourceMeta{
			Operations:    Operations{Create: putThingOp, Update: putThingOp},
			IDFormat:      orgFormat,
			RequireImport: true,
		},
	}
	mock := &mockTransport{responses: map[string]mockResponse{
		putThingsAcme: {status: 200, body: orgAcmeNewBody},
	}}
	SetTransportResolver(func(_ context.Context) (Transport, error) { return mock, nil })

	_, err := r.Create(t.Context(), p.CreateRequest{
		Properties: propMap(map[string]any{orgKey: acmeVal, valueKey: newVal}),
	})
	if err != nil {
		t.Fatalf("create should proceed without probe, got: %v\n  calls: %v", err, mock.calls)
	}
	if len(mock.calls) != 1 || mock.calls[0] != putThingsAcme {
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
		orgNameKey: testOrgName,
	})}
	_, err := r.Create(t.Context(), req)
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
			Operations: Operations{Read: getThingOp, Update: "PatchThing"},
			IDFormat:   orgIDFormat,
		},
	}
	patchBody := `{"name":"foo-renamed","description":"rotated"}`
	readBody := `{"id":"thing-1","name":"foo-renamed","description":"rotated","created":"2026-05-05T00:00:00Z"}`
	priorState := map[string]any{
		"id": thing1ID, nameKey: fooVal, descriptionKey: originalVal,
		createdKey: createdTimestamp,
	}
	mock := &mockTransport{responses: map[string]mockResponse{
		patchThingPath: {status: 200, body: patchBody},
		getThingPath:   {status: 200, body: readBody},
	}}
	SetTransportResolver(func(_ context.Context) (Transport, error) { return mock, nil })

	resp, err := r.Update(t.Context(), p.UpdateRequest{
		Inputs:    propMap(map[string]any{orgKey: acmeVal, nameKey: "foo-renamed", descriptionKey: rotatedVal}),
		OldInputs: propMap(map[string]any{orgKey: acmeVal, nameKey: fooVal, descriptionKey: originalVal}),
		State:     propMap(priorState),
	})
	if err != nil {
		t.Fatalf("update: %v\n  calls: %v", err, mock.calls)
	}
	if len(mock.calls) != 2 || mock.calls[0] != patchThingPath || mock.calls[1] != getThingPath {
		t.Errorf("expected PATCH then GET, got: %v", mock.calls)
	}
	for _, key := range []string{"id", createdKey, nameKey, descriptionKey} {
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
			IDFormat:   orgIDFormat,
		},
	}
	mock := &mockTransport{responses: map[string]mockResponse{
		patchThingPath: {status: 200, body: `{"name":"foo","description":"rotated"}`},
	}}
	SetTransportResolver(func(_ context.Context) (Transport, error) { return mock, nil })

	priorState := map[string]any{
		"id": thing1ID, nameKey: fooVal, descriptionKey: originalVal,
		createdKey: createdTimestamp,
	}
	resp, err := r.Update(t.Context(), p.UpdateRequest{
		Inputs:    propMap(map[string]any{orgKey: acmeVal, descriptionKey: rotatedVal}),
		OldInputs: propMap(map[string]any{orgKey: acmeVal, descriptionKey: originalVal}),
		State:     propMap(priorState),
	})
	if err != nil {
		t.Fatalf("update: %v\n  calls: %v", err, mock.calls)
	}
	if v, ok := resp.Properties.GetOk("id"); !ok || v.AsString() != thing1ID {
		t.Errorf("state missing or wrong `id` after update-with-no-read-op: %v (ok=%v)", v, ok)
	}
	if v, ok := resp.Properties.GetOk(createdKey); !ok || v.AsString() != createdTimestamp {
		t.Errorf("state missing prior `created` after update-with-no-read-op: %v (ok=%v)", v, ok)
	}
	if v, ok := resp.Properties.GetOk(descriptionKey); !ok || v.AsString() != rotatedVal {
		t.Errorf("state has stale description, want %q got %v (ok=%v)", rotatedVal, v, ok)
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
			inputs:  map[string]any{orgKey: acmeVal, nameKey: infraVal, descriptionKey: infraTeamVal},
			want:    map[string]any{nameKey: infraVal, descriptionKey: infraTeamVal},
		},
		{
			name:    "rename for path-param does not rewrite body field of same Pulumi name",
			op:      "CreateTeam",
			renames: map[string]string{nameKey: "teamName"}, // path-only rename
			inputs:  map[string]any{orgKey: acmeVal, nameKey: infraVal, descriptionKey: infraTeamVal},
			want:    map[string]any{nameKey: infraVal, descriptionKey: infraTeamVal},
		},
		{
			name:    "rename that targets a body field is applied",
			op:      "CreateWidget",
			renames: map[string]string{"id": widgetIDVal}, // body field rename
			inputs:  map[string]any{orgKey: acmeVal, "id": "w-1", descriptionKey: "wodget"},
			want:    map[string]any{widgetIDVal: "w-1", descriptionKey: "wodget"},
		},
		{
			name: "path param ALSO declared in body schema is kept (server validates match)",
			op:   "CreateOrgHook",
			// OrganizationWebhook-style: Pulumi-side `organizationName` maps to
			// wire path-param `orgName`, but the body schema also declares a
			// wire-side `organizationName`. The body must carry it.
			renames: map[string]string{organizationNameVal: orgNameKey},
			inputs: map[string]any{
				organizationNameVal: acmeVal, nameKey: "alerts", "payloadUrl": "https://x",
			},
			want: map[string]any{
				organizationNameVal: acmeVal, nameKey: "alerts", "payloadUrl": "https://x",
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
// treated as success. Centralizing this means every api resource is
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
			IDFormat:   orgIDFormat,
		},
	}

	t.Run("204 is success", func(t *testing.T) {
		mock := &mockTransport{responses: map[string]mockResponse{
			deleteThingsAcmeGone: {status: 204},
		}}
		SetTransportResolver(func(_ context.Context) (Transport, error) { return mock, nil })
		err := r.Delete(t.Context(), p.DeleteRequest{
			ID:         acmeGoneID,
			Properties: propMap(map[string]any{orgKey: acmeVal, "id": goneVal}),
		})
		if err != nil {
			t.Errorf("204 should be nil, got: %v", err)
		}
	})

	t.Run("404 is also success — already gone is what Delete wants", func(t *testing.T) {
		mock := &mockTransport{responses: map[string]mockResponse{
			deleteThingsAcmeGone: {status: 404, body: `{"code":404,"message":"not found"}`},
		}}
		SetTransportResolver(func(_ context.Context) (Transport, error) { return mock, nil })
		err := r.Delete(t.Context(), p.DeleteRequest{
			ID:         acmeGoneID,
			Properties: propMap(map[string]any{orgKey: acmeVal, "id": goneVal}),
		})
		if err != nil {
			t.Errorf("404 should be nil, got: %v", err)
		}
	})

	t.Run("non-404 errors still surface", func(t *testing.T) {
		mock := &mockTransport{responses: map[string]mockResponse{
			deleteThingsAcmeGone: {status: 500, body: `oops`},
		}}
		SetTransportResolver(func(_ context.Context) (Transport, error) { return mock, nil })
		err := r.Delete(t.Context(), p.DeleteRequest{
			ID:         acmeGoneID,
			Properties: propMap(map[string]any{orgKey: acmeVal, "id": goneVal}),
		})
		if err == nil {
			t.Error("500 should propagate")
		}
	})
}

// TestUpdateUsesStateForPathParamsAndInputsForBody pins the split-source
// semantics from Bug 6: path params come from State (server-resolved IDs
// must win over any stale/imported Inputs value), body fields come from
// Inputs (the user's new values). A common shape: Read returns `id`
// which Pulumi renames to `widgetID`; if a user-supplied Inputs.widgetID
// existed it shouldn't redirect the PATCH URL.
func TestUpdateUsesStateForPathParamsAndInputsForBody(t *testing.T) {
	const specJSON = `{
	  "openapi": "3.0.0",
	  "components": {"schemas": {
	    "WidgetPatch": {"type": "object", "properties": {
	      "name": {"type": "string"}
	    }},
	    "WidgetRead": {"type": "object", "properties": {
	      "id":   {"type": "string"},
	      "name": {"type": "string"}
	    }}
	  }},
	  "paths": {
	    "/widgets/{org}/{widgetID}": {
	      "patch": {
	        "operationId": "PatchWidget",
	        "parameters": [
	          {"name": "org",      "in": "path", "required": true, "schema": {"type": "string"}},
	          {"name": "widgetID", "in": "path", "required": true, "schema": {"type": "string"}}
	        ],
	        "requestBody": {"content": {"application/json": {"schema": {"$ref": "#/components/schemas/WidgetPatch"}}}},
	        "responses": {"200": {"content": {"application/json": {
	          "schema": {"$ref": "#/components/schemas/WidgetRead"}
	        }}}}
	      },
	      "get": {
	        "operationId": "GetWidget",
	        "parameters": [
	          {"name": "org",      "in": "path", "required": true, "schema": {"type": "string"}},
	          {"name": "widgetID", "in": "path", "required": true, "schema": {"type": "string"}}
	        ],
	        "responses": {"200": {"content": {"application/json": {
	          "schema": {"$ref": "#/components/schemas/WidgetRead"}
	        }}}}
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
			Operations: Operations{Update: "PatchWidget", Read: "GetWidget"},
			IDFormat:   "{org}/{widgetID}",
		},
	}

	// State has the real server-resolved widgetID; Inputs has a stale
	// value the user shouldn't be able to redirect the PATCH with.
	state := propMap(map[string]any{
		orgKey:      acmeVal,
		widgetIDVal: "real-7",
		nameKey:     "old-name",
	})
	inputs := propMap(map[string]any{
		orgKey:      acmeVal,
		widgetIDVal: "stale-from-user", // should be ignored for the URL
		nameKey:     "new-name",
	})
	oldInputs := propMap(map[string]any{
		orgKey:      acmeVal,
		widgetIDVal: "stale-from-user",
		nameKey:     "old-name",
	})

	mock := &mockTransport{responses: map[string]mockResponse{
		"PATCH /widgets/acme/real-7": {status: 200, body: `{"id":"real-7","name":"new-name"}`},
		"GET /widgets/acme/real-7":   {status: 200, body: `{"id":"real-7","name":"new-name"}`},
	}}
	SetTransportResolver(func(_ context.Context) (Transport, error) { return mock, nil })

	_, err = r.Update(t.Context(), p.UpdateRequest{
		ID:        "acme/real-7",
		Inputs:    inputs,
		OldInputs: oldInputs,
		State:     state,
	})
	if err != nil {
		t.Fatalf("Update: %v\n  calls: %v", err, mock.calls)
	}
	// Confirm the PATCH used the State-side widgetID and the GET (read-
	// after-update) did too. Without the split, the URL would be
	// /widgets/acme/stale-from-user.
	if len(mock.calls) < 1 || mock.calls[0] != "PATCH /widgets/acme/real-7" {
		t.Errorf("expected PATCH to State-side path; calls: %v", mock.calls)
	}
	if len(mock.calls) < 2 || mock.calls[1] != "GET /widgets/acme/real-7" {
		t.Errorf("expected read-after-update GET on State-side path; calls: %v", mock.calls)
	}
}

// TestDiffEmitsUpdateOnUnknownInput pins the contract: when an input
// transitions from a known value to an unknown (computed) value — or
// vice-versa — Diff must emit an entry, not silently skip the key.
// Otherwise users recovering from a partial apply see misleading
// "no changes" plans against still-unresolved values.
//
// AllStable in pulumi/property iterates every key (including computed
// ones); the safety here comes from Value.Equals returning false when
// comparing computed to non-computed by default.
func TestDiffEmitsUpdateOnUnknownInput(t *testing.T) {
	const specJSON = `{
	  "openapi": "3.0.0",
	  "paths": {
	    "/things/{org}": {
	      "post": {
	        "operationId": "CreateThing",
	        "parameters": [{"name": "org", "in": "path", "required": true, "schema": {"type": "string"}}],
	        "requestBody": {"content": {"application/json": {"schema": {
	          "type": "object", "properties": {"x": {"type": "string"}}
	        }}}},
	        "responses": {"200": {"description": "OK"}}
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
		meta: ResourceMeta{Operations: Operations{Create: createThingOp}, IDFormat: orgFormat},
	}

	cases := []struct {
		name string
		old  map[string]any
		new  map[string]any
		want string // expected DetailedDiff key
	}{
		{
			name: "known → unknown emits a diff",
			old:  map[string]any{orgKey: acmeVal, "x": "old"},
			new:  map[string]any{orgKey: acmeVal, "x": property.Computed},
			want: "x",
		},
		{
			name: "unknown → known emits a diff",
			old:  map[string]any{orgKey: acmeVal, "x": property.Computed},
			new:  map[string]any{orgKey: acmeVal, "x": newVal},
			want: "x",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := r.Diff(t.Context(), p.DiffRequest{
				ID:        acmeVal,
				OldInputs: propMap(tc.old),
				Inputs:    propMap(tc.new),
			})
			if err != nil {
				t.Fatalf("Diff: %v", err)
			}
			if !resp.HasChanges {
				t.Fatalf("expected HasChanges=true, got false (DetailedDiff=%v)", resp.DetailedDiff)
			}
			if _, ok := resp.DetailedDiff[tc.want]; !ok {
				t.Errorf("DetailedDiff missing %q; got keys: %v", tc.want, mapKeys(resp.DetailedDiff))
			}
		})
	}
}

func mapKeys[V any](m map[string]V) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

// TestErrorURLIsPostTransportRewrite pins the user-facing error path: the
// real backend host shows up in error messages, not the transport.invalid
// sentinel that buildURL uses by default when the spec lacks servers.
// authedTransport in production rewrites scheme+host on the request in
// place; this test simulates that with a tiny shim transport.
func TestErrorURLIsPostTransportRewrite(t *testing.T) {
	const specJSON = `{
	  "openapi": "3.0.0",
	  "paths": {
	    "/things/{org}": {
	      "post": {
	        "operationId": "CreateThing",
	        "parameters": [{"name": "org", "in": "path", "required": true, "schema": {"type": "string"}}],
	        "responses": {"500": {"description": "boom"}}
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
		meta: ResourceMeta{Operations: Operations{Create: createThingOp}, IDFormat: orgFormat},
	}
	// Rewriting shim: mimics authedTransport.Do — overwrites scheme+host
	// and returns a 500 so execAndDecode constructs an HTTPError.
	SetTransportResolver(func(_ context.Context) (Transport, error) {
		return rewriteTransport{scheme: "https", host: "api.real-backend.example"}, nil
	})

	_, err = r.Create(t.Context(), p.CreateRequest{
		Properties: propMap(map[string]any{orgKey: acmeVal}),
	})
	if err == nil {
		t.Fatal("expected error from 500 response")
	}
	if strings.Contains(err.Error(), "transport.invalid") {
		t.Errorf("error leaks transport.invalid sentinel: %v", err)
	}
	if !strings.Contains(err.Error(), "api.real-backend.example") {
		t.Errorf("error should reference rewritten host; got: %v", err)
	}
}

type rewriteTransport struct{ scheme, host string }

func (t rewriteTransport) Do(_ context.Context, req *http.Request) (*http.Response, error) {
	req.URL.Scheme = t.scheme
	req.URL.Host = t.host
	req.Host = t.host
	return &http.Response{
		StatusCode: 500,
		Body:       io.NopCloser(strings.NewReader(`{"code":500,"message":"boom"}`)),
	}, nil
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

// envTagProps builds the standard EnvironmentTag input/state map with the
// given tag value.
func envTagProps(value string) property.Map {
	return propMap(map[string]any{
		orgNameKey:     testOrgName,
		projectNameKey: myprojVal,
		envNameKey:     myenvVal,
		nameKey:        ownerVal,
		valueKey:       value,
	})
}

// TestUpdateEnvelopeSendsCurrentAndNewBody pins the wire contract for update
// ops declared with updateEnvelope (EnvironmentTag): the PATCH body pairs the
// prior value under currentField with the desired values under newField. The
// flat schema-driven mapping would send an empty body here — neither wrapper
// field matches an input — so in-place value updates would always fail.
func TestUpdateEnvelopeSendsCurrentAndNewBody(t *testing.T) {
	spec, meta := loadFixtures(t)
	resources := Resources(spec, meta)
	r, ok := resources["pulumiservice:api/esc:EnvironmentTag"]
	if !ok {
		t.Fatal("pulumiservice:api/esc:EnvironmentTag not in metadata")
	}
	if r.meta.UpdateEnvelope == nil {
		t.Fatal("EnvironmentTag metadata entry lost its updateEnvelope")
	}

	var patchBody []byte
	mock := &mockTransport{responseFn: func(req *http.Request) mockResponse {
		if req.Method == http.MethodPatch {
			patchBody, _ = io.ReadAll(req.Body)
		}
		return mockResponse{status: 200, body: `{"name":"owner","value":"team-y"}`}
	}}
	ctx := WithTransport(t.Context(), mock)

	if _, err := r.Update(ctx, p.UpdateRequest{
		ID:        "test-org/myproj/myenv/owner",
		Inputs:    envTagProps(teamYVal),
		OldInputs: envTagProps(teamXVal),
		State:     envTagProps(teamXVal),
	}); err != nil {
		t.Fatalf("Update: %v\n  calls: %v", err, mock.calls)
	}

	wantPatch := "PATCH /api/esc/environments/test-org/myproj/myenv/tags/owner"
	if len(mock.calls) == 0 || mock.calls[0] != wantPatch {
		t.Fatalf("expected first call %q; calls: %v", wantPatch, mock.calls)
	}
	var got map[string]any
	if err := json.Unmarshal(patchBody, &got); err != nil {
		t.Fatalf("unmarshal PATCH body %q: %v", patchBody, err)
	}
	want := map[string]any{
		"currentTag": map[string]any{valueKey: teamXVal},
		newTagKey:    map[string]any{nameKey: ownerVal, valueKey: teamYVal},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("PATCH body = %s, want %v", patchBody, want)
	}
}

// TestUpdateEnvelopeValidation pins the loud-failure matrix: any mismatch
// between the declared envelope and the actual schema or sources must error
// instead of silently sending a partial body.
func TestUpdateEnvelopeValidation(t *testing.T) {
	spec, meta := loadFixtures(t)
	base := meta.Resources["pulumiservice:api:EnvironmentTag_esc_environments"]
	if base.UpdateEnvelope == nil {
		t.Fatal("EnvironmentTag metadata entry lost its updateEnvelope")
	}

	full := envTagProps(teamYVal)
	noValue := propMap(map[string]any{
		orgNameKey:     testOrgName,
		projectNameKey: myprojVal,
		envNameKey:     myenvVal,
		nameKey:        ownerVal,
	})

	cases := []struct {
		name     string
		envelope *UpdateEnvelopeMeta // nil keeps the real metadata envelope
		src      property.Map
		wantErr  string
	}{
		{"identical wrapper fields", &UpdateEnvelopeMeta{CurrentField: newTagKey, NewField: newTagKey}, full,
			"two distinct fields"},
		{"envelope drifted from schema", &UpdateEnvelopeMeta{CurrentField: "nope", NewField: newTagKey}, full,
			`"currentTag" is outside the declared updateEnvelope`},
		{"missing required wrapper prop", nil, noValue,
			`missing required field "value"`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rm := base
			if tc.envelope != nil {
				rm.UpdateEnvelope = tc.envelope
			}
			r := &Resource{meta: rm, spec: spec}
			ctx := WithTransport(t.Context(), &mockTransport{})
			_, err := r.Update(ctx, p.UpdateRequest{
				ID:        "test-org/myproj/myenv/owner",
				Inputs:    tc.src,
				OldInputs: tc.src,
				State:     tc.src,
			})
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("want error containing %q, got %v", tc.wantErr, err)
			}
		})
	}
}

// TestUpdateEnvelopeAllOfWrapper ensures wrapper fields declared as allOf
// $ref compositions (a common generator idiom for attaching descriptions)
// resolve the same as bare $refs.
func TestUpdateEnvelopeAllOfWrapper(t *testing.T) {
	const specJSON = `{
	  "openapi": "3.0.0",
	  "components": {"schemas": {
	    "CurrentThing": {"type": "object", "properties": {"value": {"type": "string"}}, "required": ["value"]},
	    "NewThing":     {"type": "object", "properties": {"value": {"type": "string"}}},
	    "Envelope": {"type": "object", "properties": {
	      "currentThing": {"allOf": [{"$ref": "#/components/schemas/CurrentThing"}], "description": "prior"},
	      "newThing":     {"$ref": "#/components/schemas/NewThing"}
	    }}
	  }},
	  "paths": {
	    "/things/{id}": {"patch": {
	      "operationId": "UpdateThing",
	      "parameters": [{"name": "id", "in": "path", "required": true, "schema": {"type": "string"}}],
	      "requestBody": {"content": {"application/json": {"schema": {"$ref": "#/components/schemas/Envelope"}}}},
	      "responses": {"204": {}}
	    }}
	  }
	}`
	spec, err := ParseSpec([]byte(specJSON))
	if err != nil {
		t.Fatalf("parse synthetic spec: %v", err)
	}
	r := &Resource{
		spec: spec,
		meta: ResourceMeta{
			Operations:     Operations{Update: "UpdateThing"},
			UpdateEnvelope: &UpdateEnvelopeMeta{CurrentField: "currentThing", NewField: "newThing"},
		},
	}

	var patchBody []byte
	mock := &mockTransport{responseFn: func(req *http.Request) mockResponse {
		patchBody, _ = io.ReadAll(req.Body)
		return mockResponse{status: 204, body: ""}
	}}
	ctx := WithTransport(t.Context(), mock)

	src := func(v string) property.Map { return propMap(map[string]any{"id": thing1ID, valueKey: v}) }
	if _, err := r.Update(ctx, p.UpdateRequest{
		ID:        thing1ID,
		Inputs:    src(newVal),
		OldInputs: src(originalVal),
		State:     src(originalVal),
	}); err != nil {
		t.Fatalf("Update: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal(patchBody, &got); err != nil {
		t.Fatalf("unmarshal PATCH body %q: %v", patchBody, err)
	}
	want := map[string]any{
		"currentThing": map[string]any{valueKey: originalVal},
		"newThing":     map[string]any{valueKey: newVal},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("PATCH body = %s, want %v", patchBody, want)
	}
}

// TestProjectValueArrayElements: array items must project element-wise like
// maps do. GetServiceResponse.items (ServiceItem) is richer than
// CreateServiceRequest.items (AddServiceItem), so adopting the state array
// wholesale writes response-only fields into inputs as permanent drift.
func TestProjectValueArrayElements(t *testing.T) {
	prior := anyToPropertyValue([]any{map[string]any{nameKey: infraVal, typeKey: stackVal}})
	state := anyToPropertyValue([]any{map[string]any{
		nameKey:       infraVal,
		typeKey:       stackVal,
		createdKey:    createdTimestamp,
		cloudCountKey: 3.0,
	}})

	got := projectValue(prior, state)
	if !got.IsArray() {
		t.Fatalf("projected value is not an array: %v", got)
	}
	elem := got.AsArray().Get(0)
	if !elem.IsMap() {
		t.Fatalf("element is not a map: %v", elem)
	}
	for _, k := range []string{createdKey, cloudCountKey} {
		if _, ok := elem.AsMap().GetOk(k); ok {
			t.Errorf("response-only key %q leaked into projected inputs: %v", k, elem)
		}
	}
	if v, ok := elem.AsMap().GetOk(nameKey); !ok || v.AsString() != infraVal {
		t.Errorf("program-owned key %q lost: %v", nameKey, elem)
	}
}

// TestProjectValuePreservesNestedSecrets: Value.Secret() is not recursive,
// so rebuilding a map during projection must re-apply each leaf's prior
// secret marking or a user's pulumi.secret() value silently goes plaintext.
func TestProjectValuePreservesNestedSecrets(t *testing.T) {
	prior := property.New(property.NewMap(map[string]property.Value{
		envVarsKey: property.New(property.NewMap(map[string]property.Value{
			passwordKey: property.New("s3cret").WithSecret(true),
		})),
	}))
	state := property.New(property.NewMap(map[string]property.Value{
		envVarsKey: property.New(property.NewMap(map[string]property.Value{
			passwordKey: property.New("s3cret"),
		})),
	}))

	got := projectValue(prior, state)
	leaf := got.AsMap().Get(envVarsKey).AsMap().Get(passwordKey)
	if !leaf.Secret() {
		t.Errorf("nested secret marking dropped: %v", got)
	}
}

// TestUpdateSkipsAliasForReplaceTriggeringField: PolicyGroup's name is the
// {policyGroup} path param, so Diff replaces on a name change and the
// newName alias is unreachable. It must not fire on unrelated updates.
func TestUpdateSkipsAliasForReplaceTriggeringField(t *testing.T) {
	spec, meta := loadFixtures(t)
	r, ok := Resources(spec, meta)["pulumiservice:api:PolicyGroup"]
	if !ok {
		t.Fatalf("PolicyGroup resource not found")
	}

	var patchBody map[string]any
	mock := &mockTransport{
		responseFn: func(req *http.Request) mockResponse {
			switch req.Method {
			case http.MethodPatch:
				b, _ := io.ReadAll(req.Body)
				if err := json.Unmarshal(b, &patchBody); err != nil {
					t.Fatalf("unmarshal patch body %q: %v", b, err)
				}
				return mockResponse{status: 204, body: ""}
			case http.MethodGet:
				return mockResponse{status: 200, body: `{"name":"pg1"}`}
			}
			return mockResponse{status: 500, body: unexpectedBody}
		},
	}
	SetTransportResolver(func(_ context.Context) (Transport, error) { return mock, nil })

	prior := propMap(map[string]any{orgNameKey: testOrgName, nameKey: "pg1", "agentPoolId": "pool-a"})
	newInputs := propMap(map[string]any{orgNameKey: testOrgName, nameKey: "pg1", "agentPoolId": "pool-b"})
	if _, err := r.Update(t.Context(), p.UpdateRequest{
		ID: "test-org/pg1", OldInputs: prior, State: prior, Inputs: newInputs,
	}); err != nil {
		t.Fatalf("update: %v\n  calls: %v", err, mock.calls)
	}
	if v, ok := patchBody["newName"]; ok {
		t.Errorf("newName=%v sent though name is unchanged and replace-triggering; body=%v", v, patchBody)
	}
}

// TestUpdateExplicitWireNameWinsOverAlias: aliases only fill fields with no
// direct source, so an explicit new<Stem> input takes precedence over the
// base-named one.
func TestUpdateExplicitWireNameWinsOverAlias(t *testing.T) {
	spec, meta := loadFixtures(t)
	r, ok := Resources(spec, meta)["pulumiservice:api/teams:Team"]
	if !ok {
		t.Fatalf("api/teams:Team resource not found")
	}

	var patchBody map[string]any
	mock := &mockTransport{
		responseFn: func(req *http.Request) mockResponse {
			switch req.Method {
			case http.MethodPatch:
				b, _ := io.ReadAll(req.Body)
				if err := json.Unmarshal(b, &patchBody); err != nil {
					t.Fatalf("unmarshal patch body %q: %v", b, err)
				}
				return mockResponse{status: 204, body: ""}
			case http.MethodGet:
				return mockResponse{status: 200, body: `{"name":"infra","description":"explicit desc"}`}
			}
			return mockResponse{status: 500, body: unexpectedBody}
		},
	}
	SetTransportResolver(func(_ context.Context) (Transport, error) { return mock, nil })

	prior := propMap(map[string]any{
		orgNameKey: testOrgName, nameKey: infraVal, descriptionKey: "old desc",
	})
	newInputs := propMap(map[string]any{
		orgNameKey: testOrgName, nameKey: infraVal,
		descriptionKey:    newDescVal,
		newDescriptionKey: "explicit desc",
	})
	if _, err := r.Update(t.Context(), p.UpdateRequest{
		ID: testOrgInfraID, OldInputs: prior, State: prior, Inputs: newInputs,
	}); err != nil {
		t.Fatalf("update: %v\n  calls: %v", err, mock.calls)
	}
	if got := patchBody[newDescriptionKey]; got != "explicit desc" {
		t.Errorf("explicit newDescription overridden by alias: got %#v, want explicit desc", got)
	}
}

// TestReadKeepsResponseOnlyItemFieldsOutOfInputs: GetService returns
// ServiceItem (created, cloudCount, ...) where CreateService accepts only
// AddServiceItem (name, type). Refresh must not adopt the response-only
// half, or every subsequent preview reports a phantom items update.
func TestReadKeepsResponseOnlyItemFieldsOutOfInputs(t *testing.T) {
	spec, meta := loadFixtures(t)
	r, ok := Resources(spec, meta)["pulumiservice:api/services:Service"]
	if !ok {
		t.Fatalf("api/services:Service resource not found")
	}

	SetTransportResolver(func(_ context.Context) (Transport, error) {
		return &mockTransport{responseFn: func(_ *http.Request) mockResponse {
			return mockResponse{status: 200, body: `{
				"items":[{"name":"infra","type":"stack","created":"2026-05-05T00:00:00Z","cloudCount":3}],
				"service":{"name":"svc-1"}
			}`}
		}}, nil
	})

	inputs := propMap(map[string]any{
		orgNameKey:  testOrgName,
		"ownerType": "team",
		"ownerName": teamXVal,
		nameKey:     "svc-1",
		itemsKey:    []any{map[string]any{nameKey: infraVal, typeKey: stackVal}},
	})
	resp, err := r.Read(t.Context(), p.ReadRequest{
		ID: "test-org/team/team-x/svc-1", Inputs: inputs, Properties: inputs,
	})
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	item := resp.Inputs.Get(itemsKey).AsArray().Get(0)
	for _, k := range []string{createdKey, cloudCountKey} {
		if _, ok := item.AsMap().GetOk(k); ok {
			t.Errorf("response-only key %q leaked into inputs: %v", k, item)
		}
	}
}

// roleDetailsSchemaJSON and roleDetailsWireJSON are the same permission
// descriptor tree in the two naming conventions: what a user writes in their
// program, and what Pulumi Cloud accepts. They are spelled out rather than
// derived from each other so the round-trip test proves the translation
// instead of restating it.
const roleDetailsSchemaJSON = `{
  "type__": "PermissionDescriptorGroup",
  "entries": [
    {
      "type__": "PermissionDescriptorCondition",
      "condition": {
        "type__": "PermissionExpressionEqual",
        "left": {"type__": "PermissionExpressionEnvironment"},
        "right": {"type__": "PermissionLiteralExpressionEnvironment", "identity": "acme/dev"}
      },
      "subNode": {"type__": "PermissionDescriptorAllow", "permissions": ["environment:read"]}
    }
  ]
}`

const roleDetailsWireJSON = `{
  "__type": "PermissionDescriptorGroup",
  "entries": [
    {
      "__type": "PermissionDescriptorCondition",
      "condition": {
        "__type": "PermissionExpressionEqual",
        "left": {"__type": "PermissionExpressionEnvironment"},
        "right": {"__type": "PermissionLiteralExpressionEnvironment", "identity": "acme/dev"}
      },
      "subNode": {"__type": "PermissionDescriptorAllow", "permissions": ["environment:read"]}
    }
  ]
}`

func mustUnmarshalMap(t *testing.T, s string) map[string]any {
	t.Helper()
	var out map[string]any
	if err := json.Unmarshal([]byte(s), &out); err != nil {
		t.Fatalf("unmarshal %q: %v", s, err)
	}
	return out
}

// TestDiscriminatorRoundTripsThroughWireNames drives a real resource whose
// input is a deeply nested discriminated tree and pins both boundary
// crossings: the request body must carry the wire's `__type` at every level
// (objects and array elements alike), and the state decoded from the response
// must come back in the schema's `type__` form, byte-identical to what the
// user wrote.
func TestDiscriminatorRoundTripsThroughWireNames(t *testing.T) {
	spec, meta := loadFixtures(t)
	r, ok := Resources(spec, meta)["pulumiservice:api:Role"]
	if !ok {
		t.Fatal("pulumiservice:api:Role not in metadata")
	}

	schemaTree := mustUnmarshalMap(t, roleDetailsSchemaJSON)
	wireTree := mustUnmarshalMap(t, roleDetailsWireJSON)

	var postBody []byte
	mock := &mockTransport{responseFn: func(req *http.Request) mockResponse {
		if req.Method == http.MethodPost {
			postBody, _ = io.ReadAll(req.Body)
		}
		// Pulumi Cloud echoes the descriptor back in the wire convention on
		// both the create response and the read-after-create.
		return mockResponse{
			status: 200,
			body:   `{"id":"role-1","name":"env-reader","details":` + roleDetailsWireJSON + `}`,
		}
	}}
	ctx := WithTransport(t.Context(), mock)

	inputs := propMap(map[string]any{
		orgNameKey: testOrgName,
		nameKey:    "env-reader",
		"details":  schemaTree,
	})
	resp, err := r.Create(ctx, p.CreateRequest{Properties: inputs})
	if err != nil {
		t.Fatalf("Create: %v\n  calls: %v", err, mock.calls)
	}

	sent := mustUnmarshalMap(t, string(postBody))
	if got := sent["details"]; !reflect.DeepEqual(got, wireTree) {
		t.Errorf("request body details:\n got %#v\nwant %#v", got, wireTree)
	}
	if strings.Contains(string(postBody), schemaType) {
		t.Errorf("request body leaked the schema-side name %q: %s", schemaType, postBody)
	}

	details, ok := resp.Properties.GetOk("details")
	if !ok {
		t.Fatalf("state has no details; have %v", resp.Properties)
	}
	if got := propertyValueToAny(details); !reflect.DeepEqual(got, schemaTree) {
		t.Errorf("state details:\n got %#v\nwant %#v", got, schemaTree)
	}
}
