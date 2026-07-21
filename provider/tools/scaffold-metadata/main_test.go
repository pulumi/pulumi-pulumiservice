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

package main

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/rest"
)

// TestMergeAttachmentReportsChange checks mergeAttachment's changed return: the
// scaffolder relies on it to count modified-but-existing attachments, so a spec
// change to a derived field isn't silently reported as zero changes.
func TestMergeAttachmentReportsChange(t *testing.T) {
	d := attachmentDerivation{ //nolint:gosec // G101: test fixture, not a real credential.
		Token:           "pulumiservice:api:FooBarAttachment",
		MutationOp:      "UpdateFoo",
		ReadOp:          "GetFoo",
		AddField:        "addBar",
		RemoveField:     "removeBar",
		MembershipField: "bars",
		MatchKey:        []string{nameFieldKey},
		IDFormat:        "{orgName}/{foo}/{name}",
	}

	enc, changed, err := mergeAttachment(nil, d)
	if err != nil {
		t.Fatalf("merge (new): %v", err)
	}
	if !changed {
		t.Error("a newly-created attachment block must report changed=true")
	}

	if _, changed, err := mergeAttachment(enc, d); err != nil {
		t.Fatalf("merge (idempotent): %v", err)
	} else if changed {
		t.Error("an identical re-merge must report changed=false")
	}

	d.MembershipField = "differentBars"
	if _, changed, err := mergeAttachment(enc, d); err != nil {
		t.Fatalf("merge (modified): %v", err)
	} else if !changed {
		t.Error("a changed membershipField must report changed=true")
	}
}

