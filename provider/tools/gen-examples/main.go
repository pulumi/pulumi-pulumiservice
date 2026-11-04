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

// Command gen-examples turns Pulumi YAML examples into the per-language
// "Example Usage" markdown that infer resources and functions embed in their
// descriptions.
//
// Each `<name>.yaml` in the source directory is a multi-document YAML file.
// Every document is one example: a Pulumi YAML program (`variables`,
// `resources`, `outputs`, ...) plus a `description` key. The description's
// first line becomes the example's heading and any further lines a paragraph
// under it. The command converts each program with `pulumi convert` and writes
// `<name>.md` to the destination directory, with one fence per language inside
// the docs `{{% examples %}}` shortcodes. SDK generation keeps only the fence
// for each SDK's language.
//
// `pulumi convert` loads the pulumiservice plugin to resolve types, so the
// provider binary must be on PATH. `make examples` arranges that.
//
// Usage: gen-examples <yaml source dir> <markdown destination dir>
package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/pulumi/pulumi/pkg/v3/codegen"
	"github.com/pulumi/pulumi/pkg/v3/codegen/schema"
)

// languages lists the `pulumi convert` targets, in the order the fences are
// written, with the file each conversion writes its program to.
var languages = []struct {
	name, file string
}{
	{"typescript", "index.ts"},
	{"python", "__main__.py"},
	{"go", "main.go"},
	{"csharp", "Program.cs"},
	{"java", "src/main/java/generated_program/App.java"},
}

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "Usage: %s <yaml source dir> <markdown destination dir>\n", os.Args[0])
		os.Exit(1)
	}
	if err := run(os.Args[1], os.Args[2]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(srcDir, dstDir string) error {
	pkg, err := loadSchema()
	if err != nil {
		return err
	}
	paths, err := filepath.Glob(filepath.Join(srcDir, "*.yaml"))
	if err != nil {
		return err
	}
	if len(paths) == 0 {
		return fmt.Errorf("no .yaml files in %q", srcDir)
	}
	for _, path := range paths {
		md, err := render(path, pkg)
		if err != nil {
			return fmt.Errorf("processing %q: %w", path, err)
		}
		out := filepath.Join(dstDir, strings.TrimSuffix(filepath.Base(path), ".yaml")+".md")
		//nolint:gosec // G703: out is built from build-time CLI arguments, not user input.
		if err := os.WriteFile(out, []byte(md), 0o600); err != nil {
			return err
		}
		fmt.Println("Wrote", out)
	}
	return nil
}

// render converts every example in one YAML file and returns the markdown.
func render(path string, pkg *schema.Package) (string, error) {
	//nolint:gosec // G703: path is built from build-time CLI arguments, not user input.
	f, err := os.Open(filepath.Clean(path))
	if err != nil {
		return "", err
	}
	defer f.Close()

	var b strings.Builder
	b.WriteString("{{% examples %}}\n## Example Usage\n")
	decoder := yaml.NewDecoder(f)
	for {
		var doc yaml.Node
		if err := decoder.Decode(&doc); errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return "", err
		}
		description, program, err := splitExample(&doc)
		if err != nil {
			return "", err
		}
		if err := writeExample(&b, description, program, pkg); err != nil {
			return "", fmt.Errorf("example %q: %w", description, err)
		}
	}
	b.WriteString("{{% /examples %}}\n")
	return b.String(), nil
}

// splitExample separates an example's `description` from its program,
// keeping the program's keys in source order.
func splitExample(doc *yaml.Node) (string, *yaml.Node, error) {
	if doc.Kind != yaml.DocumentNode || len(doc.Content) != 1 || doc.Content[0].Kind != yaml.MappingNode {
		return "", nil, errors.New("each example must be a YAML mapping")
	}
	m := doc.Content[0]
	program := &yaml.Node{Kind: yaml.MappingNode}
	description := ""
	for i := 0; i+1 < len(m.Content); i += 2 {
		if m.Content[i].Value == "description" {
			description = m.Content[i+1].Value
			continue
		}
		program.Content = append(program.Content, m.Content[i], m.Content[i+1])
	}
	if description == "" {
		return "", nil, errors.New("example is missing a `description`")
	}
	return description, program, nil
}

