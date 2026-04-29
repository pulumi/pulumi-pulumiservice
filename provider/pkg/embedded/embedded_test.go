// Copyright 2016-2026, Pulumi Corporation.

package embedded

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/gen"
)

// TestEmbedded_SpecParses confirms the embedded OpenAPI spec is valid
// JSON and contains operations. A trivial smoke test, but it catches
// the case where the embed file ends up empty (e.g., a corrupted copy).
func TestEmbedded_SpecParses(t *testing.T) {
	require.NotEmpty(t, Spec(), "embedded spec must not be empty")

	spec, err := gen.LoadSpecFromBytes(Spec())
	require.NoError(t, err)
	assert.Greater(t, len(spec.Operations), 100,
		"embedded spec should expose hundreds of operations")
}

// TestEmbedded_ResourceMapParses confirms the embedded resource map is
// valid YAML and decodes into a populated ResourceMap. Catches both
// "empty embed" and "we accidentally introduced a YAML syntax error".
func TestEmbedded_ResourceMapParses(t *testing.T) {
	require.NotEmpty(t, ResourceMap(), "embedded resource map must not be empty")

	rm, err := gen.LoadResourceMapFromBytes(ResourceMap())
	require.NoError(t, err)
	assert.NotEmpty(t, rm.Modules, "resource map must declare at least one module")
}

// TestEmbedded_CoverageGate is the strict coverage check: every
// operationId in the embedded spec must be claimed by the embedded
// resource map (as a resource, function, method, or explicit
// exclusion). CI fails when a spec refresh introduces new operations
// that no human has triaged.
//
// Iwahbe's spec for v2 says GetSchema "errors on incomplete or
// incorrect mappings" — this is the test-time enforcement of that
// commitment; the runtime variant lives in provider.GetSchema.
func TestEmbedded_CoverageGate(t *testing.T) {
	report, err := gen.CoverageReportFromBytes(Spec(), ResourceMap())
	require.NoError(t, err)

	if report.UnmappedCount > 0 {
		t.Fatalf("coverage gate failed: %d unmapped operationId(s)\n\n%s",
			report.UnmappedCount, report.Markdown())
	}
}
