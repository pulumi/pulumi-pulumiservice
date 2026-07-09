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
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/urn"
	"github.com/pulumi/pulumi/sdk/v3/go/property"
)

const (
	modeKey           = "mode"
	accountsKey       = "accounts"
	policyGroupKey    = "policyGroup"
	grpVal            = "grp"
	stacksVal         = "stacks"
	devVal            = "dev"
	myProjectVal      = "my-project"
	acct1Val          = "acct-1"
	devProjectID      = "test-org/grp/dev/my-project"
	acct1ID           = "test-org/grp/acct-1"
	routingProjectKey = "routingProject"
)

const stackAttachmentToken = "pulumiservice:api:PolicyGroupStackAttachment" //nolint:gosec // type token

// pgPath is the shared parent path for UpdatePolicyGroup (PATCH) and
// GetPolicyGroup (GET) used across the attachment tests.
const pgPath = "/api/orgs/test-org/policygroups/grp"

// groupWithStacks renders a GetPolicyGroup response whose membership list
// contains the given stack references.
func groupWithStacks(stacks ...map[string]string) string {
	arr := make([]map[string]string, 0, len(stacks))
	arr = append(arr, stacks...)
	body, _ := json.Marshal(map[string]any{
		nameKey:              grpVal,
		"isOrgDefault":       false,
		"entityType":         stacksVal,
		modeKey:              "audit",
		"appliedPolicyPacks": []any{},
		accountsKey:          []any{},
		stacksVal:            arr,
	})
	return string(body)
}

func stackAttachmentResource(t *testing.T) *Resource {
	t.Helper()
	spec, meta := loadFixtures(t)
	r := Resources(spec, meta)[stackAttachmentToken]
	if r == nil {
		t.Fatalf("%s not in factory output", stackAttachmentToken)
	}
	if r.meta.Attachment == nil {
		t.Fatalf("%s metadata is missing the attachment descriptor", stackAttachmentToken)
	}
	return r
}

func stackAttachmentInputs() property.Map {
	return propMap(map[string]any{
		orgNameKey:        testOrgName,
		policyGroupKey:    grpVal,
		nameKey:           devVal,
		routingProjectKey: myProjectVal,
	})
}

// TestAttachmentCreate: Create PATCHes the parent with {addStack: {...}} then
// reads the membership back, synthesizing the composite ID.
func TestAttachmentCreate(t *testing.T) {
	r := stackAttachmentResource(t)

	var patchBody string
	mock := &mockTransport{responseFn: func(req *http.Request) mockResponse {
		if req.Method == http.MethodPatch {
			b, _ := io.ReadAll(req.Body)
			patchBody = string(b)
			return mockResponse{status: 204}
		}
		return mockResponse{status: 200, body: groupWithStacks(
			map[string]string{nameKey: devVal, routingProjectKey: myProjectVal},
		)}
	}}
	SetTransportResolver(func(_ context.Context) (Transport, error) { return mock, nil })

	resp, err := r.Create(t.Context(), p.CreateRequest{Properties: stackAttachmentInputs()})
	if err != nil {
		t.Fatalf("create: %v\n  calls: %v", err, mock.calls)
	}
	if want := devProjectID; resp.ID != want {
		t.Errorf("ID: got %q, want %q", resp.ID, want)
	}
	wantCalls := []string{"PATCH " + pgPath, "GET " + pgPath}
	if len(mock.calls) != 2 || mock.calls[0] != wantCalls[0] || mock.calls[1] != wantCalls[1] { //nolint:gosec // guarded
		t.Errorf("calls: got %v, want %v", mock.calls, wantCalls)
	}
	// The edge is nested under addStack — not flattened to top-level body fields.
	var sent map[string]any
	if err := json.Unmarshal([]byte(patchBody), &sent); err != nil {
		t.Fatalf("patch body not JSON: %q", patchBody)
	}
	add, ok := sent["addStack"].(map[string]any)
	if !ok {
		t.Fatalf("patch body missing addStack object: %v", sent)
	}
	if add[nameKey] != devVal || add[routingProjectKey] != myProjectVal { //nolint:goconst // test fixture
		t.Errorf("addStack edge: got %v, want {name:dev, routingProject:my-project}", add)
	}
	if _, leaked := sent["removeStack"]; leaked {
		t.Errorf("create body must not carry removeStack: %v", sent)
	}
	// State carries both parent path params and edge fields.
	for k, want := range map[string]string{
		orgNameKey: testOrgName, policyGroupKey: grpVal, nameKey: devVal, routingProjectKey: myProjectVal,
	} {
		v, ok := resp.Properties.GetOk(k)
		if !ok || v.AsString() != want {
			t.Errorf("state[%q]: got %q (ok=%v), want %q", k, v.AsString(), ok, want)
		}
	}
}