// TestMergeAttachmentPreservesPinnedIDFormat verifies a hand-pinned idFormat
// survives a re-merge even when the derived idFormat differs: the pin wins and
// the scaffolder logs the divergence rather than silently overwriting it.
func TestMergeAttachmentPreservesPinnedIDFormat(t *testing.T) {
	d := attachmentDerivation{ //nolint:gosec // G101: test fixture, not a real credential.
		Token:           "pulumiservice:api:FooBarAttachment",
		MutationOp:      "UpdateFoo",
		ReadOp:          "GetFoo",
		AddField:        "addBar",
		RemoveField:     "removeBar",
		MembershipField: "bars",
		MatchKey:        []string{nameFieldKey},
		IDFormat:        "{orgName}/{foo}/{name}",
	}
	existing := json.RawMessage(`{"idFormat":"{orgName}/{name}","token":"pulumiservice:api:FooBarAttachment"}`)

	encoded, _, err := mergeAttachment(existing, d)
	if err != nil {
		t.Fatalf("merge: %v", err)
	}
	var out map[string]any
	if err := json.Unmarshal(encoded, &out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got := out["idFormat"]; got != "{orgName}/{name}" {
		t.Errorf("pinned idFormat must be preserved over the derived value, got %v", got)
	}
}

const (
	updateEnvelopeThingOp = "UpdateEnvelopeThing"
	updateThingUnionOp    = "UpdateThingUnion"
	createThingOp         = "CreateThing"
	newTagField           = "newTag"
	tokenField            = "token"
	memberActionField     = "memberAction"
	newDisplayNameField   = "newDisplayName"
)

// envelopeSpec is the shared fixture for the envelope-detection and
// validation tests: one qualifying current/new body plus the near-miss
// shapes the rules must reject.
const envelopeSpec = `{
  "openapi": "3.0.0",
  "components": {"schemas": {
    "CurrentTag": {"type": "object", "properties": {"value": {"type": "string"}}},
    "NewTag": {"type": "object", "properties": {
      "name":  {"type": "string"},
      "value": {"type": "string"}
    }},
    "Envelope": {"type": "object", "properties": {
      "currentTag": {"$ref": "#/components/schemas/CurrentTag"},
      "newTag":     {"$ref": "#/components/schemas/NewTag"}
    }},
    "StemMismatch": {"type": "object", "properties": {
      "currentTag": {"$ref": "#/components/schemas/CurrentTag"},
      "newThing":   {"$ref": "#/components/schemas/NewTag"}
    }},
    "ThreeFields": {"type": "object", "properties": {
      "currentTag": {"$ref": "#/components/schemas/CurrentTag"},
      "newTag":     {"$ref": "#/components/schemas/NewTag"},
      "extra":      {"type": "string"}
    }},
    "ScalarWrappers": {"type": "object", "properties": {
      "currentTag": {"type": "string"},
      "newTag":     {"type": "string"}
    }},
    "NonceTag": {"type": "object", "properties": {
      "value": {"type": "string"},
      "nonce": {"type": "string"}
    }, "required": ["nonce"]},
    "NonceEnvelope": {"type": "object", "properties": {
      "currentTag": {"$ref": "#/components/schemas/CurrentTag"},
      "newTag":     {"$ref": "#/components/schemas/NonceTag"}
    }}
  }},
  "paths": {
    "/things/{id}": {"patch": {
      "operationId": "UpdateEnvelopeThing",
      "parameters": [{"name": "id", "in": "path", "required": true, "schema": {"type": "string"}}],
      "requestBody": {"content": {"application/json": {"schema": {"$ref": "#/components/schemas/Envelope"}}}},
      "responses": {"204": {}}
    }},
    "/wrapper-inputs": {"post": {
      "operationId": "CreateWithWrappers",
      "requestBody": {"content": {"application/json": {"schema": {"$ref": "#/components/schemas/Envelope"}}}},
      "responses": {"204": {}}
    }},
    "/mismatch/{id}": {"patch": {
      "operationId": "UpdateStemMismatch",
      "parameters": [{"name": "id", "in": "path", "required": true, "schema": {"type": "string"}}],
      "requestBody": {"content": {"application/json": {"schema": {"$ref": "#/components/schemas/StemMismatch"}}}},
      "responses": {"204": {}}
    }},
    "/three/{id}": {"patch": {
      "operationId": "UpdateThreeFields",
      "parameters": [{"name": "id", "in": "path", "required": true, "schema": {"type": "string"}}],
      "requestBody": {"content": {"application/json": {"schema": {"$ref": "#/components/schemas/ThreeFields"}}}},
      "responses": {"204": {}}
    }},
    "/nonce/{id}": {"patch": {
      "operationId": "UpdateNonceThing",
      "parameters": [{"name": "id", "in": "path", "required": true, "schema": {"type": "string"}}],
      "requestBody": {"content": {"application/json": {"schema": {"$ref": "#/components/schemas/NonceEnvelope"}}}},
      "responses": {"204": {}}
    }},
    "/scalars/{id}": {"patch": {
      "operationId": "UpdateScalarWrappers",
      "parameters": [{"name": "id", "in": "path", "required": true, "schema": {"type": "string"}}],
      "requestBody": {"content": {"application/json": {"schema": {"$ref": "#/components/schemas/ScalarWrappers"}}}},
      "responses": {"204": {}}
    }}
  }
}`

// TestInferUpdateEnvelope pins the current<Stem>/new<Stem> detection rule:
// exactly two object-typed request properties with a shared stem, neither of
// which is itself a create input. Anything looser would mis-tag ordinary
// update bodies, anything stricter would miss the shape the flat body
// mapping can't express.
func TestInferUpdateEnvelope(t *testing.T) {
	spec, err := rest.ParseSpec([]byte(envelopeSpec))
	if err != nil {
		t.Fatalf("parse synthetic spec: %v", err)
	}

	cases := []struct {
		name string
		ops  derivedOps
		want *rest.UpdateEnvelopeMeta
	}{
		{"detected", derivedOps{Create: "C", Update: updateEnvelopeThingOp},
			&rest.UpdateEnvelopeMeta{CurrentField: "currentTag", NewField: newTagField}},
		{"stem mismatch", derivedOps{Create: "C", Update: "UpdateStemMismatch"}, nil},
		{"three fields", derivedOps{Create: "C", Update: "UpdateThreeFields"}, nil},
		{"scalar wrappers", derivedOps{Create: "C", Update: "UpdateScalarWrappers"}, nil},
		{"no update op", derivedOps{Create: "C"}, nil},
		{"create==update (upsert)", derivedOps{Create: updateEnvelopeThingOp, Update: updateEnvelopeThingOp}, nil},
		{"wrapper fields are create inputs", derivedOps{Create: "CreateWithWrappers", Update: updateEnvelopeThingOp}, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := inferUpdateEnvelope(spec, tc.ops)
			switch {
			case got == nil && tc.want == nil:
			case got == nil || tc.want == nil || *got != *tc.want:
				t.Errorf("inferUpdateEnvelope = %+v, want %+v", got, tc.want)
			}
		})
	}
}

