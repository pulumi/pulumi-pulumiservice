// Copyright 2016-2026, Pulumi Corporation.
package cloud_test

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/cloud"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/rest"
	pschema "github.com/pulumi/pulumi/pkg/v3/codegen/schema"
)

// variantWaivers lists union variants no example exercises, with the reason.
// A waived variant that gains example coverage fails the test until the
// waiver is removed, so this list only ever shrinks.
var variantWaivers = map[string]string{
	// PermissionDescriptor exotics: Compose references other descriptors by
	// name and Select/IfThenElse express org-specific routing; the
	// rbac-scoped example covers the load-bearing Allow/Group/Condition path.
	"__type=PermissionDescriptorCompose":    "references sibling descriptors by name; no self-contained example shape",
	"__type=PermissionDescriptorIfThenElse": "routing variant; Condition covers the gating path",
	"__type=PermissionDescriptorSelect":     "routing variant; Condition covers the gating path",

	// PermissionExpression algebra beyond what rbac-scoped exercises
	// (And/Not/Equal + environment context/literal are covered there).
	"__type=PermissionExpressionHasTag":                 "tag-based (ABAC) scoping; needs org tag fixtures",
	"__type=PermissionExpressionTag":                    "tag-based (ABAC) scoping; needs org tag fixtures",
	"__type=PermissionExpressionOr":                     "boolean algebra covered by And/Not in rbac-scoped",
	"__type=PermissionExpressionStack":                  "stack scoping mirrors the environment-scoped example",
	"__type=PermissionExpressionTeam":                   "team scoping mirrors the environment-scoped example",
	"__type=PermissionExpressionInsightsAccount":        "insights scoping mirrors the environment-scoped example",
	"__type=PermissionLiteralExpressionBool":            "literal for boolean algebra corners",
	"__type=PermissionLiteralExpressionNumber":          "literal for numeric comparisons; no numeric context today",
	"__type=PermissionLiteralExpressionString":          "literal for string comparisons; no string context today",
	"__type=PermissionLiteralExpressionStack":           "stack scoping mirrors the environment-scoped example",
	"__type=PermissionLiteralExpressionTeam":            "team scoping mirrors the environment-scoped example",
	"__type=PermissionLiteralExpressionInsightsAccount": "insights scoping mirrors the environment-scoped example",

	// AgentEntity/AgentUserEvent variants ride on api/agents:Task, whose
	// example exercises the message flow, not entity diffs. (`type=stack`
	// matches incidentally elsewhere in the corpus; the generic tag name
	// makes it unwaivable without false staleness.)
	"type=policy_issue": "agent entity diff shape; task example covers the resource",
	"type=pull_request": "agent entity diff shape; task example covers the resource",
	"type=repository":   "agent entity diff shape; task example covers the resource",

	// ApprovalRuleEligibility variants beyond the team_member entry that
	// platform-bootstrap uses.
	"eligibilityType=specific_user": "requires a real userLogin present in the test org",
	"eligibilityType=has_permission_on_target": "platform-bootstrap uses team_member; " +
		"permission-based eligibility needs role fixtures",

	// VCS providers beyond the github block deployment-settings carries:
	// wire-identical modulo the tag.
	"provider=azure_devops": "wire-identical to github modulo tag",
	"provider=bitbucket":    "wire-identical to github modulo tag",
	"provider=custom":       "wire-identical to github modulo tag",
	"provider=gitlab":       "wire-identical to github modulo tag",
}

// TestEveryUnionVariantHasExample walks every discriminator in the generated
// schema and requires each mapping tag to appear (as `tagProp: tag`) in the
// yaml example corpus, or to carry an explicit waiver above. This is the
// guard that would have caught the unwritable Condition.condition slot: a
// union variant nobody can express in an example is a union variant nobody
// has ever exercised.
func TestEveryUnionVariantHasExample(t *testing.T) {
	pkg, err := rest.BuildSchema(cloud.Spec(), cloud.Metadata(), "pulumiservice")
	if err != nil {
		t.Fatalf("BuildSchema: %v", err)
	}

	type site struct{ tagProp, tag string }
	found := map[site]bool{}
	var collect func(ts pschema.TypeSpec)
	collect = func(ts pschema.TypeSpec) {
		if ts.Discriminator != nil {
			tags := make([]string, 0, len(ts.Discriminator.Mapping))
			for tag := range ts.Discriminator.Mapping {
				tags = append(tags, tag)
			}
			for _, tag := range tags {
				found[site{ts.Discriminator.PropertyName, tag}] = true
			}
		}
		for _, m := range ts.OneOf {
			collect(m)
		}
		if ts.Items != nil {
			collect(*ts.Items)
		}
		if ts.AdditionalProperties != nil {
			collect(*ts.AdditionalProperties)
		}
	}
	for _, res := range pkg.Resources {
		for _, p := range res.InputProperties {
			collect(p.TypeSpec)
		}
		for _, p := range res.Properties {
			collect(p.TypeSpec)
		}
	}
	for _, ct := range pkg.Types {
		for _, p := range ct.Properties {
			collect(p.TypeSpec)
		}
	}
	if len(found) == 0 {
		t.Fatal("no discriminators found in generated schema; collection is broken")
	}

	corpus := yamlExampleCorpus(t)

	var missing, staleWaivers []string
	seenWaivers := map[string]bool{}
	for s := range found {
		key := fmt.Sprintf("%s=%s", s.tagProp, s.tag)
		re := regexp.MustCompile(fmt.Sprintf(`(?m)%s:\s*"?%s"?\s*$`,
			regexp.QuoteMeta(s.tagProp), regexp.QuoteMeta(s.tag)))
		covered := re.Match(corpus)
		if _, waived := variantWaivers[key]; waived {
			seenWaivers[key] = true
			if covered {
				staleWaivers = append(staleWaivers, key)
			}
			continue
		}
		if !covered {
			missing = append(missing, key)
		}
	}
	for key := range variantWaivers {
		if !seenWaivers[key] {
			staleWaivers = append(staleWaivers, key+" (variant no longer in schema)")
		}
	}
	sort.Strings(missing)
	sort.Strings(staleWaivers)

	if len(missing) > 0 {
		t.Errorf("union variants with no yaml example coverage (add an example or a documented waiver):\n  %s",
			strings.Join(missing, "\n  "))
	}
	if len(staleWaivers) > 0 {
		t.Errorf("stale waivers (covered by examples now, or gone from the schema) — remove them:\n  %s",
			strings.Join(staleWaivers, "\n  "))
	}
}

func yamlExampleCorpus(t *testing.T) []byte {
	t.Helper()
	pkgRoot, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	examplesAPI := filepath.Join(pkgRoot, "..", "..", "..", "examples", "api")
	var corpus []byte
	err = filepath.Walk(examplesAPI, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(strings.ToLower(path), ".yaml") {
			return nil
		}
		if filepath.Base(filepath.Dir(path)) != "yaml" {
			return nil
		}
		data, readErr := os.ReadFile(path) //nolint:gosec // G304: reading repo's own example YAML under test
		if readErr != nil {
			return readErr
		}
		corpus = append(corpus, data...)
		corpus = append(corpus, '\n')
		return nil
	})
	if err != nil {
		t.Fatalf("walk examples/api: %v", err)
	}
	return corpus
}