// TestAttachmentReadPresent: a membership hit returns the edge state and keeps
// the resource ID.
func TestAttachmentReadPresent(t *testing.T) {
	r := stackAttachmentResource(t)
	mock := &mockTransport{responses: map[string]mockResponse{
		"GET " + pgPath: {status: 200, body: groupWithStacks(
			map[string]string{nameKey: "other", routingProjectKey: "p"},
			map[string]string{nameKey: devVal, routingProjectKey: myProjectVal},
		)},
	}}
	SetTransportResolver(func(_ context.Context) (Transport, error) { return mock, nil })

	resp, err := r.Read(t.Context(), p.ReadRequest{
		ID:         devProjectID,
		Properties: stackAttachmentInputs(),
	})
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if resp.ID != devProjectID {
		t.Errorf("present edge should keep its ID, got %q", resp.ID)
	}
	if v, ok := resp.Properties.GetOk(routingProjectKey); !ok || v.AsString() != myProjectVal {
		t.Errorf("state missing routingProject: %v (ok=%v)", v, ok)
	}
}

// TestAttachmentReadAbsent: when the edge isn't in the membership list, Read
// returns an empty response (empty ID) so the engine drops it from state.
func TestAttachmentReadAbsent(t *testing.T) {
	r := stackAttachmentResource(t)
	mock := &mockTransport{responses: map[string]mockResponse{
		"GET " + pgPath: {status: 200, body: groupWithStacks(
			map[string]string{nameKey: "other", routingProjectKey: "p"},
		)},
	}}
	SetTransportResolver(func(_ context.Context) (Transport, error) { return mock, nil })

	resp, err := r.Read(t.Context(), p.ReadRequest{
		ID:         devProjectID,
		Properties: stackAttachmentInputs(),
	})
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if resp.ID != "" {
		t.Errorf("absent edge must return empty ID (got %q) so the engine deletes it", resp.ID)
	}
}

// TestAttachmentReadParentGone: a 404 on the parent means the edge is gone too.
func TestAttachmentReadParentGone(t *testing.T) {
	r := stackAttachmentResource(t)
	mock := &mockTransport{responses: map[string]mockResponse{
		"GET " + pgPath: {status: 404, body: `{"code":404,"message":"not found"}`},
	}}
	SetTransportResolver(func(_ context.Context) (Transport, error) { return mock, nil })

	resp, err := r.Read(t.Context(), p.ReadRequest{
		ID:         devProjectID,
		Properties: stackAttachmentInputs(),
	})
	if err != nil {
		t.Fatalf("read should swallow parent 404, got: %v", err)
	}
	if resp.ID != "" {
		t.Errorf("parent-gone must return empty ID, got %q", resp.ID)
	}
}

// TestAttachmentDelete: Delete PATCHes the parent with {removeStack: {...}}.
func TestAttachmentDelete(t *testing.T) {
	r := stackAttachmentResource(t)
	var body string
	mock := &mockTransport{responseFn: func(req *http.Request) mockResponse {
		b, _ := io.ReadAll(req.Body)
		body = string(b)
		return mockResponse{status: 204}
	}}
	SetTransportResolver(func(_ context.Context) (Transport, error) { return mock, nil })

	err := r.Delete(t.Context(), p.DeleteRequest{
		ID:         devProjectID,
		Properties: stackAttachmentInputs(),
	})
	if err != nil {
		t.Fatalf("delete: %v\n  calls: %v", err, mock.calls)
	}
	if len(mock.calls) != 1 || mock.calls[0] != "PATCH "+pgPath {
		t.Errorf("calls: got %v, want one PATCH", mock.calls)
	}
	var sent map[string]any
	if err := json.Unmarshal([]byte(body), &sent); err != nil {
		t.Fatalf("delete body not JSON: %q", body)
	}
	if _, ok := sent["removeStack"].(map[string]any); !ok {
		t.Errorf("delete body must nest the edge under removeStack: %v", sent)
	}
	if _, leaked := sent["addStack"]; leaked {
		t.Errorf("delete body must not carry addStack: %v", sent)
	}
}

