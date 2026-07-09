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
