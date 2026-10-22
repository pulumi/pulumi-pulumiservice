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

package util

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"unicode/utf8"

	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/pulumi/pulumi/sdk/v3/go/property"
)

// autoNameSuffixLen is the number of hex characters in a generated name's
// random suffix.
const autoNameSuffixLen = 7

// GenerateAutoName produces a name of the form `<logical-name>-<suffix>` from
// a resource URN and the engine-supplied random seed. The suffix is derived
// from the seed, so the same seed always yields the same name. When maxLen is
// positive, the logical name is truncated so the result fits. Falls back to
// crypto/rand when seed is empty.
func GenerateAutoName(urn string, seed []byte, maxLen int) string {
	parts := strings.Split(urn, "::")
	base := parts[len(parts)-1]
	if base == "" {
		base = "resource"
	}
	var suffix string
	if len(seed) > 0 {
		h := sha256.Sum256(seed)
		suffix = hex.EncodeToString(h[:4])[:autoNameSuffixLen]
	} else {
		buf := make([]byte, 4)
		_, _ = rand.Read(buf)
		suffix = hex.EncodeToString(buf)[:autoNameSuffixLen]
	}
	candidate := base + "-" + suffix
	if maxLen > 0 && len(candidate) > maxLen {
		budget := maxLen - autoNameSuffixLen - 1
		if budget < 1 {
			return truncateUTF8(candidate, maxLen)
		}
		candidate = truncateUTF8(base, budget) + "-" + suffix
	}
	return candidate
}

// truncateUTF8 cuts s, which must be longer than n bytes, to at most n bytes
// without splitting a multi-byte rune.
func truncateUTF8(s string, n int) string {
	for n > 0 && !utf8.RuneStart(s[n]) {
		n--
	}
	return s[:n]
}

type autoNameContextKey struct{}

type autoNameContext struct {
	urn  string
	seed []byte
}

// WithAutoNameContext wraps prov so that each Check call carries the request's
// URN and random seed in its context. `infer` drops both from the
// infer.CheckRequest it hands to a resource's Check, and AutoName needs them.
func WithAutoNameContext(prov p.Provider) p.Provider {
	inner := prov.Check
	if inner == nil {
		return prov
	}
	prov.Check = func(ctx context.Context, req p.CheckRequest) (p.CheckResponse, error) {
		ctx = context.WithValue(ctx, autoNameContextKey{}, autoNameContext{
			urn:  string(req.Urn),
			seed: req.RandomSeed,
		})
		return inner(ctx, req)
	}
	return prov
}

// AutoName fills the string input `field` in req.NewInputs for an `infer`
// resource's Check, and returns the resulting inputs. Call it before
// infer.DefaultCheck.
//
//   - If the program sets the field (including to an unknown value), it's left
//     alone.
//   - Otherwise, if the old inputs carry a name, it's reused, so an update
//     doesn't rename the resource.
//   - Otherwise a name is generated with GenerateAutoName from the URN and the
//     random seed that WithAutoNameContext put in ctx, capped at maxLen.
//
// When the engine decides to replace a resource, it calls Check again without
// old inputs, so the replacement gets a freshly generated name.
func AutoName(ctx context.Context, req infer.CheckRequest, field string, maxLen int) property.Map {
	if v, ok := req.NewInputs.GetOk(field); ok && !v.IsNull() {
		return req.NewInputs
	}
	name := ""
	if v, ok := req.OldInputs.GetOk(field); ok && v.IsString() && v.AsString() != "" {
		name = v.AsString()
	} else {
		urn, seed := req.Name, []byte(nil)
		if c, ok := ctx.Value(autoNameContextKey{}).(autoNameContext); ok {
			if c.urn != "" {
				urn = c.urn
			}
			seed = c.seed
		}
		name = GenerateAutoName(urn, seed, maxLen)
	}
	return req.NewInputs.Set(field, property.New(name))
}
