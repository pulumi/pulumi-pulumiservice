// Copyright 2016-2026, Pulumi Corporation.
package cloud_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/pulumi/pulumi/sdk/v3/go/property"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/cloud"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/rest"
)

const (
	discKey       = "type__"
	detailsKey    = "details"
	apiRoleType   = "pulumiservice:api:Role"
	yamlDirName   = "yaml"
	permStackRead = "stack:read"
)

// TestValidatorAcceptsEveryYamlExample runs the Check-time input validator
// over every api resource in the yaml example corpus. A validator stricter
// than the corpus is a validator that breaks working programs, so any
// failure here is a validator bug or a broken example — both fail CI.
func TestValidatorAcceptsEveryYamlExample(t *testing.T) {
	spec, meta := cloud.Spec(), cloud.Metadata()
	byToken := resourceMetaByToken(meta)

	checked := 0
	walkYamlExamples(t, func(path string, doc map[string]any) {
		resources, _ := doc["resources"].(map[string]any)
		for resName, raw := range resources {
			res, _ := raw.(map[string]any)
			token, _ := res["type"].(string)
			rm, ok := byToken[token]
			if !ok || rm.Operations.Create == "" {
				continue
			}
			op, ok := spec.Op(rm.Operations.Create)
			if !ok {
				continue
			}
			props, _ := res["properties"].(map[string]any)
			inputs := map[string]property.Value{}
			for k, v := range props {
				inputs[k] = yamlValue(v)
			}
			failures := rest.ValidateInputs(spec, meta.Types, op, rm, property.NewMap(inputs))
			for _, f := range failures {
				example := filepath.Base(filepath.Dir(filepath.Dir(path)))
				t.Errorf("%s: resource %q (%s): %s: %s", example, resName, token, f.Property, f.Reason)
			}
			checked++
		}
	})
	if checked == 0 {
		t.Fatal("no api resources found in the yaml example corpus")
	}
}

// TestValidatorRejectsKnownBadShapes pins the failure modes the validator
// exists for, using shapes verified against the live service: an
// out-of-subset node the service 400s, a misspelled tag, and a misspelled
// field the service silently drops.
func TestValidatorRejectsKnownBadShapes(t *testing.T) {
	spec, meta := cloud.Spec(), cloud.Metadata()
	byToken := resourceMetaByToken(meta)

	cases := []struct {
		name   string
		token  string
		inputs map[string]any
		wantIn string
		wantAt string
	}{
		{
			name:  "out-of-subset literal in Condition.condition",
			token: apiRoleType,
			inputs: map[string]any{detailsKey: map[string]any{
				discKey:     "PermissionDescriptorCondition",
				"condition": map[string]any{discKey: "PermissionLiteralExpressionBool", "value": true},
				"subNode":   map[string]any{discKey: "PermissionDescriptorAllow", "permissions": []any{permStackRead}},
			}},
			wantIn: "expected one of: PermissionExpressionAnd, PermissionExpressionEqual, " +
				"PermissionExpressionHasTag, PermissionExpressionNot, PermissionExpressionOr",
			wantAt: "details.condition." + discKey,
		},
		{
			name:  "misspelled discriminator tag",
			token: apiRoleType,
			inputs: map[string]any{detailsKey: map[string]any{
				discKey: "PermissionDescriptorAllw", "permissions": []any{permStackRead},
			}},
			wantIn: `did you mean "PermissionDescriptorAllow"`,
			wantAt: "details." + discKey,
		},
		// Top-level field membership (the audit-log-export incident) is the
		// schema typecheck's domain: YAML programs fail typecheck and typed
		// SDKs cannot express unknown inputs. The validator owns the nested
		// positions the schema cannot reject at the top level.
		{
			name:  "misspelled field the service drops silently",
			token: apiRoleType,
			inputs: map[string]any{detailsKey: map[string]any{
				discKey: "PermissionDescriptorAllow", "permission": []any{permStackRead},
			}},
			wantIn: `did you mean "permissions"`,
			wantAt: "details.permission",
		},
		// The wire spelling is what Pulumi Cloud's REST docs show, so users
		// reach for it. The engine strips `__`-prefixed keys before Check
		// runs, so the tag would simply be absent: the validator must say the
		// discriminator is missing and name the form the user can actually
		// write, never echo `__type` back at them.
		{
			name:  "wire-form discriminator is not accepted",
			token: apiRoleType,
			inputs: map[string]any{detailsKey: map[string]any{
				"__type": "PermissionDescriptorAllow", "permissions": []any{permStackRead},
			}},
			wantIn: `missing "type__"`,
			wantAt: "details." + discKey,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rm, ok := byToken[tc.token]
			if !ok {
				// Token lookup is exact; fall back to a suffix match so the
				// fixture survives token spelling drift.
				for tok, m := range byToken {
					if strings.HasSuffix(tok, strings.TrimPrefix(tc.token, "pulumiservice:api:")) {
						rm, ok = m, true
						break
					}
				}
			}
			if !ok {
				t.Skipf("resource %s not in metadata", tc.token)
			}
			op, opOK := spec.Op(rm.Operations.Create)
			if !opOK {
				t.Fatalf("create op for %s not found", tc.token)
			}
			inputs := map[string]property.Value{}
			for k, v := range tc.inputs {
				inputs[k] = yamlValue(v)
			}
			failures := rest.ValidateInputs(spec, meta.Types, op, rm, property.NewMap(inputs))
			if len(failures) == 0 {
				t.Fatalf("expected failures, got none")
			}
			found := false
			for _, f := range failures {
				if strings.Contains(f.Reason, tc.wantIn) && (tc.wantAt == "" || f.Property == tc.wantAt) {
					found = true
				}
			}
			if !found {
				t.Errorf("no failure matched reason~%q property=%q; got %+v", tc.wantIn, tc.wantAt, failures)
			}
		})
	}
}

