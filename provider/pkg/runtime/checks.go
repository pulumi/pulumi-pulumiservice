// Copyright 2016-2026, Pulumi Corporation.
//
// checks.go — runtime evaluation of the declarative Check rules carried in
// CloudAPIResource.Checks (requireOneOf, requireTogether, requireIf). These
// replace the ad-hoc cross-field validation a hand-written Check() RPC
// would otherwise have to implement.

package runtime

import (
	"fmt"
	"strings"

	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
)

// CheckFailure mirrors the Pulumi gRPC CheckFailure shape. The provider
// package's public Check wrapper converts these to the RPC response.
type CheckFailure struct {
	Property string
	Reason   string
}

// EvaluateChecks runs every rule on a resource's Checks list against the
// provided inputs and returns accumulated failures. Empty result = pass.
func EvaluateChecks(res *CloudAPIResource, inputs resource.PropertyMap) []CheckFailure {
	if res == nil || len(res.Checks) == 0 {
		return nil
	}
	var failures []CheckFailure
	for _, rule := range res.Checks {
		switch {
		case len(rule.RequireOneOf) > 0:
			failures = append(failures, evalRequireOneOf(rule, inputs)...)
		case len(rule.RequireTogether) > 0:
			failures = append(failures, evalRequireTogether(rule, inputs)...)
		case rule.RequireIf != "" && rule.Field != "":
			failures = append(failures, evalRequireIf(rule, inputs)...)
		}
	}
	return failures
}

// evalRequireOneOf checks that exactly one of the named fields is set.
func evalRequireOneOf(rule CheckRule, inputs resource.PropertyMap) []CheckFailure {
	setCount := 0
	for _, name := range rule.RequireOneOf {
		if isPropertySet(inputs, name) {
			setCount++
		}
	}
	if setCount == 1 {
		return nil
	}
	msg := rule.Message
	if msg == "" {
		msg = fmt.Sprintf("exactly one of %v must be set; %d were set", rule.RequireOneOf, setCount)
	}
	out := make([]CheckFailure, 0, len(rule.RequireOneOf))
	for _, name := range rule.RequireOneOf {
		out = append(out, CheckFailure{Property: name, Reason: msg})
	}
	return out
}

// evalRequireTogether checks that either all or none of the named fields
// are set — rejects partial presence.
func evalRequireTogether(rule CheckRule, inputs resource.PropertyMap) []CheckFailure {
	setCount := 0
	for _, name := range rule.RequireTogether {
		if isPropertySet(inputs, name) {
			setCount++
		}
	}
	if setCount == 0 || setCount == len(rule.RequireTogether) {
		return nil
	}
	msg := rule.Message
	if msg == "" {
		msg = fmt.Sprintf("fields %v must all be set together (or none of them)", rule.RequireTogether)
	}
	out := make([]CheckFailure, 0, len(rule.RequireTogether))
	for _, name := range rule.RequireTogether {
		if !isPropertySet(inputs, name) {
			out = append(out, CheckFailure{Property: name, Reason: msg})
		}
	}
	return out
}

// evalRequireIf evaluates simple `field == value` predicates. Extension to
// richer expressions (!=, &&, ||) comes when v1-parity stresses it.
func evalRequireIf(rule CheckRule, inputs resource.PropertyMap) []CheckFailure {
	if !evalSimplePredicate(rule.RequireIf, inputs) {
		return nil
	}
	if isPropertySet(inputs, rule.Field) {
		return nil
	}
	msg := rule.Message
	if msg == "" {
		msg = fmt.Sprintf("%s is required when %s", rule.Field, rule.RequireIf)
	}
	return []CheckFailure{{Property: rule.Field, Reason: msg}}
}

// isPropertySet reports whether a property exists and has a non-null,
// non-empty value.
func isPropertySet(inputs resource.PropertyMap, name string) bool {
	v, ok := inputs[resource.PropertyKey(name)]
	if !ok {
		return false
	}
	if v.IsSecret() {
		v = v.SecretValue().Element
	}
	if v.IsNull() {
		return false
	}
	if v.IsString() && v.StringValue() == "" {
		return false
	}
	return true
}

// evalSimplePredicate parses `<field> == <literal>` (literal bare or
// quoted) and reports whether it holds against the inputs. Only `==` on
// string-valued fields is supported — the full set of requireIf predicates
// in use today. Extend when the map grows a case that needs more.
func evalSimplePredicate(expr string, inputs resource.PropertyMap) bool {
	field, rhs, ok := splitPredicate(expr)
	if !ok {
		return false
	}
	v, present := inputs[resource.PropertyKey(field)]
	if !present {
		return false
	}
	if v.IsSecret() {
		v = v.SecretValue().Element
	}
	if !v.IsString() {
		return false
	}
	return v.StringValue() == rhs
}

// splitPredicate parses `field == rhs`. RHS may be bare or quoted;
// whitespace around either side is trimmed.
func splitPredicate(expr string) (field, rhs string, ok bool) {
	idx := strings.Index(expr, "==")
	if idx < 0 {
		return "", "", false
	}
	field = strings.TrimSpace(expr[:idx])
	rhs = strings.Trim(strings.TrimSpace(expr[idx+2:]), `"'`)
	if field == "" || rhs == "" {
		return "", "", false
	}
	return field, rhs, true
}
