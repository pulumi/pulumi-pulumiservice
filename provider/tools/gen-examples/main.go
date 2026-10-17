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
// `resources`, `outputs`, ...) plus a `description` key that becomes the
// example's heading. The command converts each program with `pulumi convert`
// and writes `<name>.md` to the destination directory, with one fence per
// language inside the docs `{{% examples %}}` shortcodes. SDK generation keeps
// only the fence for each SDK's language.
//
// `pulumi convert` loads the pulumiservice plugin to resolve types, so the
// provider binary must be on PATH. `make generate_examples` arranges that.
//
// Usage: gen-examples <yaml source dir> <markdown destination dir>
package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
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
	paths, err := filepath.Glob(filepath.Join(srcDir, "*.yaml"))
	if err != nil {
		return err
	}
	if len(paths) == 0 {
		return fmt.Errorf("no .yaml files in %q", srcDir)
	}
	for _, path := range paths {
		md, err := render(path)
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
func render(path string) (string, error) {
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
		if err := writeExample(&b, description, program); err != nil {
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

func writeExample(b *strings.Builder, description string, program *yaml.Node) error {
	programYAML, err := encodeYAML(program)
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

	fmt.Fprintf(b, "\n{{%% example %%}}\n### %s\n", description)
	for _, lang := range languages {
		code, err := convert(src, filepath.Join(dir, lang.name), lang.name, lang.file)
		if err != nil {
			return err
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
