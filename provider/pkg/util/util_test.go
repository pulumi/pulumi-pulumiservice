// Copyright 2016-2025, Pulumi Corporation.
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

package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestElementsEqual(t *testing.T) {
	t.Run("equal slices same order", func(t *testing.T) {
		a := []int{1, 2, 3}
		b := []int{1, 2, 3}
		assert.True(t, ElementsEqual(a, b))
	})

	t.Run("equal slices different order", func(t *testing.T) {
		a := []int{1, 2, 3}
		b := []int{3, 1, 2}
		assert.True(t, ElementsEqual(a, b))
	})

	t.Run("different lengths", func(t *testing.T) {
		a := []int{1, 2, 3}
		b := []int{1, 2}
		assert.False(t, ElementsEqual(a, b))
	})

	t.Run("different elements", func(t *testing.T) {
		a := []int{1, 2, 3}
		b := []int{1, 2, 4}
		assert.False(t, ElementsEqual(a, b))
	})

	t.Run("empty slices", func(t *testing.T) {
		a := []int{}
		b := []int{}
		assert.True(t, ElementsEqual(a, b))
	})

	t.Run("with strings", func(t *testing.T) {
		a := []string{"foo", "bar", "baz"}
		b := []string{"baz", "foo", "bar"}
		assert.True(t, ElementsEqual(a, b))
	})
}

type testStruct struct {
	Name  string
	Value int
}

func TestElementsEqualFunc(t *testing.T) {
	cmpFunc := func(a, b testStruct) int {
		if a.Name < b.Name {
			return -1
		}
		if a.Name > b.Name {
			return 1
		}
		if a.Value < b.Value {
			return -1
		}
		if a.Value > b.Value {
			return 1
		}
		return 0
	}

	eqFunc := func(a, b testStruct) bool {
		return a.Name == b.Name && a.Value == b.Value
	}

	t.Run("equal slices same order", func(t *testing.T) {
		a := []testStruct{{Name: "a", Value: 1}, {Name: "b", Value: 2}}
		b := []testStruct{{Name: "a", Value: 1}, {Name: "b", Value: 2}}
		assert.True(t, ElementsEqualFunc(a, b, cmpFunc, eqFunc))
	})

	t.Run("equal slices different order", func(t *testing.T) {
		a := []testStruct{{Name: "a", Value: 1}, {Name: "b", Value: 2}}
		b := []testStruct{{Name: "b", Value: 2}, {Name: "a", Value: 1}}
		assert.True(t, ElementsEqualFunc(a, b, cmpFunc, eqFunc))
	})

	t.Run("different lengths", func(t *testing.T) {
		a := []testStruct{{Name: "a", Value: 1}, {Name: "b", Value: 2}}
		b := []testStruct{{Name: "a", Value: 1}}
		assert.False(t, ElementsEqualFunc(a, b, cmpFunc, eqFunc))
	})

	t.Run("different elements", func(t *testing.T) {
		a := []testStruct{{Name: "a", Value: 1}, {Name: "b", Value: 2}}
		b := []testStruct{{Name: "a", Value: 1}, {Name: "c", Value: 2}}
		assert.False(t, ElementsEqualFunc(a, b, cmpFunc, eqFunc))
	})

	t.Run("empty slices", func(t *testing.T) {
		a := []testStruct{}
		b := []testStruct{}
		assert.True(t, ElementsEqualFunc(a, b, cmpFunc, eqFunc))
	})

	t.Run("custom equality function", func(t *testing.T) {
		// Only compare by Name, ignore Value
		eqByName := func(a, b testStruct) bool {
			return a.Name == b.Name
		}
		a := []testStruct{{Name: "a", Value: 1}, {Name: "b", Value: 2}}
		b := []testStruct{{Name: "b", Value: 999}, {Name: "a", Value: 888}}
		assert.True(t, ElementsEqualFunc(a, b, cmpFunc, eqByName))
	})
}
