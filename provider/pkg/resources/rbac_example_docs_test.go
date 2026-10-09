// Copyright 2026, Pulumi Corporation.
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

package resources

import (
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRbacExampleDocsMatchYamlExample keeps the schema's Example Usage in sync
// with the integration-tested YAML program it is generated from. On failure,
// run `make provider && scripts/gen-rbac-example-docs.sh`.
func TestRbacExampleDocsMatchYamlExample(t *testing.T) {
	t.Parallel()
	raw, err := os.ReadFile("../../../examples/yaml-rbac-roles/Pulumi.yaml")
	require.NoError(t, err)
	example := strings.ReplaceAll(string(raw), "-${digits}", "")
	example = regexp.MustCompile("(?s)# Everything below `resources:`.*?enforces it\\.\n").ReplaceAllString(example, "")

	m := regexp.MustCompile("(?s)```yaml\n(.*?)```").FindStringSubmatch(rbacExampleDocs)
	require.NotNil(t, m, "rbac_example.md has no yaml block")

	fromResources := func(s string) string { return strings.TrimSpace(s[strings.Index(s, "\nresources:"):]) }
	assert.Equal(t, fromResources(example), fromResources(m[1]),
		"docs/rbac_example.md is stale; run `make provider && scripts/gen-rbac-example-docs.sh`")

	for _, lang := range []string{"typescript", "python", "go", "csharp", "java"} {
		assert.Contains(t, rbacExampleDocs, "```"+lang+"\n")
	}
}
