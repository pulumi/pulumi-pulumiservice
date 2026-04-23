// Copyright 2016-2026, Pulumi Corporation.

package runtime

import (
	"testing"

	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/stretchr/testify/assert"
)

// Tests EvaluateChecks against the rule types v2 ships: requireOneOf,
// requireTogether, requireIf. Names chosen to mirror Webhook/Team shapes.

func TestEvaluateChecks_RequireOneOf(t *testing.T) {
	res := &CloudAPIResource{Checks: []CheckRule{{
		RequireOneOf: []string{"stackName", "environmentName", "organizationName"},
		Message:      "exactly one of stackName, environmentName, or organizationName must be set",
	}}}
	// Zero set → failure on all three.
	failures := EvaluateChecks(res, resource.PropertyMap{})
	assert.Len(t, failures, 3)
	// Exactly one set → no failures.
	failures = EvaluateChecks(res, resource.PropertyMap{
		"organizationName": resource.NewStringProperty("acme"),
	})
	assert.Empty(t, failures)
	// Two set → failures on all three (message is on the rule, not cleared).
	failures = EvaluateChecks(res, resource.PropertyMap{
		"organizationName": resource.NewStringProperty("acme"),
		"stackName":        resource.NewStringProperty("prod"),
	})
	assert.Len(t, failures, 3)
}

func TestEvaluateChecks_RequireTogether(t *testing.T) {
	res := &CloudAPIResource{Checks: []CheckRule{{
		RequireTogether: []string{"stackName", "projectName"},
	}}}
	// Both unset → pass.
	assert.Empty(t, EvaluateChecks(res, resource.PropertyMap{}))
	// Both set → pass.
	assert.Empty(t, EvaluateChecks(res, resource.PropertyMap{
		"stackName":   resource.NewStringProperty("prod"),
		"projectName": resource.NewStringProperty("webapp"),
	}))
	// Only one set → failure on the missing one.
	failures := EvaluateChecks(res, resource.PropertyMap{
		"stackName": resource.NewStringProperty("prod"),
	})
	assert.Len(t, failures, 1)
	assert.Equal(t, "projectName", failures[0].Property)
}

func TestEvaluateChecks_RequireIf(t *testing.T) {
	res := &CloudAPIResource{Checks: []CheckRule{{
		RequireIf: "teamType == pulumi",
		Field:     "name",
		Message:   "name is required for pulumi teams",
	}}}
	// Predicate false → no check.
	assert.Empty(t, EvaluateChecks(res, resource.PropertyMap{
		"teamType": resource.NewStringProperty("github"),
	}))
	// Predicate true, field present → pass.
	assert.Empty(t, EvaluateChecks(res, resource.PropertyMap{
		"teamType": resource.NewStringProperty("pulumi"),
		"name":     resource.NewStringProperty("platform"),
	}))
	// Predicate true, field missing → failure.
	failures := EvaluateChecks(res, resource.PropertyMap{
		"teamType": resource.NewStringProperty("pulumi"),
	})
	assert.Len(t, failures, 1)
	assert.Equal(t, "name", failures[0].Property)
}

func TestEvaluateChecks_QuotedPredicateLiteral(t *testing.T) {
	// Accept both bare (`pulumi`) and quoted (`"pulumi"`) rhs in predicates.
	res := &CloudAPIResource{Checks: []CheckRule{{
		RequireIf: `teamType == "pulumi"`,
		Field:     "name",
	}}}
	assert.Empty(t, EvaluateChecks(res, resource.PropertyMap{
		"teamType": resource.NewStringProperty("github"),
	}))
	assert.Len(t, EvaluateChecks(res, resource.PropertyMap{
		"teamType": resource.NewStringProperty("pulumi"),
	}), 1)
}

func TestEvaluateChecks_NoChecks(t *testing.T) {
	assert.Empty(t, EvaluateChecks(&CloudAPIResource{}, resource.PropertyMap{}))
	assert.Empty(t, EvaluateChecks(nil, resource.PropertyMap{}))
}