// TestValidateUpdateEnvelope pins the regen-time drift check that mirrors the
// runtime's buildEnvelopeBody strictness.
func TestValidateUpdateEnvelope(t *testing.T) {
	spec, err := rest.ParseSpec([]byte(envelopeSpec))
	if err != nil {
		t.Fatalf("parse synthetic spec: %v", err)
	}
	envelope := &rest.UpdateEnvelopeMeta{CurrentField: "currentTag", NewField: newTagField}

	cases := []struct {
		name    string
		ops     derivedOps
		env     *rest.UpdateEnvelopeMeta
		wantErr string
	}{
		{"valid", derivedOps{Create: "C", Update: updateEnvelopeThingOp}, envelope, ""},
		{"field gone from schema", derivedOps{Create: "C", Update: "UpdateStemMismatch"}, envelope,
			`"newTag" is missing`},
		{"sibling outside envelope", derivedOps{Create: "C", Update: "UpdateThreeFields"}, envelope,
			"outside the declared envelope"},
		{"identical fields", derivedOps{Create: "C", Update: updateEnvelopeThingOp},
			&rest.UpdateEnvelopeMeta{CurrentField: newTagField, NewField: newTagField}, "two distinct fields"},
		{"no update body", derivedOps{Create: "C"}, envelope, "no request body"},
		{"required wrapper prop unsourceable", derivedOps{Create: "C", Update: "UpdateNonceThing"}, envelope,
			`requires field "nonce"`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateUpdateEnvelope(spec, tc.ops, tc.env, nil, nil)
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("want nil error, got %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("want error containing %q, got %v", tc.wantErr, err)
			}
		})
	}
}