func writeExample(b *strings.Builder, description string, program *yaml.Node, pkg *schema.Package) error {
	programYAML, err := encodeYAML(program)
	if err != nil {
		return err
	}
	enums, err := enumValues(pkg, program)
	if err != nil {
		return err
	}

	dir, err := os.MkdirTemp("", "gen-examples-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)

	src := filepath.Join(dir, "src")
	if err := os.Mkdir(src, 0o700); err != nil {
		return err
	}
	project := "name: example\nruntime: yaml\n" + programYAML
	if err := os.WriteFile(filepath.Join(src, "Pulumi.yaml"), []byte(project), 0o600); err != nil {
		return err
	}

	heading, body, _ := strings.Cut(strings.TrimSpace(description), "\n")
	fmt.Fprintf(b, "\n{{%% example %%}}\n### %s\n", heading)
	if body = strings.TrimSpace(body); body != "" {
		fmt.Fprintf(b, "\n%s\n", body)
	}
	for _, lang := range languages {
		code, err := convert(src, filepath.Join(dir, lang.name), lang.name, lang.file)
		if err != nil {
			return err
		}
		if lang.name == "java" {
			if code, err = fixJavaEnums(code, enums); err != nil {
				return err
			}
		}
		fmt.Fprintf(b, "\n```%s\n%s\n```\n", lang.name, strings.TrimSpace(code))
	}
	fmt.Fprintf(b, "\n```yaml\n%s\n```\n", strings.TrimSpace(programYAML))
	b.WriteString("{{% /example %}}\n")
	return nil
}

func convert(src, out, language, file string) (string, error) {
	var stderr bytes.Buffer
	//nolint:gosec // Arguments come from this command, not user input.
	cmd := exec.Command("pulumi", "convert", "--language", language, "--out", out, "--generate-only")
	cmd.Dir = src
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("converting to %s: %w\n%s", language, err, stderr.String())
	}
	code, err := os.ReadFile(filepath.Join(out, file))
	if err != nil {
		return "", fmt.Errorf("reading %s output: %w", language, err)
	}
	return string(code), nil
}

func encodeYAML(n *yaml.Node) (string, error) {
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(n); err != nil {
		return "", err
	}
	if err := enc.Close(); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// enumValue is an enum-typed value in an example program, with the Java enum
// class and constant it stands for.
type enumValue struct {
	value, class, constant string
}

// loadSchema reads the schema from the provider binary on PATH rather than
// schema.json, which may not have been regenerated since the provider changed.
func loadSchema() (*schema.Package, error) {
	provider, err := exec.LookPath("pulumi-resource-pulumiservice")
	if err != nil {
		return nil, fmt.Errorf("the provider binary must be on PATH: %w", err)
	}
	var stderr bytes.Buffer
	//nolint:gosec // provider is the binary found on PATH, not user input.
	cmd := exec.Command("pulumi", "package", "get-schema", provider)
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("reading the provider schema: %w\n%s", err, stderr.String())
	}
	var spec schema.PackageSpec
	if err := json.Unmarshal(out, &spec); err != nil {
		return nil, fmt.Errorf("parsing the provider schema: %w", err)
	}
	pkg, err := schema.ImportSpec(spec, nil, schema.NewNullLoader(), schema.ValidationOptions{})
	if err != nil {
		return nil, fmt.Errorf("binding the provider schema: %w", err)
	}
	return pkg, nil
}

// enumValues lists the enum-typed values passed to `fn::invoke` calls in
// `variables` and to `resources`, in program order.
func enumValues(pkg *schema.Package, program *yaml.Node) ([]enumValue, error) {
	var out []enumValue
	for i := 0; i+1 < len(program.Content); i += 2 {
		section := program.Content[i+1]
		switch program.Content[i].Value {
		case "variables":
			for j := 1; j < len(section.Content); j += 2 {
				invoke := mapValue(section.Content[j], "fn::invoke")
				if invoke == nil {
					continue
				}
				fn, ok := pkg.GetFunction(schemaToken(mapValue(invoke, "function")))
				if !ok || fn.Inputs == nil {
					continue
				}
				if err := walkProperties(mapValue(invoke, "arguments"), fn.Inputs.Properties, &out); err != nil {
					return nil, err
				}
			}
		case "resources":
			for j := 1; j < len(section.Content); j += 2 {
				res, ok := pkg.GetResource(schemaToken(mapValue(section.Content[j], "type")))
				if !ok {
					continue
				}
				props := mapValue(section.Content[j], "properties")
				if err := walkProperties(props, res.InputProperties, &out); err != nil {
					return nil, err
				}
			}
		}
	}
	return out, nil
}

