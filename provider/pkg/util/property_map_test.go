package util

import (
	"testing"

	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/stretchr/testify/assert"
)

func TestToPropertyMap(t *testing.T) {
	type A struct {
		IntField      int      `pulumi:"int_field"`
		FloatField    float64  `pulumi:"float_field"`
		StringField   string   `pulumi:"string_field"`
		FloatPtrField *float64 `pulumi:"floatptr_field"`
		IntPtrField   *int64   `json:"intptr_field"`
	}
	intvalue := int64(1)
	floatValue := 2.5
	a := A{
		IntField:      int(intvalue),
		FloatField:    floatValue,
		StringField:   "example",
		FloatPtrField: &floatValue,
		IntPtrField:   &intvalue,
	}
	propertyMap := ToPropertyMap(a)
	expected := map[string]interface{}{
		// NewPropertyValue() auto converts ints to float64
		"int_field":    float64(1),
		"float_field":  float64(2.5),
		"string_field": "example",
	}
	for key, value := range expected {
		actualValue := propertyMap[resource.PropertyKey(key)].V
		assert.Equal(t, value, actualValue, key)
	}
}

func TestFromPropertyMap(t *testing.T) {
	type A struct {
		IntField      int      `pulumi:"int_field"`
		FloatField    float64  `pulumi:"float_field"`
		StringField   string   `pulumi:"string_field"`
		FloatPtrField *float64 `pulumi:"floatptr_field"`
		IntPtrField   *int64   `pulumi:"intptr_field"`
	}
	t.Run("Unmarshals Props Properly", func(t *testing.T) {
		intValue := int64(1)
		floatValue := 2.5
		want := A{
			IntField:      int(intValue),
			FloatField:    floatValue,
			StringField:   "example",
			FloatPtrField: &floatValue,
			IntPtrField:   &intValue,
		}
		propertyMap := resource.PropertyMap{}
		expected := map[string]interface{}{
			// NewPropertyValue() auto converts ints to float64
			"int_field":      float64(1),
			"float_field":    float64(2.5),
			"string_field":   "example",
			"floatptr_field": &floatValue,
			"intptr_field":   &intValue,
		}
		for key, value := range expected {
			propertyMap[resource.PropertyKey(key)] = resource.NewPropertyValue(value)
		}
		got := A{}
		FromPropertyMap(propertyMap, &got)
		assert.Equal(t, want, got)
	})

	t.Run("Skips Missing Keys", func(t *testing.T) {
		propertyMap := resource.PropertyMap{}
		// property map is empty, no fields should be set
		want := A{}

		got := A{}
		FromPropertyMap(propertyMap, &got)

		// no error should be returned, even though struct has pulumi tags
		// that aren't present in propertyMap
		assert.Equal(t, want, got)
	})

}