func resourceMetaByToken(meta *rest.Metadata) map[string]rest.ResourceMeta {
	out := map[string]rest.ResourceMeta{}
	for key, rm := range meta.Resources {
		token := key
		if rm.Token != "" {
			token = rm.Token
		}
		out[token] = rm
	}
	return out
}

// walkYamlExamples parses every yaml example program under examples/api and
// hands the decoded document to fn.
func walkYamlExamples(t *testing.T, fn func(path string, doc map[string]any)) {
	t.Helper()
	pkgRoot, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	examplesAPI := filepath.Join(pkgRoot, "..", "..", "..", "examples", "api")
	err = filepath.Walk(examplesAPI, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if info.IsDir() || !strings.HasSuffix(strings.ToLower(path), ".yaml") {
			return nil
		}
		if filepath.Base(filepath.Dir(path)) != yamlDirName {
			return nil
		}
		data, readErr := os.ReadFile(path) //nolint:gosec // G304: reading repo's own example YAML under test
		if readErr != nil {
			return readErr
		}
		var doc map[string]any
		if err := yaml.Unmarshal(data, &doc); err != nil {
			t.Errorf("%s: parse: %v", path, err)
			return nil
		}
		fn(path, doc)
		return nil
	})
	if err != nil {
		t.Fatalf("walk examples/api: %v", err)
	}
}

// yamlValue converts decoded YAML into a property.Value. Interpolations and
// fn:: invokes are unresolvable statically, so they become computed values,
// which the validator skips.
func yamlValue(x any) property.Value {
	switch v := x.(type) {
	case string:
		if strings.Contains(v, "${") {
			return property.New(property.Computed)
		}
		return property.New(v)
	case bool:
		return property.New(v)
	case int:
		return property.New(float64(v))
	case int64:
		return property.New(float64(v))
	case float64:
		return property.New(v)
	case map[string]any:
		for k := range v {
			if strings.HasPrefix(k, "fn::") {
				return property.New(property.Computed)
			}
		}
		out := map[string]property.Value{}
		for k, elem := range v {
			out[k] = yamlValue(elem)
		}
		return property.New(out)
	case []any:
		out := make([]property.Value, len(v))
		for i, elem := range v {
			out[i] = yamlValue(elem)
		}
		return property.New(out)
	default:
		return property.Value{}
	}
}
