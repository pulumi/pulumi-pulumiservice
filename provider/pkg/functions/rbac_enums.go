// Copyright 2026, Pulumi Corporation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package functions

import (
	"sort"
	"strings"
	"unicode"

	"github.com/pulumi/pulumi-cloud-sdk/go/apitype"
	"github.com/pulumi/pulumi-go-provider/infer"
)

// RbacScope is a Pulumi Cloud RBAC permission scope, such as `stack:read`.
type RbacScope string

// Values lists every scope the pinned Cloud SDK knows, so the enum and
// rbacPermissionSlice's validation always agree.
func (RbacScope) Values() []infer.EnumValue[RbacScope] {
	var out []infer.EnumValue[RbacScope]
	for _, p := range apitype.RbacPermissionNoPermission.AllValues() {
		if p == apitype.RbacPermissionNoPermission {
			continue
		}
		out = append(out, infer.EnumValue[RbacScope]{
			Name:  pascalCase(string(p)),
			Value: RbacScope(p),
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Value < out[j].Value })
	return out
}

// RbacEntityType is the kind of entity a permission rule can target.
type RbacEntityType string

const (
	RbacEntityTypeEnvironment     RbacEntityType = "environment"
	RbacEntityTypeInsightsAccount RbacEntityType = "insights-account"
	RbacEntityTypeStack           RbacEntityType = "stack"
)

func (RbacEntityType) Values() []infer.EnumValue[RbacEntityType] {
	return []infer.EnumValue[RbacEntityType]{
		{Name: "Environment", Value: RbacEntityTypeEnvironment, Description: "ESC environments."},
		{Name: "InsightsAccount", Value: RbacEntityTypeInsightsAccount, Description: "Insights accounts."},
		{Name: "Stack", Value: RbacEntityTypeStack, Description: "Stacks."},
	}
}

// pascalCase turns a scope such as `insights_account:read` into `InsightsAccountRead`.
func pascalCase(s string) string {
	var b strings.Builder
	upper := true
	for _, r := range s {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			upper = true
			continue
		}
		if upper {
			r = unicode.ToUpper(r)
			upper = false
		}
		b.WriteRune(r)
	}
	return b.String()
}

func scopeStrings(scopes []RbacScope) []string {
	out := make([]string, len(scopes))
	for i, s := range scopes {
		out[i] = string(s)
	}
	return out
}
