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

// TestInferUpdateEnvelope pins the current<Stem>/new<Stem> detection rule:
// exactly two object-typed request properties with a shared stem. Anything
// looser would mis-tag ordinary two-field update bodies, anything stricter
// would miss the shape the flat body mapping can't express.
func TestInferUpdateEnvelope(t *testing.T) {
	const specJSON = `{
	  "openapi": "3.0.0",
	  "components": {"schemas": {
	    "CurrentTag": {"type": "object", "properties": {"value": {"type": "string"}}},
	    "NewTag":     {"type": "object", "properties": {"name": {"type": "string"}, "value": {"type": "string"}}},
	    "Envelope":       {"type": "object", "properties": {"currentTag": {"$ref": "#/components/schemas/CurrentTag"}, "newTag": {"$ref": "#/components/schemas/NewTag"}}},
	    "StemMismatch":   {"type": "object", "properties": {"currentTag": {"$ref": "#/components/schemas/CurrentTag"}, "newThing": {"$ref": "#/components/schemas/NewTag"}}},
	    "ThreeFields":    {"type": "object", "properties": {"currentTag": {"$ref": "#/components/schemas/CurrentTag"}, "newTag": {"$ref": "#/components/schemas/NewTag"}, "extra": {"type": "string"}}},
	    "ScalarWrappers": {"type": "object", "properties": {"currentTag": {"type": "string"}, "newTag": {"type": "string"}}}
	  }},
	  "paths": {
	    "/things/{id}":        {"patch": {"operationId": "UpdateEnvelopeThing", "parameters": [{"name": "id", "in": "path", "required": true, "schema": {"type": "string"}}], "requestBody": {"content": {"application/json": {"schema": {"$ref": "#/components/schemas/Envelope"}}}}, "responses": {"204": {}}}},
	    "/mismatch/{id}":      {"patch": {"operationId": "UpdateStemMismatch", "parameters": [{"name": "id", "in": "path", "required": true, "schema": {"type": "string"}}], "requestBody": {"content": {"application/json": {"schema": {"$ref": "#/components/schemas/StemMismatch"}}}}, "responses": {"204": {}}}},
	    "/three/{id}":         {"patch": {"operationId": "UpdateThreeFields", "parameters": [{"name": "id", "in": "path", "required": true, "schema": {"type": "string"}}], "requestBody": {"content": {"application/json": {"schema": {"$ref": "#/components/schemas/ThreeFields"}}}}, "responses": {"204": {}}}},
	    "/scalars/{id}":       {"patch": {"operationId": "UpdateScalarWrappers", "parameters": [{"name": "id", "in": "path", "required": true, "schema": {"type": "string"}}], "requestBody": {"content": {"application/json": {"schema": {"$ref": "#/components/schemas/ScalarWrappers"}}}}, "responses": {"204": {}}}}
	  }
	}`
	spec, err := rest.ParseSpec([]byte(specJSON))
	if err != nil {
		t.Fatalf("parse synthetic spec: %v", err)
	}

	cases := []struct {
		name string
		ops  derivedOps
		want *rest.UpdateEnvelopeMeta
	}{
		{"detected", derivedOps{Create: "C", Update: "UpdateEnvelopeThing"},
			&rest.UpdateEnvelopeMeta{CurrentField: "currentTag", NewField: "newTag"}},
		{"stem mismatch", derivedOps{Create: "C", Update: "UpdateStemMismatch"}, nil},
		{"three fields", derivedOps{Create: "C", Update: "UpdateThreeFields"}, nil},
		{"scalar wrappers", derivedOps{Create: "C", Update: "UpdateScalarWrappers"}, nil},
		{"no update op", derivedOps{Create: "C"}, nil},
		{"create==update (upsert)", derivedOps{Create: "UpdateEnvelopeThing", Update: "UpdateEnvelopeThing"}, nil},
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
