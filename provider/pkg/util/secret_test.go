// Copyright 2016-2026, Pulumi Corporation.
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

	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
)

func TestGetSecretOrStringValue(t *testing.T) {
	t.Run("plaintext string", func(t *testing.T) {
		prop := resource.NewPropertyValue("hello")
		assert.Equal(t, "hello", GetSecretOrStringValue(prop))
	})

	t.Run("secret string", func(t *testing.T) {
		prop := resource.MakeSecret(resource.NewPropertyValue("secret-val"))
		assert.Equal(t, "secret-val", GetSecretOrStringValue(prop))
	})

	t.Run("nil value", func(t *testing.T) {
		prop := resource.PropertyValue{}
		assert.Equal(t, "", GetSecretOrStringValue(prop))
	})
}

func TestGetSecretOrStringNullableValue(t *testing.T) {
	t.Run("plaintext string", func(t *testing.T) {
		prop := resource.NewPropertyValue("hello")
		result := GetSecretOrStringNullableValue(prop)
		assert.NotNil(t, result)
		assert.Equal(t, "hello", *result)
	})

	t.Run("secret string", func(t *testing.T) {
		prop := resource.MakeSecret(resource.NewPropertyValue("secret-val"))
		result := GetSecretOrStringNullableValue(prop)
		assert.NotNil(t, result)
		assert.Equal(t, "secret-val", *result)
	})

	t.Run("nil value", func(t *testing.T) {
		prop := resource.PropertyValue{}
		assert.Nil(t, GetSecretOrStringNullableValue(prop))
	})
}

func TestGetSecretOrBoolValue(t *testing.T) {
	t.Run("plaintext bool", func(t *testing.T) {
		prop := resource.NewPropertyValue(true)
		assert.True(t, GetSecretOrBoolValue(prop))
	})

	t.Run("secret bool", func(t *testing.T) {
		prop := resource.MakeSecret(resource.NewPropertyValue(true))
		assert.True(t, GetSecretOrBoolValue(prop))
	})

	t.Run("plaintext false", func(t *testing.T) {
		prop := resource.NewPropertyValue(false)
		assert.False(t, GetSecretOrBoolValue(prop))
	})
}

func TestGetSecretOrNumberValue(t *testing.T) {
	t.Run("plaintext number", func(t *testing.T) {
		prop := resource.NewPropertyValue(42.0)
		assert.Equal(t, 42.0, GetSecretOrNumberValue(prop))
	})

	t.Run("secret number", func(t *testing.T) {
		prop := resource.MakeSecret(resource.NewPropertyValue(3.14))
		assert.Equal(t, 3.14, GetSecretOrNumberValue(prop))
	})
}

func TestGetSecretOrArrayValue(t *testing.T) {
	t.Run("plaintext array", func(t *testing.T) {
		arr := []resource.PropertyValue{
			resource.NewPropertyValue("a"),
			resource.NewPropertyValue("b"),
		}
		prop := resource.NewPropertyValue(arr)
		result := GetSecretOrArrayValue(prop)
		assert.Len(t, result, 2)
		assert.Equal(t, "a", result[0].StringValue())
		assert.Equal(t, "b", result[1].StringValue())
	})

	t.Run("secret array", func(t *testing.T) {
		arr := []resource.PropertyValue{
			resource.NewPropertyValue("x"),
		}
		prop := resource.MakeSecret(resource.NewPropertyValue(arr))
		result := GetSecretOrArrayValue(prop)
		assert.Len(t, result, 1)
		assert.Equal(t, "x", result[0].StringValue())
	})
}

func TestGetSecretOrObjectValue(t *testing.T) {
	t.Run("plaintext object", func(t *testing.T) {
		obj := resource.PropertyMap{
			"key": resource.NewPropertyValue("value"),
		}
		prop := resource.NewPropertyValue(obj)
		result := GetSecretOrObjectValue(prop)
		assert.Equal(t, "value", result["key"].StringValue())
	})

	t.Run("secret object", func(t *testing.T) {
		obj := resource.PropertyMap{
			"key": resource.NewPropertyValue("secret-value"),
		}
		prop := resource.MakeSecret(resource.NewPropertyValue(obj))
		result := GetSecretOrObjectValue(prop)
		assert.Equal(t, "secret-value", result["key"].StringValue())
	})
}

// nested builds {"a": {"b": {"c": leaf}}}.
func nested(leaf resource.PropertyValue) resource.PropertyMap {
	return resource.PropertyMap{
		"a": resource.NewObjectProperty(resource.PropertyMap{
			"b": resource.NewObjectProperty(resource.PropertyMap{
				"c": leaf,
			}),
		}),
	}
}

func TestMakeNestedSecret(t *testing.T) {
	t.Run("promotes a plaintext leaf", func(t *testing.T) {
		props := nested(resource.NewStringProperty("hunter2"))

		MakeNestedSecret(props, "a", "b", "c")

		leaf := props["a"].ObjectValue()["b"].ObjectValue()["c"]
		assert.True(t, leaf.IsSecret(), "leaf should have been promoted")
		assert.Equal(t, "hunter2", leaf.SecretValue().Element.StringValue())
	})

	t.Run("leaves an already secret leaf untouched", func(t *testing.T) {
		props := nested(resource.MakeSecret(resource.NewStringProperty("hunter2")))

		MakeNestedSecret(props, "a", "b", "c")

		leaf := props["a"].ObjectValue()["b"].ObjectValue()["c"]
		assert.True(t, leaf.IsSecret())
		// Not double-wrapped: the element is still the string, not another secret.
		assert.Equal(t, "hunter2", leaf.SecretValue().Element.StringValue())
	})

	t.Run("walks through a secret intermediate object", func(t *testing.T) {
		props := resource.PropertyMap{
			"a": resource.MakeSecret(resource.NewObjectProperty(resource.PropertyMap{
				"b": resource.NewObjectProperty(resource.PropertyMap{
					"c": resource.NewStringProperty("hunter2"),
				}),
			})),
		}

		MakeNestedSecret(props, "a", "b", "c")

		leaf := props["a"].SecretValue().Element.ObjectValue()["b"].ObjectValue()["c"]
		assert.True(t, leaf.IsSecret())
		assert.Equal(t, "hunter2", leaf.SecretValue().Element.StringValue())
	})

	t.Run("no-op when the path is missing", func(t *testing.T) {
		props := nested(resource.NewStringProperty("hunter2"))
		before := props.Copy()

		MakeNestedSecret(props, "a", "nope", "c")
		MakeNestedSecret(props, "a", "b", "nope")
		MakeNestedSecret(props, "nope")

		assert.Equal(t, before, props)
	})

	t.Run("no-op when an intermediate level is not an object", func(t *testing.T) {
		props := resource.PropertyMap{
			"a": resource.NewStringProperty("not-an-object"),
		}
		before := props.Copy()

		MakeNestedSecret(props, "a", "b", "c")

		assert.Equal(t, before, props)
	})

	t.Run("no-op on an empty path", func(t *testing.T) {
		props := nested(resource.NewStringProperty("hunter2"))
		before := props.Copy()

		MakeNestedSecret(props)

		assert.Equal(t, before, props)
	})

	t.Run("no-op on a null leaf", func(t *testing.T) {
		props := nested(resource.NewNullProperty())
		before := props.Copy()

		MakeNestedSecret(props, "a", "b", "c")

		assert.Equal(t, before, props)
	})
}
