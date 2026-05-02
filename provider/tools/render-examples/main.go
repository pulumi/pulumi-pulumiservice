// Copyright 2016-2026, Pulumi Corporation.
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

// Command render-examples expands canonical PCL example programs into
// runnable per-language projects via `pulumi convert`.
//
// Walks examples/v2/<example>/main.pp (or any *.pp in the example dir) and,
// for each target language, runs:
//
//	pulumi convert --from pcl --language <lang> --out <example>/<lang>
//
// The per-language output directories are wiped and re-emitted on every
// run so stale files don't linger when the canonical PCL changes. Failures
// for one language don't abort the run; the tool continues with the
// remaining languages and exits non-zero if anything failed overall.
//
// Usage:
//
//	go run ./provider/tools/render-examples           # walk examples/v2
//	go run ./provider/tools/render-examples -dir foo  # alternative root
//	go run ./provider/tools/render-examples -lang ts,py,yaml
package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

// defaultLanguages is the set of conversion targets attempted unless -lang
// overrides. Each value is the string `pulumi convert --language` accepts.
var defaultLanguages = []string{"typescript", "python", "go", "csharp", "java", "yaml"}

// pcl-source-language flag for pulumi convert.
const sourceLanguage = "pcl"

func main() {
	root := flag.String("dir", "examples/v2", "Root directory containing example subdirectories with .pp files")
	langs := flag.String("lang", strings.Join(defaultLanguages, ","), "Comma-separated languages to render")
	flag.Parse()

	languages := strings.Split(*langs, ",")
	for i, l := range languages {
		languages[i] = strings.TrimSpace(l)
	}

	pulumiBin, err := exec.LookPath("pulumi")
	if err != nil {
		fmt.Fprintln(os.Stderr, "render-examples: pulumi CLI not found in PATH")
		os.Exit(1)
	}

	examples, err := findExamples(*root)
	if err != nil {
		fmt.Fprintln(os.Stderr, "render-examples:", err)
		os.Exit(1)
	}
	if len(examples) == 0 {
		fmt.Fprintf(os.Stderr, "render-examples: no .pp files found under %s\n", *root)
		os.Exit(1)
	}

	type result struct {
		example, language string
		err               error
	}
	var results []result

	for _, exDir := range examples {
		fmt.Printf("== %s\n", exDir)
		for _, lang := range languages {
			outDir := filepath.Join(exDir, languageDirName(lang))
			err := convert(pulumiBin, exDir, outDir, lang)
			results = append(results, result{example: exDir, language: lang, err: err})
			status := "ok"
			if err != nil {
				status = "FAIL: " + err.Error()
			}
			fmt.Printf("   %-12s %s\n", lang, status)
		}
	}

	failures := 0
	for _, r := range results {
		if r.err != nil {
			failures++
		}
	}
	fmt.Printf("\nrendered %d examples × %d languages; %d failure(s)\n",
		len(examples), len(languages), failures)
	if failures > 0 {
		os.Exit(1)
	}
}

// findExamples returns the set of directories under root that contain at
// least one .pp file, sorted by path.
func findExamples(root string) ([]string, error) {
	dirs := map[string]struct{}{}
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Ext(path) == ".pp" {
			dirs[filepath.Dir(path)] = struct{}{}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(dirs))
	for d := range dirs {
		out = append(out, d)
	}
	sort.Strings(out)
	return out, nil
}

// languageDirName maps a `pulumi convert --language` value to the local
// subdirectory we write the output to. Keep the directory short and
// stable — these become parts of paths users will type.
func languageDirName(lang string) string {
	switch lang {
	case "typescript":
		return "typescript"
	case "python":
		return "python"
	case "go":
		return "go"
	case "csharp":
		return "csharp"
	case "java":
		return "java"
	case "yaml":
		return "yaml"
	default:
		return lang
	}
}

// convert wipes outDir and runs `pulumi convert --from pcl --language <lang>`
// against the .pp source(s) in srcDir.
//
// The .pp files are first staged into a temporary directory and convert is
// invoked from there. Without this indirection, `pulumi convert` would
// read the entire source directory (including outDir if it lives inside
// srcDir) and recurse infinitely while copying source files into output.
func convert(pulumiBin, srcDir, outDir, lang string) error {
	if err := os.RemoveAll(outDir); err != nil {
		return fmt.Errorf("clean %s: %w", outDir, err)
	}
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", outDir, err)
	}

	stage, err := os.MkdirTemp("", "pcl-convert-*")
	if err != nil {
		return fmt.Errorf("stage tmp dir: %w", err)
	}
	defer os.RemoveAll(stage)

	matches, err := filepath.Glob(filepath.Join(srcDir, "*.pp"))
	if err != nil {
		return fmt.Errorf("glob *.pp in %s: %w", srcDir, err)
	}
	for _, m := range matches {
		data, err := os.ReadFile(m)
		if err != nil {
			return fmt.Errorf("read %s: %w", m, err)
		}
		if err := os.WriteFile(filepath.Join(stage, filepath.Base(m)), data, 0o644); err != nil {
			return fmt.Errorf("stage %s: %w", m, err)
		}
	}

	absOut, err := filepath.Abs(outDir)
	if err != nil {
		return fmt.Errorf("absolute path %s: %w", outDir, err)
	}

	// Project name is derived from the example dir, not the staging temp
	// dir; otherwise C# emits e.g. `pcl-convert-1234.csproj` and other
	// languages pick up similarly meaningless names.
	projectName := filepath.Base(srcDir)

	// `--generate-only` skips dependency install / SDK linking. We just
	// want the rendered source files; users wire up SDK linking per
	// language on their own (yarn link, pip install -e, etc.). Without
	// this, conversion fails for any language whose published SDK doesn't
	// yet contain the v2 namespace.
	cmd := exec.Command(pulumiBin, "convert",
		"--from", sourceLanguage,
		"--language", lang,
		"--out", absOut,
		"--name", projectName,
		"--generate-only",
	)
	cmd.Dir = stage
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		// Don't leave a half-rendered output dir — the conversion failed,
		// so any partial files are misleading.
		_ = os.RemoveAll(outDir)
		return err
	}
	return nil
}
