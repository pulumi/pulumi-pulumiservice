// Copyright 2016-2026, Pulumi Corporation.
//
// coverage.go — the load-bearing check for sustainable 100% coverage.
//
// Every operationId in the OpenAPI spec must either:
//   (a) be claimed by a resource, function, method, or custom entry in the
//       resource map, or
//   (b) appear in the exclusions list with a reason.
//
// The coverage report enumerates unmapped, duplicate-claimed, and stale
// (claimed-but-not-in-spec) operations. CI runs this in strict mode; any
// unmapped operation fails the build until a human decides "expose, exclude,
// or extend the map schema."

package gen

import (
	"fmt"
	"os"
	"sort"
	"strings"
)

// Report summarizes the state of the map against the spec.
type Report struct {
	TotalOperations int
	MappedCount     int
	ExcludedCount   int
	UnmappedCount   int

	// Per-section lists, sorted for stable output.
	Mapped      []OperationClaim
	Unmapped    []SpecOperation
	Excluded    []OperationClaim
	Duplicates  []DuplicateClaim
	Stale       []OperationClaim // claimed but not present in spec
	TodoMarkers []string         // e.g. "TODO:CreateOidcIssuer" — flagged for visibility
}

// DuplicateClaim means two map entries both claim the same operationId.
// We want exactly one owner per operation.
type DuplicateClaim struct {
	OperationID string
	Claims      []OperationClaim
}

// CoverageReport builds the report from spec + resource-map files on disk.
// For byte-based input (e.g. embedded copies), see CoverageReportFromBytes.
// Does not mutate any input file.
func CoverageReport(specPath, mapPath string) (*Report, error) {
	specBytes, err := os.ReadFile(specPath)
	if err != nil {
		return nil, fmt.Errorf("reading spec %s: %w", specPath, err)
	}
	mapBytes, err := os.ReadFile(mapPath)
	if err != nil {
		return nil, fmt.Errorf("reading resource-map %s: %w", mapPath, err)
	}
	return CoverageReportFromBytes(specBytes, mapBytes)
}

// CoverageReportFromBytes is the path-free form of CoverageReport.
func CoverageReportFromBytes(specBytes, mapBytes []byte) (*Report, error) {
	spec, err := LoadSpecFromBytes(specBytes)
	if err != nil {
		return nil, err
	}
	rm, err := LoadResourceMapFromBytes(mapBytes)
	if err != nil {
		return nil, err
	}

	claims := rm.ExtractOperationIDs()

	// Index claims by operationId so we can detect duplicates and
	// claimed-but-missing entries.
	byID := make(map[string][]OperationClaim, len(claims))
	for _, c := range claims {
		byID[c.OperationID] = append(byID[c.OperationID], c)
	}

	rep := &Report{}
	rep.TotalOperations = len(spec.Operations)

	// Sweep the spec: for each operation, is it claimed?
	claimedThisPass := map[string]bool{}
	for _, op := range spec.Operations {
		owners, ok := byID[op.OperationID]
		if !ok {
			rep.Unmapped = append(rep.Unmapped, op)
			continue
		}
		claimedThisPass[op.OperationID] = true

		// Count once per operation (even if duplicated).
		// If all owners are exclusions, count as excluded; otherwise mapped.
		allExclusion := true
		for _, o := range owners {
			if o.Kind != ClaimExclusion {
				allExclusion = false
				break
			}
		}
		if allExclusion {
			rep.Excluded = append(rep.Excluded, owners[0])
			rep.ExcludedCount++
		} else {
			rep.Mapped = append(rep.Mapped, owners[0])
			rep.MappedCount++
		}

		if len(owners) > 1 {
			rep.Duplicates = append(rep.Duplicates, DuplicateClaim{
				OperationID: op.OperationID,
				Claims:      owners,
			})
		}
	}

	// Stale claims: entries in the map that reference an operationId the
	// spec no longer has. Could indicate a spec rename.
	for id, owners := range byID {
		if !claimedThisPass[id] {
			rep.Stale = append(rep.Stale, owners...)
		}
	}

	// TODO markers: scan the raw map file text for "TODO:" so reviewers
	// can see which mappings are incomplete without traversing the YAML.
	// (We're kind to future-us: the bring-up file deliberately contains
	// placeholders; this surfaces them cleanly.)
	rep.TodoMarkers = collectTodoMarkers(mapBytes)

	rep.UnmappedCount = len(rep.Unmapped)

	// Sort outputs for stability.
	sort.Slice(rep.Unmapped, func(i, j int) bool {
		return rep.Unmapped[i].OperationID < rep.Unmapped[j].OperationID
	})
	sort.Slice(rep.Mapped, func(i, j int) bool {
		return rep.Mapped[i].OperationID < rep.Mapped[j].OperationID
	})
	sort.Slice(rep.Excluded, func(i, j int) bool {
		return rep.Excluded[i].OperationID < rep.Excluded[j].OperationID
	})
	sort.Slice(rep.Stale, func(i, j int) bool {
		return rep.Stale[i].OperationID < rep.Stale[j].OperationID
	})
	sort.Slice(rep.Duplicates, func(i, j int) bool {
		return rep.Duplicates[i].OperationID < rep.Duplicates[j].OperationID
	})
	sort.Strings(rep.TodoMarkers)

	return rep, nil
}