// TestAttachmentDeleteIdempotentOn404: a 404 on the parent during delete is success.
func TestAttachmentDeleteIdempotentOn404(t *testing.T) {
	r := stackAttachmentResource(t)
	mock := &mockTransport{responses: map[string]mockResponse{
		"PATCH " + pgPath: {status: 404, body: `{"code":404}`},
	}}
	SetTransportResolver(func(_ context.Context) (Transport, error) { return mock, nil })

	err := r.Delete(t.Context(), p.DeleteRequest{
		ID:         devProjectID,
		Properties: stackAttachmentInputs(),
	})
	if err != nil {
		t.Errorf("404 on delete should be success, got: %v", err)
	}
}

// TestAttachmentDiffReplaces: any input change is a replace (edges have no
// in-place update).
func TestAttachmentDiffReplaces(t *testing.T) {
	r := stackAttachmentResource(t)
	resp, err := r.Diff(t.Context(), p.DiffRequest{
		ID:        devProjectID,
		OldInputs: stackAttachmentInputs(),
		Inputs: propMap(map[string]any{
			orgNameKey: testOrgName, policyGroupKey: grpVal,
			nameKey: devVal, routingProjectKey: "other-project",
		}),
	})
	if err != nil {
		t.Fatalf("diff: %v", err)
	}
	if !resp.HasChanges {
		t.Fatal("expected HasChanges")
	}
	d, ok := resp.DetailedDiff[routingProjectKey]
	if !ok {
		t.Fatalf("expected a diff on routingProject, got %v", resp.DetailedDiff)
	}
	if d.Kind != p.UpdateReplace {
		t.Errorf("routingProject diff kind: got %v, want UpdateReplace", d.Kind)
	}
}

// TestAttachmentNoDiffWhenEqual: identical inputs produce no diff.
func TestAttachmentNoDiffWhenEqual(t *testing.T) {
	r := stackAttachmentResource(t)
	resp, err := r.Diff(t.Context(), p.DiffRequest{
		ID:        devProjectID,
		OldInputs: stackAttachmentInputs(),
		Inputs:    stackAttachmentInputs(),
	})
	if err != nil {
		t.Fatalf("diff: %v", err)
	}
	if resp.HasChanges {
		t.Errorf("equal inputs should not diff: %v", resp.DetailedDiff)
	}
}

// TestAttachmentReadMatchesSecretAndDependentInput: a MatchKey input that is
// secret-wrapped or carries dependencies (e.g. wired from another resource's
// output) must still match its plain membership element — Value.Equals is
// secret/dependency-sensitive, so the match must compare values only.
func TestAttachmentReadMatchesSecretAndDependentInput(t *testing.T) {
	r := stackAttachmentResource(t)
	mock := &mockTransport{responses: map[string]mockResponse{
		"GET " + pgPath: {status: 200, body: groupWithStacks(
			map[string]string{nameKey: devVal, routingProjectKey: myProjectVal},
		)},
	}}
	SetTransportResolver(func(_ context.Context) (Transport, error) { return mock, nil })

	// Model a refresh: the secret/dependency-bearing values arrive in Inputs
	// (old inputs from state), where they survive into the membership compare —
	// unlike an import, where empty Inputs are reconstructed plain from the ID.
	inputs := property.NewMap(map[string]property.Value{
		orgNameKey:     property.New(testOrgName),
		policyGroupKey: property.New(grpVal),
		nameKey:        property.New(devVal).WithSecret(true),
		routingProjectKey: property.New(myProjectVal).
			WithDependencies([]urn.URN{"urn:pulumi:dev::proj::pulumiservice:api/stacks:Stack::s"}),
	})
	resp, err := r.Read(t.Context(), p.ReadRequest{
		ID:         devProjectID,
		Inputs:     inputs,
		Properties: inputs,
	})
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if resp.ID == "" {
		t.Fatal("secret/dependency-bearing inputs must still match their membership element")
	}
}