func walkProperties(n *yaml.Node, props []*schema.Property, out *[]enumValue) error {
	if n == nil || n.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i+1 < len(n.Content); i += 2 {
		k := slices.IndexFunc(props, func(p *schema.Property) bool { return p.Name == n.Content[i].Value })
		if k < 0 {
			continue
		}
		if err := walk(n.Content[i+1], props[k].Type, out); err != nil {
			return err
		}
	}
	return nil
}

func walk(n *yaml.Node, t schema.Type, out *[]enumValue) error {
	switch t := codegen.UnwrapType(t).(type) {
	case *schema.EnumType:
		if n.Kind != yaml.ScalarNode {
			return nil
		}
		module, class, _ := strings.Cut(strings.TrimPrefix(t.Token, "pulumiservice:"), ":")
		if module != "index" {
			return fmt.Errorf("enum %s: only index enums are supported", t.Token)
		}
		for _, e := range t.Elements {
			if fmt.Sprint(e.Value) == n.Value {
				*out = append(*out, enumValue{value: n.Value, class: class, constant: e.Name})
				return nil
			}
		}
		return fmt.Errorf("%q is not a %s value", n.Value, t.Token)
	case *schema.ObjectType:
		return walkProperties(n, t.Properties, out)
	case *schema.ArrayType:
		if n.Kind != yaml.SequenceNode {
			return nil
		}
		for _, item := range n.Content {
			if err := walk(item, t.ElementType, out); err != nil {
				return err
			}
		}
	case *schema.MapType:
		if n.Kind != yaml.MappingNode {
			return nil
		}
		for i := 1; i < len(n.Content); i += 2 {
			if err := walk(n.Content[i], t.ElementType, out); err != nil {
				return err
			}
		}
	}
	return nil
}

// mapValue returns the value under key in a YAML mapping, or nil.
func mapValue(n *yaml.Node, key string) *yaml.Node {
	if n == nil || n.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i+1 < len(n.Content); i += 2 {
		if n.Content[i].Value == key {
			return n.Content[i+1]
		}
	}
	return nil
}

// schemaToken expands a Pulumi YAML type token such as `pulumiservice:Team`,
// which may omit the `index` module, to its schema form.
func schemaToken(n *yaml.Node) string {
	if n == nil {
		return ""
	}
	if parts := strings.Split(n.Value, ":"); len(parts) == 2 {
		return parts[0] + ":index:" + parts[1]
	}
	return n.Value
}

// javaEnumValue matches a string literal, or an enum constant that the Java
// generator already wrote as a fully qualified reference.
var javaEnumValue = regexp.MustCompile(`"((?:[^"\\]|\\.)*)"|com\.pulumi\.pulumiservice\.enums\.(\w+)\.(\w+)`)

// fixJavaEnums rewrites enum values in a converted Java program into enum
// constants. `pulumi convert --language java` writes enum values inside lists
// as string literals, which don't compile against the Java SDK
// (https://github.com/pulumi/pulumi-java/issues/2422). Remove this once the
// Java language host bundled with the CLI includes the fix.
//
// Values are matched against enums in program order, since the same string can
// belong to more than one enum (`stack:read` is both an organization and a
// stack scope). Enums the generator already converted are shortened to match.
func fixJavaEnums(code string, enums []enumValue) (string, error) {
	if len(enums) == 0 {
		return code, nil
	}
	next := 0
	var classes []string
	code = javaEnumValue.ReplaceAllStringFunc(code, func(match string) string {
		if next == len(enums) {
			return match
		}
		e := enums[next]
		m := javaEnumValue.FindStringSubmatch(match)
		literal := strings.HasPrefix(match, `"`) && m[1] == e.value
		qualified := m[2] == e.class && m[3] == e.constant
		if !literal && !qualified {
			return match
		}
		next++
		if !slices.Contains(classes, e.class) {
			classes = append(classes, e.class)
		}
		return e.class + "." + e.constant
	})
	if next != len(enums) {
		return "", fmt.Errorf("java output: found %d of %d enum values; next missing is %q",
			next, len(enums), enums[next].value)
	}
	firstImport := strings.Index(code, "\nimport ")
	if firstImport < 0 {
		return "", errors.New("java output has no imports")
	}
	slices.Sort(classes)
	var imports strings.Builder
	for _, c := range classes {
		fmt.Fprintf(&imports, "import com.pulumi.pulumiservice.enums.%s;\n", c)
	}
	return code[:firstImport+1] + imports.String() + code[firstImport+1:], nil
}
