// Copyright 2016-2026, Pulumi Corporation.
package cloud_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestScaffoldMetadataIdempotent re-runs the scaffolder against the embedded
// OpenAPI spec into a temp metadata.json and asserts the output is byte-equal
// to the committed metadata.json. Catches "someone bumped the spec but
// forgot to regenerate" and "scaffolder produces unstable output."
//
// Runs under -short despite shelling out to `go run ./tools/scaffold-metadata` —
// this is the most load-bearing invariant of the v2 layer, so it has to be on
// the default CI path. The shell-out completes in a few seconds on a modern
// dev box.
func TestScaffoldMetadataIdempotent(t *testing.T) {
	pkgRoot, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	repoRoot := filepath.Join(pkgRoot, "..", "..", "..")
	specPath := filepath.Join(pkgRoot, "spec.json")
	metaPath := filepath.Join(pkgRoot, "metadata.json")

	committed, err := os.ReadFile(metaPath)
	if err != nil {
		t.Fatalf("read metadata.json: %v", err)
	}

	tmpDir := t.TempDir()
	tmpMeta := filepath.Join(tmpDir, "metadata.json")
	if err := os.WriteFile(tmpMeta, committed, 0o600); err != nil {
		t.Fatalf("seed temp metadata: %v", err)
	}

	cmd := exec.Command("go", "run", "./provider/tools/scaffold-metadata",
		"-in", specPath,
		"-out", tmpMeta,
	)
	cmd.Dir = repoRoot
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("scaffold-metadata: %v\n%s", err, string(out))
	}

	after, err := os.ReadFile(tmpMeta)
	if err != nil {
		t.Fatalf("read regenerated metadata: %v", err)
	}

	if bytes.Equal(committed, after) {
		return
	}

	diffCmd := exec.Command("diff", "-u", metaPath, tmpMeta)
	diff, _ := diffCmd.CombinedOutput()
	t.Fatalf("metadata.json out of sync with scaffold-metadata.\n"+
		"Run `go generate ./provider/pkg/cloud/...` and commit the result.\n\n%s",
		string(diff))
}