// TestUnmappedUpdateFields pins the diagnostic that catches update bodies the
// flat input→body mapping can't serve, and the population surface it checks
// against: create-body fields, path/query params, and read-response fields
// (which land in state, the runtime's bodySrc fallback).
func TestUnmappedUpdateFields(t *testing.T) {
	const specJSON = `{
	  "openapi": "3.0.0",
	  "components": {"schemas": {
	    "CreateThing": {"type": "object", "properties": {
	      "name":        {"type": "string"},
	      "displayName": {"type": "string"}
	    }},
	    "ActionUnion": {"type": "object", "properties": {
	      "newDisplayName": {"type": "string"},
	      "memberAction":   {"type": "string"},
	      "mode":           {"type": "string"},
	      "token":          {"type": "string"}
	    }, "required": ["token"]},
	    "ThingRead": {"type": "object", "properties": {"newDisplayName": {"type": "string"}}},
	    "FlatPatch": {"type": "object", "properties": {"displayName": {"type": "string"}}}
	  }},
	  "paths": {
	    "/things/{org}/{thingName}": {
	      "post": {
	        "operationId": "CreateThing",
	        "parameters": [
	          {"name": "org",       "in": "path",  "required": true, "schema": {"type": "string"}},
	          {"name": "thingName", "in": "path",  "required": true, "schema": {"type": "string"}},
	          {"name": "mode",      "in": "query", "required": false, "schema": {"type": "string"}}
	        ],
	        "requestBody": {"content": {"application/json": {"schema": {"$ref": "#/components/schemas/CreateThing"}}}},
	        "responses": {"204": {}}
	      },
	      "get": {
	        "operationId": "GetThing",
	        "parameters": [
	          {"name": "org",       "in": "path", "required": true, "schema": {"type": "string"}},
	          {"name": "thingName", "in": "path", "required": true, "schema": {"type": "string"}}
	        ],
	        "responses": {"200": {"content": {"application/json": {"schema": {"$ref": "#/components/schemas/ThingRead"}}}}}
	      },
	      "patch": {
	        "operationId": "UpdateThingUnion",
	        "parameters": [
	          {"name": "org",       "in": "path", "required": true, "schema": {"type": "string"}},
	          {"name": "thingName", "in": "path", "required": true, "schema": {"type": "string"}}
	        ],
	        "requestBody": {"content": {"application/json": {"schema": {"$ref": "#/components/schemas/ActionUnion"}}}},
	        "responses": {"204": {}}
	      }
	    },
	    "/flat/{org}": {"patch": {
	      "operationId": "UpdateThingFlat",
	      "parameters": [{"name": "org", "in": "path", "required": true, "schema": {"type": "string"}}],
	      "requestBody": {"content": {"application/json": {"schema": {"$ref": "#/components/schemas/FlatPatch"}}}},
	      "responses": {"204": {}}
	    }}
	  }
	}`
	spec, err := rest.ParseSpec([]byte(specJSON))
	if err != nil {
		t.Fatalf("parse synthetic spec: %v", err)
	}
	renames := map[string]string{"name": "thingName"}

	// Without a read op: "mode" is covered by the create query param, the
	// rest of the action-union body is unmapped, "token" required.
	req, opt := unmappedUpdateFields(spec,
		derivedOps{Create: createThingOp, Update: updateThingUnionOp}, renames, nil)
	if len(req) != 1 || req[0] != tokenField {
		t.Errorf("required = %v, want [token]", req)
	}
	if len(opt) != 2 || opt[0] != memberActionField || opt[1] != newDisplayNameField {
		t.Errorf("optional = %v, want [memberAction newDisplayName]", opt)
	}

	// With a read op: newDisplayName is echoed by the read response (state
	// fallback), leaving only memberAction unmapped-optional.
	req, opt = unmappedUpdateFields(spec,
		derivedOps{Create: createThingOp, Read: "GetThing", Update: updateThingUnionOp}, renames, nil)
	if len(req) != 1 || req[0] != tokenField {
		t.Errorf("with read: required = %v, want [token]", req)
	}
	if len(opt) != 1 || opt[0] != memberActionField {
		t.Errorf("with read: optional = %v, want [memberAction]", opt)
	}

	// outputsExclude re-hides a read-response field from the surface.
	req, opt = unmappedUpdateFields(spec,
		derivedOps{Create: createThingOp, Read: "GetThing", Update: updateThingUnionOp}, renames,
		[]string{newDisplayNameField})
	if len(opt) != 2 || opt[0] != memberActionField || opt[1] != newDisplayNameField {
		t.Errorf("with exclude: optional = %v, want [memberAction newDisplayName]", opt)
	}
	if len(req) != 1 || req[0] != tokenField {
		t.Errorf("with exclude: required = %v, want [token]", req)
	}

	// Flat update body whose fields all map to inputs: nothing to report.
	req, opt = unmappedUpdateFields(spec, derivedOps{Create: createThingOp, Update: "UpdateThingFlat"}, nil, nil)
	if len(req)+len(opt) != 0 {
		t.Errorf("flat body: required=%v optional=%v, want none", req, opt)
	}

	// Upsert-shaped (create == update) resources are exempt.
	req, opt = unmappedUpdateFields(spec, derivedOps{Create: updateThingUnionOp, Update: updateThingUnionOp}, nil, nil)
	if len(req)+len(opt) != 0 {
		t.Errorf("upsert: required=%v optional=%v, want none", req, opt)
	}
}
