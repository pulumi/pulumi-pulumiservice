// Copyright 2016-2026, Pulumi Corporation.

//go:build hammerv2

// Hammering loop for the v2 e2e suite — runs the yaml lanes of every
// FullE2E case in v2Cases sequentially N times against the configured
// backend. Surfaces flakes that only show on repeat runs (sweeper gaps,
// resource-cleanup races, ID collisions, idempotency drift).
//
// Off by default; opt in with `-tags hammerv2`. Each iteration randomizes
// the per-case Config helpers, so resources don't collide across rounds.
//
// Run locally:
//
//	cd examples
//	go test -tags hammerv2 -v -count=1 -timeout 6h -run TestV2Hammer
//
// Override iteration count via PULUMI_V2_HAMMER_ITERATIONS (default 3).

package examples

import (
	"fmt"
	"os"
	"strconv"
	"testing"
)

func TestV2Hammer(t *testing.T) {
	n := 3
	if v := os.Getenv("PULUMI_V2_HAMMER_ITERATIONS"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			n = parsed
		}
	}

	sweepV2TestState(t)

	for i := 0; i < n; i++ {
		i := i
		// Iterations run sequentially so each sees the cleaned state from
		// the prior one. Cases inside an iteration parallelize, matching
		// TestV2's behavior.
		t.Run(fmt.Sprintf("iter-%d", i), func(t *testing.T) {
			for _, ex := range v2Cases {
				ex := ex
				t.Run(ex.Name, func(t *testing.T) {
					t.Parallel()
					if ex.SkipReason != "" {
						t.Skip(ex.SkipReason)
					}
					if reason := skipLangReason(t, "yaml"); reason != "" {
						t.Skip(reason)
					}
					if reason, ok := ex.SkipLang["yaml"]; ok {
						t.Skip(reason)
					}
					runV2Case(t, ex, "yaml")
				})
			}
		})
	}
}
