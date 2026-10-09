// Copyright 2026, Pulumi Corporation.
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
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestOrganizationRoleExampleDocs keeps the schema's Example Usage in sync
// with the YAML program it is generated from, and checks that every language
// carries each entity rule's comment. On failure, run
// `make provider && scripts/gen-organization-role-example-docs.sh`.
func TestOrganizationRoleExampleDocs(t *testing.T) {
	t.Parallel()
	raw, err := os.ReadFile("docs/organization_role_example/Pulumi.yaml")
	require.NoError(t, err)
	source := string(raw)

	blocks := map[string]string{}
	fences := regexp.MustCompile("(?s)```(\\w+)\n(.*?)```")
	for _, m := range fences.FindAllStringSubmatch(organizationRoleExampleDocs, -1) {
		blocks[m[1]] = m[2]
	}
	require.Contains(t, blocks, "yaml", "organization_role_example.md has no yaml block")
	assert.Equal(t, strings.TrimSpace(source[strings.Index(source, "\nconfig:"):]), strings.TrimSpace(blocks["yaml"]),
		"docs/organization_role_example.md is stale; run "+
			"`make provider && scripts/gen-organization-role-example-docs.sh`")

	// Every rule's comment, e.g. "Read every stack in the organization". Rules
	// are the items of the `*Rules` lists, 10 spaces in.
	comments := regexp.MustCompile(`(?m)^ {10}# (.*)$`).FindAllStringSubmatch(source, -1)
	require.NotEmpty(t, comments)
	for _, lang := range []string{"typescript", "python", "go", "csharp", "java"} {
		require.Contains(t, blocks, lang)
		for _, c := range comments {
			assert.Containsf(t, blocks[lang], c[1], "%s example is missing a rule comment", lang)
		}
	}
	assert.NotContains(t, organizationRoleExampleDocs, "__type",
		"the example must use the helper functions, not raw descriptors")
}
