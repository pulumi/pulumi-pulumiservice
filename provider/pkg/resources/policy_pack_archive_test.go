// Copyright 2016-2026, Pulumi Corporation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package resources

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTarballDirectory_HashAndContents(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "PulumiPolicy.yaml"), []byte("runtime: nodejs\n"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "index.js"), []byte("// policy\n"), 0o600))

	got, hash, err := tarballDirectory(dir)
	require.NoError(t, err)
	require.NotEmpty(t, got)
	assert.Len(t, hash, 64) // sha256 hex

	names := readTarballEntries(t, got)
	assert.ElementsMatch(t, []string{"PulumiPolicy.yaml", "index.js"}, names)
}

func TestTarballDirectory_HashStableAcrossRuns(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "a.js"), []byte("a"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "b.js"), []byte("b"), 0o600))

	_, h1, err := tarballDirectory(dir)
	require.NoError(t, err)
	_, h2, err := tarballDirectory(dir)
	require.NoError(t, err)
	assert.Equal(t, h1, h2, "identical contents should hash identically")
}

func TestTarballDirectory_HashChangesWithContent(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "policy.js")
	require.NoError(t, os.WriteFile(file, []byte("v1"), 0o600))
	_, h1, err := tarballDirectory(dir)
	require.NoError(t, err)

	require.NoError(t, os.WriteFile(file, []byte("v2"), 0o600))
	_, h2, err := tarballDirectory(dir)
	require.NoError(t, err)
	assert.NotEqual(t, h1, h2)
}

func TestTarballDirectory_SkipsNodeModulesAndGitAndPulumi(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "keep.js"), []byte("ok"), 0o600))
	for _, skipDir := range []string{"node_modules", ".git", ".pulumi"} {
		require.NoError(t, os.MkdirAll(filepath.Join(dir, skipDir), 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(dir, skipDir, "ignored.txt"), []byte("x"), 0o600))
	}

	got, _, err := tarballDirectory(dir)
	require.NoError(t, err)

	names := readTarballEntries(t, got)
	assert.Equal(t, []string{"keep.js"}, names)
}

func TestTarballDirectory_RejectsNonDirectory(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "not-a-dir")
	require.NoError(t, os.WriteFile(file, []byte("x"), 0o600))

	_, _, err := tarballDirectory(file)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not a directory")
}

func TestTarballDirectory_RejectsMissingDirectory(t *testing.T) {
	_, _, err := tarballDirectory(filepath.Join(t.TempDir(), "does-not-exist"))
	require.Error(t, err)
}

func TestTarballDirectory_RejectsEmptyAfterSkips(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "node_modules"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "node_modules", "ignored.txt"), []byte("x"), 0o600))

	_, _, err := tarballDirectory(dir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "empty")
}

func TestTarballDirectory_SymlinkHashIncludesTarget(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlinks require elevated permissions on Windows")
	}
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "real.js"), []byte("policy"), 0o600))
	require.NoError(t, os.Symlink("real.js", filepath.Join(dir, "alias.js")))
	_, h1, err := tarballDirectory(dir)
	require.NoError(t, err)

	// Re-point the symlink at a different (also-existing) target — content of
	// the directory is identical except the link target. Hash should differ.
	require.NoError(t, os.WriteFile(filepath.Join(dir, "real2.js"), []byte("policy"), 0o600))
	require.NoError(t, os.Remove(filepath.Join(dir, "alias.js")))
	require.NoError(t, os.Symlink("real2.js", filepath.Join(dir, "alias.js")))
	_, h2, err := tarballDirectory(dir)
	require.NoError(t, err)
	assert.NotEqual(t, h1, h2)
}

func readTarballEntries(t *testing.T, data []byte) []string {
	t.Helper()
	gz, err := gzip.NewReader(bytes.NewReader(data))
	require.NoError(t, err)
	t.Cleanup(func() { _ = gz.Close() })
	tr := tar.NewReader(gz)
	var names []string
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
		names = append(names, hdr.Name)
	}
	return names
}
