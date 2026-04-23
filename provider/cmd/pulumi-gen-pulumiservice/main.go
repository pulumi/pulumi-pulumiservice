// Copyright 2016-2026, Pulumi Corporation.
//
// pulumi-gen-pulumiservice is the code generator for the Pulumi Service
// Provider v2. It reads the pinned Pulumi Cloud OpenAPI spec and a
// repo-local resource-map.yaml, and emits (1) a Pulumi schema.json,
// (2) a compact metadata.json embedded into the runtime provider,
// and (3) a coverage-report.md audit artifact.
//
// v2.0 bring-up order: the coverage subcommand lands first so we can see
// the state of the map against the spec before any code generation work.

package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/gen"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	subcmd := os.Args[1]
	args := os.Args[2:]

	switch subcmd {
	case "coverage":
		os.Exit(runCoverage(args))
	case "schema":
		os.Exit(runSchema(args))
	case "metadata":
		os.Exit(runMetadata(args))
	case "help", "-h", "--help":
		usage()
		os.Exit(0)
	default:
		fmt.Fprintf(os.Stderr, "unknown subcommand %q\n\n", subcmd)
		usage()
		os.Exit(2)
	}
}

func runMetadata(args []string) int {
	fs := flag.NewFlagSet("metadata", flag.ExitOnError)
	specPath := fs.String("spec", "provider/spec/openapi_public.json", "path to the pinned OpenAPI spec")
	mapPath := fs.String("map", "provider/resource-map.yaml", "path to the resource map")
	outPath := fs.String("out", "bin/metadata.json", "path to write the generated runtime metadata")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	raw, err := gen.EmitMetadata(*specPath, *mapPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "metadata: %v\n", err)
		return 1
	}
	if err := os.WriteFile(*outPath, raw, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "metadata: writing %s: %v\n", *outPath, err)
		return 1
	}
	fmt.Fprintf(os.Stderr, "Wrote %s (%d bytes)\n", *outPath, len(raw))
	return 0
}

func runSchema(args []string) int {
	fs := flag.NewFlagSet("schema", flag.ExitOnError)
	specPath := fs.String("spec", "provider/spec/openapi_public.json", "path to the pinned OpenAPI spec")
	mapPath := fs.String("map", "provider/resource-map.yaml", "path to the resource map")
	outPath := fs.String("out", "bin/schema.json", "path to write the generated Pulumi schema")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	raw, err := gen.EmitSchema(*specPath, *mapPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "schema: %v\n", err)
		return 1
	}
	if err := os.WriteFile(*outPath, raw, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "schema: writing %s: %v\n", *outPath, err)
		return 1
	}
	fmt.Fprintf(os.Stderr, "Wrote %s (%d bytes)\n", *outPath, len(raw))
	return 0
}

func usage() {
	fmt.Fprintf(os.Stderr, `pulumi-gen-pulumiservice — Pulumi Service Provider v2 generator

Usage:
  pulumi-gen-pulumiservice <subcommand> [flags]

Subcommands:
  coverage    Print a coverage report of operationIds in the spec vs.
              entries in the resource map. Exits non-zero if any
              operationId is unmapped (i.e. not in a module, exclusion,
              or a custom resource's consumes list).

  schema      Emit a Pulumi schema.json from the resource-map + OpenAPI spec.
              Consumed by 'pulumi package gen-sdk' to produce language SDKs.

  metadata    Emit the runtime metadata.json. Embedded in the provider binary
              and driven by the CRUD dispatcher at request time.

Use "pulumi-gen-pulumiservice <subcommand> -h" for per-subcommand help.
`)
}

func runCoverage(args []string) int {
	fs := flag.NewFlagSet("coverage", flag.ExitOnError)
	specPath := fs.String("spec", "provider/spec/openapi_public.json", "path to the pinned OpenAPI spec")
	mapPath := fs.String("map", "provider/resource-map.yaml", "path to the resource map")
	outPath := fs.String("out", "", "optional path to write the report (defaults to stdout)")
	strict := fs.Bool("strict", true, "exit non-zero if any operationId is unmapped")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	report, err := gen.CoverageReport(*specPath, *mapPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "coverage: %v\n", err)
		return 1
	}

	out := os.Stdout
	if *outPath != "" {
		f, err := os.Create(*outPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "coverage: creating %s: %v\n", *outPath, err)
			return 1
		}
		defer f.Close()
		out = f
	}
	if _, err := fmt.Fprint(out, report.Markdown()); err != nil {
		fmt.Fprintf(os.Stderr, "coverage: writing report: %v\n", err)
		return 1
	}

	if *strict && report.UnmappedCount > 0 {
		fmt.Fprintf(os.Stderr, "\ncoverage: %d unmapped operationId(s); failing strict mode\n", report.UnmappedCount)
		return 1
	}
	return 0
}