// TestAttachmentReadIgnoresExtraElementFields: a membership element richer than
// the schema (an extra server-only field) must not leak that field into state.
func TestAttachmentReadIgnoresExtraElementFields(t *testing.T) {
	r := stackAttachmentResource(t)
	mock := &mockTransport{responses: map[string]mockResponse{
		"GET " + pgPath: {status: 200, body: groupWithStacks(
			map[string]string{nameKey: devVal, routingProjectKey: myProjectVal, "serverOnly": "leak"},
		)},
	}}
	SetTransportResolver(func(_ context.Context) (Transport, error) { return mock, nil })

	resp, err := r.Read(t.Context(), p.ReadRequest{
		ID:         devProjectID,
		Properties: stackAttachmentInputs(),
	})
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if _, leaked := resp.Properties.GetOk("serverOnly"); leaked {
		t.Errorf("undeclared server field leaked into state: %v", resp.Properties)
	}
	if v, ok := resp.Properties.GetOk(routingProjectKey); !ok || v.AsString() != myProjectVal {
		t.Errorf("declared edge field missing from state: %v (ok=%v)", v, ok)
	}
}

// TestAttachmentCheckPassthrough: Check returns attachment inputs unchanged
// (no create-op normalization applies).
func TestAttachmentCheckPassthrough(t *testing.T) {
	r := stackAttachmentResource(t)
	in := stackAttachmentInputs()
	resp, err := r.Check(t.Context(), p.CheckRequest{Inputs: in})
	if err != nil {
		t.Fatalf("check: %v", err)
	}
	if !mapEqual(resp.Inputs, in) {
		t.Errorf("check must pass attachment inputs through unchanged: got %v", resp.Inputs)
	}
}

// TestAttachmentUpdateRejected: Update is unreachable for replace-only
// attachments and errors clearly if a caller bypasses Diff.
func TestAttachmentUpdateRejected(t *testing.T) {
	r := stackAttachmentResource(t)
	_, err := r.Update(t.Context(), p.UpdateRequest{
		ID:        devProjectID,
		OldInputs: stackAttachmentInputs(),
		Inputs:    stackAttachmentInputs(),
	})
	if err == nil {
		t.Fatal("Update must reject replace-only attachment resources")
	}
}

const accountAttachmentToken = "pulumiservice:api:PolicyGroupInsightsAccountAttachment" //nolint:gosec // type token

// groupWithAccounts renders a GetPolicyGroup response whose membership list is
// a list of scalar account names (the spec types `accounts` as []string).
func groupWithAccounts(accounts ...string) string {
	arr := make([]string, 0, len(accounts))
	arr = append(arr, accounts...)
	body, _ := json.Marshal(map[string]any{
		nameKey:              grpVal,
		"isOrgDefault":       false,
		"entityType":         accountsKey,
		modeKey:              "audit",
		"appliedPolicyPacks": []any{},
		stacksVal:            []any{},
		accountsKey:          arr,
	})
	return string(body)
}

func accountAttachmentResource(t *testing.T) *Resource {
	t.Helper()
	spec, meta := loadFixtures(t)
	r := Resources(spec, meta)[accountAttachmentToken]
	if r == nil {
		t.Fatalf("%s not in factory output", accountAttachmentToken)
	}
	if r.meta.Attachment == nil {
		t.Fatalf("%s metadata is missing the attachment descriptor", accountAttachmentToken)
	}
	return r
}

func accountAttachmentInputs() property.Map {
	return propMap(map[string]any{
		orgNameKey:     testOrgName,
		policyGroupKey: grpVal,
		nameKey:        acct1Val,
	})
}

// TestAccountAttachmentCreate: the curated account edge nests under
// addInsightsAccount as {name}, and read-back matches the scalar accounts list.
func TestAccountAttachmentCreate(t *testing.T) {
	r := accountAttachmentResource(t)
	var patchBody string
	mock := &mockTransport{responseFn: func(req *http.Request) mockResponse {
		if req.Method == http.MethodPatch {
			b, _ := io.ReadAll(req.Body)
			patchBody = string(b)
			return mockResponse{status: 204}
		}
		return mockResponse{status: 200, body: groupWithAccounts(acct1Val, "acct-2")}
	}}
	SetTransportResolver(func(_ context.Context) (Transport, error) { return mock, nil })

	resp, err := r.Create(t.Context(), p.CreateRequest{Properties: accountAttachmentInputs()})
	if err != nil {
		t.Fatalf("create: %v\n  calls: %v", err, mock.calls)
	}
	if want := acct1ID; resp.ID != want {
		t.Errorf("ID: got %q, want %q", resp.ID, want)
	}
	var sent map[string]any
	if err := json.Unmarshal([]byte(patchBody), &sent); err != nil {
		t.Fatalf("patch body not JSON: %q", patchBody)
	}
	add, ok := sent["addInsightsAccount"].(map[string]any)
	if !ok || add[nameKey] != acct1Val { //nolint:goconst // test fixture
		t.Errorf("addInsightsAccount edge: got %v, want {name: acct-1}", sent["addInsightsAccount"])
	}
	// Spec-only account fields must NOT leak into the body.
	for _, leaked := range []string{"id", "ownedBy", "provider"} {
		if _, bad := add[leaked]; bad {
			t.Errorf("edge body leaked spec-only field %q: %v", leaked, add)
		}
	}
}

