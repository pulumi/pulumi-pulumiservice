// Copyright 2016-2026, Pulumi Corporation.
package cloud_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
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

// exampleCoverage is how TestApi runs an example's yaml program.
type exampleCoverage int

const (
	notRun exampleCoverage = iota
	previewOnly
	fullE2E
)

// TestEveryApiResourceHasExample asserts that every api token declared in
// cloud.Metadata() appears as a `type: <token>` declaration in the yaml program
// of an example that TestApi runs. Tokens may opt out via
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
	slices.Sort(tokens)

	pkgRoot, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	examplesDir := filepath.Join(pkgRoot, "..", "..", "..", "examples")
	examplesAPI := filepath.Join(examplesDir, "api")

	registered := parseAPICases(t, filepath.Join(examplesDir, "examples_api_test.go"))

	entries, err := os.ReadDir(examplesAPI)
	if err != nil {
		t.Fatalf("read examples/api: %v", err)
	}
	programs := map[string][]byte{}
	var unregistered []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if _, ok := registered[e.Name()]; !ok {
			unregistered = append(unregistered, e.Name())
			continue
		}
		programs[e.Name()] = readYAMLProgram(t, filepath.Join(examplesAPI, e.Name(), "yaml"))
	}
	if len(unregistered) > 0 {
		t.Errorf("examples/api directories that TestApi never runs:\n  %s\n\n"+
			"Register each one in apiCases in examples/examples_api_test.go.",
			strings.Join(unregistered, "\n  "))
	}

	var missing, preview []string
	full := 0
	for _, tok := range tokens {
		if _, waived := resourceExampleWaivers[tok]; waived {
			continue
		}
		// Match `type: <token>` as a YAML resource declaration. The
		// type-line anchor rules out comment mentions of the token
		// (e.g. "# Note: pulumiservice:api/... isn't covered here").
		re := regexp.MustCompile(`(?m)^\s+type:\s*["']?` + regexp.QuoteMeta(tok) + `["']?\s*$`)
		best := notRun
		for dir, program := range programs {
			if re.Match(program) {
				best = max(best, registered[dir])
			}
		}
		switch best {
		case notRun:
			missing = append(missing, tok)
		case previewOnly:
			preview = append(preview, tok)
		case fullE2E:
			full++
		}
	}

	if len(missing) > 0 {
		t.Errorf(
			"api tokens with no yaml example that TestApi runs:\n  %s\n\n"+
				"Add a yaml example exercising the resource "+
				"(examples/api/<name>/yaml/Main.yaml) and register it in apiCases, "+
				"or add a waiver entry to resourceExampleWaivers in this file with a clear reason.",
			strings.Join(missing, "\n  "))
	}

	t.Logf("api resources run with full create, refresh and destroy: %d of %d", full, len(tokens))
	t.Logf("api resources run with preview only: %d", len(preview))
	for _, tok := range preview {
		t.Logf("  %s", tok)
	}
	if len(resourceExampleWaivers) > 0 {
		keys := make([]string, 0, len(resourceExampleWaivers))
		for k := range resourceExampleWaivers {
			keys = append(keys, k)
		}
		slices.Sort(keys)
		t.Logf("explicitly waived (no yaml example): %d", len(keys))
		for _, k := range keys {
			t.Logf("  %s — %s", k, resourceExampleWaivers[k])
		}
	}
}

func readYAMLProgram(t *testing.T, dir string) []byte {
	t.Helper()
	files, err := filepath.Glob(filepath.Join(dir, "*.yaml"))
	if err != nil {
		t.Fatalf("glob %s: %v", dir, err)
	}
	var program []byte
	for _, f := range files {
		data, err := os.ReadFile(f) //nolint:gosec // G122: reading repo's own example YAML under test
		if err != nil {
			t.Fatalf("read %s: %v", f, err)
		}
		program = append(program, data...)
		program = append(program, '\n')
	}
	return program
}

// parseAPICases reads the apiCases table from the source of TestApi and returns
// how each example directory's yaml program runs. It parses the file because
// the table lives in a test file behind the `api` build tag in another package.
func parseAPICases(t *testing.T, path string) map[string]exampleCoverage {
	t.Helper()
	f, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
	var table *ast.CompositeLit
	ast.Inspect(f, func(n ast.Node) bool {
		spec, ok := n.(*ast.ValueSpec)
		if !ok || len(spec.Names) != 1 || spec.Names[0].Name != "apiCases" || len(spec.Values) != 1 {
			return true
		}
		table, _ = spec.Values[0].(*ast.CompositeLit)
		return false
	})
	if table == nil {
		t.Fatalf("no apiCases table in %s", path)
	}

	cases := map[string]exampleCoverage{}
	for _, elt := range table.Elts {
		lit, ok := elt.(*ast.CompositeLit)
		if !ok {
			t.Fatalf("apiCases entry is not a struct literal: %T", elt)
		}
		var name, dir string
		var fullE2EField, previewOnlyAll, skipped bool
		for _, kv := range lit.Elts {
			kv, ok := kv.(*ast.KeyValueExpr)
			if !ok {
				t.Fatalf("apiCases entry uses positional fields")
			}
			switch kv.Key.(*ast.Ident).Name {
			case "Name":
				name = stringLit(t, kv.Value)
			case "Dir":
				dir = stringLit(t, kv.Value)
			case "FullE2E":
				fullE2EField = isTrue(kv.Value)
			case "PreviewOnlyAll":
				previewOnlyAll = isTrue(kv.Value)
			case "SkipReason":
				skipped = true
			}
		}
		if dir == "" {
			dir = name
		}
		coverage := previewOnly
		switch {
		case skipped:
			coverage = notRun
		case fullE2EField && !previewOnlyAll:
			coverage = fullE2E
		}
		cases[dir] = max(cases[dir], coverage)
	}
	return cases
}

func stringLit(t *testing.T, e ast.Expr) string {
	t.Helper()
	lit, ok := e.(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		t.Fatalf("apiCases Name and Dir must be string literals")
	}
	s, err := strconv.Unquote(lit.Value)
	if err != nil {
		t.Fatalf("unquote %s: %v", lit.Value, err)
	}
	return s
}

func isTrue(e ast.Expr) bool {
	id, ok := e.(*ast.Ident)
	return ok && id.Name == "true"
}
