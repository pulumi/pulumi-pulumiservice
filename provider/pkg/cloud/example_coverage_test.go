// Copyright 2016-2026, Pulumi Corporation.
package cloud_test

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/cloud"
)

// resourceExampleWaivers lists api tokens deliberately not covered by any
// yaml example, with the reason. Add an entry only when the resource genuinely
// can't be exercised end-to-end from a yaml program (function-shaped APIs,
// internal session resources, etc.). Each entry signals to a future
// maintainer that the gap is intentional rather than forgotten.
var resourceExampleWaivers = map[string]string{
	"pulumiservice:api/esc:OpenEnvironmentRequest": "Function-shaped resource: create/read/update only, no delete. " +
		"Models an internal ESC env-editing session, not a user-managed IaC resource.",
}

// TestEveryApiResourceHasExample asserts that every api token declared in
// cloud.Metadata() appears as a `type: <token>` declaration in at least one
// yaml example under examples/api/<name>/yaml/. Tokens may opt out via
// resourceExampleWaivers with a written reason.
//
// Why yaml-only: yaml is the canonical lane (in-process provider, no SDK
// build dependency); covering yaml is sufficient evidence the resource is
// exercised end-to-end somewhere.
func TestEveryApiResourceHasExample(t *testing.T) {
	md := cloud.Metadata()

	tokens := make([]string, 0, len(md.Resources))
	for k, rm := range md.Resources {
		tok := rm.Token
		if tok == "" {
			tok = k
		}
		tokens = append(tokens, tok)
	}

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
		if info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(strings.ToLower(path), ".yaml") {
			return nil
		}
		if filepath.Base(filepath.Dir(path)) != "yaml" {
			return nil
		}
		data, readErr := os.ReadFile(path)
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

	var missing []string
	for _, tok := range tokens {
		if _, waived := resourceExampleWaivers[tok]; waived {
			continue
		}
		// Match `type: <token>` as a YAML resource declaration. The
		// type-line anchor rules out comment mentions of the token
		// (e.g. "# Note: pulumiservice:api/... isn't covered here").
		re := regexp.MustCompile(`(?m)^\s+type:\s*["']?` + regexp.QuoteMeta(tok) + `["']?\s*$`)
		if !re.Match(corpus) {
			missing = append(missing, tok)
		}
	}

	if len(missing) > 0 {
		sort.Strings(missing)
		t.Fatalf(
			"api tokens with no yaml example coverage:\n  %s\n\n"+
				"Add a yaml example exercising the resource "+
				"(examples/api/<name>/yaml/Main.yaml) or add a waiver entry "+
				"to resourceExampleWaivers in this file with a clear reason.",
			strings.Join(missing, "\n  "))
	}

	if len(resourceExampleWaivers) > 0 {
		keys := make([]string, 0, len(resourceExampleWaivers))
		for k := range resourceExampleWaivers {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		t.Logf("explicitly waived (no yaml example): %d", len(keys))
		for _, k := range keys {
			t.Logf("  %s — %s", k, resourceExampleWaivers[k])
		}
	}
}
