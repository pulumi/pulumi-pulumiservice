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
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
	presource "github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/property"
)

const (
	testAutoNameURN      = "urn:pulumi:dev::demo::pulumiservice:index:OrgAccessToken::ci-token"
	testAutoNameLogical  = "ci-token"
	testAutoNameField    = "name"
	testAutoNameExplicit = "explicit-name"
	testAutoNameMaxLen   = 40
)

var testAutoNameSeed = []byte("deterministic-seed")

func TestGenerateAutoName(t *testing.T) {
	t.Run("seeded names are deterministic", func(t *testing.T) {
		a := GenerateAutoName(testAutoNameURN, testAutoNameSeed, testAutoNameMaxLen)
		b := GenerateAutoName(testAutoNameURN, testAutoNameSeed, testAutoNameMaxLen)
		assert.Equal(t, a, b)
		assert.True(t, strings.HasPrefix(a, testAutoNameLogical+"-"), a)
		assert.Len(t, a, len(testAutoNameLogical)+1+autoNameSuffixLen)
	})

	t.Run("different seeds give different names", func(t *testing.T) {
		assert.NotEqual(t,
			GenerateAutoName(testAutoNameURN, testAutoNameSeed, testAutoNameMaxLen),
			GenerateAutoName(testAutoNameURN, []byte("another-seed"), testAutoNameMaxLen))
	})

	t.Run("no seed falls back to a random suffix", func(t *testing.T) {
		name := GenerateAutoName(testAutoNameURN, nil, testAutoNameMaxLen)
		assert.True(t, strings.HasPrefix(name, testAutoNameLogical+"-"), name)
		assert.Len(t, name, len(testAutoNameLogical)+1+autoNameSuffixLen)
	})

	t.Run("long logical names are truncated to maxLen", func(t *testing.T) {
		urn := "urn:pulumi:dev::demo::pulumiservice:index:OrgAccessToken::" + strings.Repeat("x", 60)
		name := GenerateAutoName(urn, testAutoNameSeed, testAutoNameMaxLen)
		assert.Len(t, name, testAutoNameMaxLen)
		assert.True(t, strings.HasPrefix(name, strings.Repeat("x", testAutoNameMaxLen-autoNameSuffixLen-1)+"-"), name)
	})

	t.Run("truncation keeps multi-byte runes whole", func(t *testing.T) {
		urn := "urn:pulumi:dev::demo::pulumiservice:index:OrgAccessToken::" + strings.Repeat("é", 30)
		// An odd byte budget lands in the middle of a two-byte rune.
		for _, maxLen := range []int{testAutoNameMaxLen, testAutoNameMaxLen + 1} {
			name := GenerateAutoName(urn, testAutoNameSeed, maxLen)
			assert.LessOrEqual(t, len(name), maxLen)
			assert.True(t, utf8.ValidString(name), name)
		}
	})

	t.Run("maxLen too small for a suffix", func(t *testing.T) {
		assert.Len(t, GenerateAutoName(testAutoNameURN, testAutoNameSeed, 5), 5)
	})

	t.Run("no maxLen", func(t *testing.T) {
		urn := "urn:pulumi:dev::demo::pulumiservice:index:OrgAccessToken::" + strings.Repeat("x", 60)
		assert.Len(t, GenerateAutoName(urn, testAutoNameSeed, 0), 60+1+autoNameSuffixLen)
	})

	t.Run("empty logical name", func(t *testing.T) {
		assert.True(t, strings.HasPrefix(GenerateAutoName("", testAutoNameSeed, 0), "resource-"))
	})
}

func autoNameCtx(t *testing.T) context.Context {
	t.Helper()
	var got context.Context
	prov := WithAutoNameContext(p.Provider{
		Check: func(ctx context.Context, _ p.CheckRequest) (p.CheckResponse, error) {
			got = ctx
			return p.CheckResponse{}, nil
		},
	})
	_, err := prov.Check(t.Context(), p.CheckRequest{
		Urn:        presource.URN(testAutoNameURN),
		RandomSeed: testAutoNameSeed,
	})
	require.NoError(t, err)
	return got
}

func TestWithAutoNameContextWithoutCheck(t *testing.T) {
	assert.Nil(t, WithAutoNameContext(p.Provider{}).Check)
}

func TestAutoName(t *testing.T) {
	ctx := autoNameCtx(t)
	want := GenerateAutoName(testAutoNameURN, testAutoNameSeed, testAutoNameMaxLen)
	nameOf := func(m property.Map) property.Value { return m.Get(testAutoNameField) }

	t.Run("explicit name is left alone", func(t *testing.T) {
		got := AutoName(ctx, infer.CheckRequest{
			Name:      testAutoNameLogical,
			NewInputs: property.NewMap(map[string]property.Value{testAutoNameField: property.New(testAutoNameExplicit)}),
			OldInputs: property.NewMap(map[string]property.Value{testAutoNameField: property.New("old-name")}),
		}, testAutoNameField, testAutoNameMaxLen)
		assert.Equal(t, testAutoNameExplicit, nameOf(got).AsString())
	})

	t.Run("unknown name is left alone", func(t *testing.T) {
		got := AutoName(ctx, infer.CheckRequest{
			Name:      testAutoNameLogical,
			NewInputs: property.NewMap(map[string]property.Value{testAutoNameField: property.New(property.Computed)}),
		}, testAutoNameField, testAutoNameMaxLen)
		assert.True(t, nameOf(got).IsComputed())
	})

	t.Run("old name is reused", func(t *testing.T) {
		got := AutoName(ctx, infer.CheckRequest{
			Name:      testAutoNameLogical,
			NewInputs: property.NewMap(nil),
			OldInputs: property.NewMap(map[string]property.Value{testAutoNameField: property.New("ci-token-1234567")}),
		}, testAutoNameField, testAutoNameMaxLen)
		assert.Equal(t, "ci-token-1234567", nameOf(got).AsString())
	})

	t.Run("name is generated from the URN and seed", func(t *testing.T) {
		got := AutoName(ctx, infer.CheckRequest{
			Name:      testAutoNameLogical,
			NewInputs: property.NewMap(map[string]property.Value{"other": property.New("kept")}),
		}, testAutoNameField, testAutoNameMaxLen)
		assert.Equal(t, want, nameOf(got).AsString())
		assert.Equal(t, "kept", got.Get("other").AsString())
	})

	t.Run("null name is generated", func(t *testing.T) {
		got := AutoName(ctx, infer.CheckRequest{
			Name:      testAutoNameLogical,
			NewInputs: property.NewMap(map[string]property.Value{testAutoNameField: property.New(property.Null)}),
		}, testAutoNameField, testAutoNameMaxLen)
		assert.Equal(t, want, nameOf(got).AsString())
	})

	t.Run("without the context it falls back to the logical name", func(t *testing.T) {
		got := AutoName(t.Context(), infer.CheckRequest{
			Name:      testAutoNameLogical,
			NewInputs: property.NewMap(nil),
		}, testAutoNameField, testAutoNameMaxLen)
		assert.True(t, strings.HasPrefix(nameOf(got).AsString(), testAutoNameLogical+"-"))
	})
}