// TestAccountAttachmentReadAbsent: scalar membership miss → empty ID (gone).
func TestAccountAttachmentReadAbsent(t *testing.T) {
	r := accountAttachmentResource(t)
	mock := &mockTransport{responses: map[string]mockResponse{
		"GET " + pgPath: {status: 200, body: groupWithAccounts("someone-else")},
	}}
	SetTransportResolver(func(_ context.Context) (Transport, error) { return mock, nil })

	resp, err := r.Read(t.Context(), p.ReadRequest{ID: acct1ID, Properties: accountAttachmentInputs()})
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if resp.ID != "" {
		t.Errorf("absent scalar member must return empty ID, got %q", resp.ID)
	}
}

// TestAccountAttachmentReadPresent: scalar membership hit → state retained.
func TestAccountAttachmentReadPresent(t *testing.T) {
	r := accountAttachmentResource(t)
	mock := &mockTransport{responses: map[string]mockResponse{
		"GET " + pgPath: {status: 200, body: groupWithAccounts(acct1Val)},
	}}
	SetTransportResolver(func(_ context.Context) (Transport, error) { return mock, nil })

	resp, err := r.Read(t.Context(), p.ReadRequest{ID: acct1ID, Properties: accountAttachmentInputs()})
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if resp.ID == "" {
		t.Fatal("present scalar member should keep its ID")
	}
	if v, ok := resp.Properties.GetOk(nameKey); !ok || v.AsString() != acct1Val {
		t.Errorf("state name: got %q (ok=%v), want acct-1", v.AsString(), ok)
	}
}

// TestAccountAttachmentDelete: Delete nests under removeInsightsAccount.
func TestAccountAttachmentDelete(t *testing.T) {
	r := accountAttachmentResource(t)
	var body string
	mock := &mockTransport{responseFn: func(req *http.Request) mockResponse {
		b, _ := io.ReadAll(req.Body)
		body = string(b)
		return mockResponse{status: 204}
	}}
	SetTransportResolver(func(_ context.Context) (Transport, error) { return mock, nil })

	if err := r.Delete(t.Context(), p.DeleteRequest{
		ID:         acct1ID,
		Properties: accountAttachmentInputs(),
	}); err != nil {
		t.Fatalf("delete: %v", err)
	}
	var sent map[string]any
	if err := json.Unmarshal([]byte(body), &sent); err != nil {
		t.Fatalf("delete body not JSON: %q", body)
	}
	if rm, ok := sent["removeInsightsAccount"].(map[string]any); !ok || rm[nameKey] != acct1Val {
		t.Errorf("delete body must nest under removeInsightsAccount {name}: %v", sent)
	}
}

// TestAttachmentSchema: the generated resource exposes the parent path params
// and the edge fields, all replace-on-change and required.
func TestAttachmentSchema(t *testing.T) {
	spec, meta := loadFixtures(t)
	pkg, err := BuildSchema(spec, meta, "pulumiservice")
	if err != nil {
		t.Fatalf("BuildSchema: %v", err)
	}
	rs, ok := pkg.Resources[stackAttachmentToken]
	if !ok {
		t.Fatalf("%s missing from package", stackAttachmentToken)
	}
	for _, name := range []string{orgNameKey, policyGroupKey, nameKey, routingProjectKey} {
		ps, ok := rs.InputProperties[name]
		if !ok {
			t.Errorf("input %q missing", name)
			continue
		}
		if !ps.ReplaceOnChanges {
			t.Errorf("input %q must be replaceOnChanges (edges have no update)", name)
		}
	}
	wantRequired := map[string]bool{orgNameKey: true, policyGroupKey: true, nameKey: true, routingProjectKey: true}
	for _, req := range rs.RequiredInputs {
		delete(wantRequired, req)
	}
	if len(wantRequired) != 0 {
		t.Errorf("missing required inputs: %v", wantRequired)
	}
}