// Markdown renders the report as a human-readable Markdown document suitable
// for CI artifact or terminal output.
func (r *Report) Markdown() string {
	var b strings.Builder
	fmt.Fprintln(&b, "# Pulumi Service Provider v2 — coverage report")
	fmt.Fprintln(&b)
	fmt.Fprintf(&b, "- Total operations in spec: **%d**\n", r.TotalOperations)
	fmt.Fprintf(&b, "- Mapped: **%d**\n", r.MappedCount)
	fmt.Fprintf(&b, "- Excluded: **%d**\n", r.ExcludedCount)
	fmt.Fprintf(&b, "- Unmapped: **%d**\n", r.UnmappedCount)
	if len(r.Stale) > 0 {
		fmt.Fprintf(&b, "- Stale (claimed but not in spec): **%d**\n", len(r.Stale))
	}
	if len(r.Duplicates) > 0 {
		fmt.Fprintf(&b, "- Duplicate claims: **%d**\n", len(r.Duplicates))
	}
	if len(r.TodoMarkers) > 0 {
		fmt.Fprintf(&b, "- TODO markers in map: **%d**\n", len(r.TodoMarkers))
	}
	fmt.Fprintln(&b)

	if r.UnmappedCount > 0 {
		fmt.Fprintln(&b, "## Unmapped operations")
		fmt.Fprintln(&b)
		fmt.Fprintln(&b, "Each of these needs a home in the map (resource, function, method, custom, or exclusion).")
		fmt.Fprintln(&b)
		fmt.Fprintln(&b, "| operationId | method | path |")
		fmt.Fprintln(&b, "|---|---|---|")
		for _, op := range r.Unmapped {
			fmt.Fprintf(&b, "| `%s` | %s | `%s` |\n", op.OperationID, op.Method, op.Path)
		}
		fmt.Fprintln(&b)
	}
	if len(r.Duplicates) > 0 {
		fmt.Fprintln(&b, "## Duplicate claims")
		fmt.Fprintln(&b)
		fmt.Fprintln(&b, "An operation is claimed in more than one place. Exactly one owner is expected.")
		fmt.Fprintln(&b)
		for _, d := range r.Duplicates {
			fmt.Fprintf(&b, "- `%s`\n", d.OperationID)
			for _, c := range d.Claims {
				fmt.Fprintf(&b, "  - %s (%s)\n", c.ClaimedBy, c.Kind)
			}
		}
		fmt.Fprintln(&b)
	}
	if len(r.Stale) > 0 {
		fmt.Fprintln(&b, "## Stale claims")
		fmt.Fprintln(&b)
		fmt.Fprintln(&b, "Map references an operationId that is not in the spec. May indicate a rename or a removed endpoint.")
		fmt.Fprintln(&b)
		for _, c := range r.Stale {
			fmt.Fprintf(&b, "- `%s` (%s)\n", c.OperationID, c.ClaimedBy)
		}
		fmt.Fprintln(&b)
	}
	if len(r.TodoMarkers) > 0 {
		fmt.Fprintln(&b, "## TODO markers")
		fmt.Fprintln(&b)
		fmt.Fprintln(&b, "Placeholders left in the map during bring-up. Each needs a real operationId.")
		fmt.Fprintln(&b)
		for _, t := range r.TodoMarkers {
			fmt.Fprintf(&b, "- %s\n", t)
		}
		fmt.Fprintln(&b)
	}
	if r.ExcludedCount > 0 {
		fmt.Fprintln(&b, "## Excluded operations")
		fmt.Fprintln(&b)
		fmt.Fprintln(&b, "Intentionally not exposed as Pulumi resources/functions.")
		fmt.Fprintln(&b)
		for _, c := range r.Excluded {
			fmt.Fprintf(&b, "- `%s` — %s\n", c.OperationID, c.Reason)
		}
		fmt.Fprintln(&b)
	}
	return b.String()
}

// collectTodoMarkers greps the resource-map bytes for TODO:... tokens.
// We keep this as a text scan (not a structural decode) because TODOs
// appear in value positions we otherwise don't decode.
func collectTodoMarkers(mapBytes []byte) []string {
	var out []string
	for _, line := range strings.Split(string(mapBytes), "\n") {
		i := strings.Index(line, "TODO:")
		if i < 0 {
			continue
		}
		// Skip TODO inside comments (`# TODO: ...`) — those are documentation
		// for humans, not placeholder operationId values. A TODO is a real
		// placeholder only when it appears in a value position.
		if hash := strings.Index(line, "#"); hash >= 0 && hash < i {
			continue
		}
		rest := line[i:]
		for _, stop := range []string{" ", "\t", "#", ","} {
			if j := strings.Index(rest, stop); j >= 0 {
				rest = rest[:j]
			}
		}
		out = append(out, rest)
	}
	return out
}
