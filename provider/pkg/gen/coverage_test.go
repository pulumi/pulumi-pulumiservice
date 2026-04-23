// Copyright 2016-2026, Pulumi Corporation.

package gen

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCoverageReport_Basic exercises the coverage gate end-to-end against a
// minimal hand-authored spec and map pair. Keeps the fixture small so
// future regressions are easy to localize.
func TestCoverageReport_Basic(t *testing.T) {
	dir := t.TempDir()
	spec := `{
		"paths": {
			"/api/things": {"post": {"operationId": "CreateThing"}, "get": {"operationId": "ListThings"}},
			"/api/things/{id}": {"get": {"operationId": "GetThing"}, "delete": {"operationId": "DeleteThing"}},
			"/api/internal/healthz": {"get": {"operationId": "InternalHealthz"}}
		}
	}`
	rmap := `modules:
  core:
    resources:
      Thing:
        operations:
          create: CreateThing
          read:   GetThing
          delete: DeleteThing
    functions:
      listThings: { operationId: ListThings }
exclusions:
  - { operationId: InternalHealthz, reason: "not user-facing" }
`
	specPath := filepath.Join(dir, "spec.json")
	mapPath := filepath.Join(dir, "map.yaml")
	require.NoError(t, os.WriteFile(specPath, []byte(spec), 0644))
	require.NoError(t, os.WriteFile(mapPath, []byte(rmap), 0644))

	rep, err := CoverageReport(specPath, mapPath)
	require.NoError(t, err)
	assert.Equal(t, 5, rep.TotalOperations)
	assert.Equal(t, 4, rep.MappedCount, "resource CRUD + function")
	assert.Equal(t, 1, rep.ExcludedCount)
	assert.Equal(t, 0, rep.UnmappedCount)
	assert.Empty(t, rep.Duplicates)
	assert.Empty(t, rep.Stale)
}

// TestCoverageReport_DetectsUnmapped validates the load-bearing failure case.
func TestCoverageReport_DetectsUnmapped(t *testing.T) {
	dir := t.TempDir()
	spec := `{"paths": {"/api/new": {"post": {"operationId": "CreateSomething"}}}}`
	rmap := `modules: {}
`
	specPath := filepath.Join(dir, "spec.json")
	mapPath := filepath.Join(dir, "map.yaml")
	require.NoError(t, os.WriteFile(specPath, []byte(spec), 0644))
	require.NoError(t, os.WriteFile(mapPath, []byte(rmap), 0644))

	rep, err := CoverageReport(specPath, mapPath)
	require.NoError(t, err)
	assert.Equal(t, 1, rep.UnmappedCount)
	assert.Contains(t, rep.Markdown(), "CreateSomething")
}

// TestCoverageReport_PolymorphicOperations ensures nested operation blocks
// (case/scopes, polymorphic creates) are walked correctly without claiming
// metadata keys like `case: scope` as operationIds.
func TestCoverageReport_PolymorphicOperations(t *testing.T) {
	dir := t.TempDir()
	spec := `{"paths": {
		"/api/org/hooks":      {"post": {"operationId": "CreateOrgHook"}},
		"/api/stack/hooks":    {"post": {"operationId": "CreateStackHook"}}
	}}`
	rmap := `modules:
  hooks:
    resources:
      Webhook:
        operations:
          case: scope
          scopes:
            org:
              create: CreateOrgHook
            stack:
              create: CreateStackHook
`
	specPath := filepath.Join(dir, "spec.json")
	mapPath := filepath.Join(dir, "map.yaml")
	require.NoError(t, os.WriteFile(specPath, []byte(spec), 0644))
	require.NoError(t, os.WriteFile(mapPath, []byte(rmap), 0644))

	rep, err := CoverageReport(specPath, mapPath)
	require.NoError(t, err)
	assert.Equal(t, 0, rep.UnmappedCount, rep.Markdown())
	// Neither `scope` nor `org`/`stack` should be claimed as operationIds.
	for _, s := range rep.Stale {
		assert.NotContains(t, []string{"scope", "org", "stack"}, s.OperationID,
			"metadata keys should not be claimed as operationIds")
	}
}

// TestCoverageReport_IgnoresTodoMarkers confirms TODO placeholders are not
// treated as claims (they'd otherwise show up as stale on every run).
func TestCoverageReport_IgnoresTodoMarkers(t *testing.T) {
	dir := t.TempDir()
	spec := `{"paths": {"/api/x": {"get": {"operationId": "GetX"}}}}`
	rmap := `modules:
  x:
    resources:
      X:
        operations:
          create: TODO:CreateX
          read:   GetX
`
	specPath := filepath.Join(dir, "spec.json")
	mapPath := filepath.Join(dir, "map.yaml")
	require.NoError(t, os.WriteFile(specPath, []byte(spec), 0644))
	require.NoError(t, os.WriteFile(mapPath, []byte(rmap), 0644))

	rep, err := CoverageReport(specPath, mapPath)
	require.NoError(t, err)
	assert.Empty(t, rep.Stale, "TODO markers must not appear as stale claims")
	assert.Contains(t, strings.Join(rep.TodoMarkers, ","), "TODO:CreateX",
		"TODO markers should be surfaced in their own section")
}
