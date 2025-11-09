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

package resources

import (
	"testing"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/util"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/stretchr/testify/assert"
)

func TestConvertMapToPropertyMap(t *testing.T) {
	tests := []struct {
		name           string
		input          map[string]interface{}
		expectedOutput resource.PropertyMap
	}{
		{
			name: "Simple string values",
			input: map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
			},
			expectedOutput: resource.PropertyMap{
				"key1": resource.NewPropertyValue("value1"),
				"key2": resource.NewPropertyValue("value2"),
			},
		},
		{
			name: "Array of strings (the main fix)",
			input: map[string]interface{}{
				"approvedAmiIds": []interface{}{"ami-0abcdef1234567890", "ami-1234567890abcdef"},
				"regions":        []interface{}{"us-east-1", "us-west-2"},
			},
			expectedOutput: resource.PropertyMap{
				"approvedAmiIds": resource.NewArrayProperty([]resource.PropertyValue{
					resource.NewPropertyValue("ami-0abcdef1234567890"),
					resource.NewPropertyValue("ami-1234567890abcdef"),
				}),
				"regions": resource.NewArrayProperty([]resource.PropertyValue{
					resource.NewPropertyValue("us-east-1"),
					resource.NewPropertyValue("us-west-2"),
				}),
			},
		},
		{
			name: "Nested object",
			input: map[string]interface{}{
				"nestedObj": map[string]interface{}{
					"innerKey": "innerValue",
					"innerNum": float64(42),
				},
			},
			expectedOutput: resource.PropertyMap{
				"nestedObj": resource.NewObjectProperty(resource.PropertyMap{
					"innerKey": resource.NewPropertyValue("innerValue"),
					"innerNum": resource.NewPropertyValue(float64(42)),
				}),
			},
		},
		{
			name: "Complex nested structure with arrays",
			input: map[string]interface{}{
				"config": map[string]interface{}{
					"approvedAmiIds": []interface{}{"ami-123", "ami-456"},
					"settings": map[string]interface{}{
						"enabled": true,
						"count":   float64(5),
					},
				},
			},
			expectedOutput: resource.PropertyMap{
				"config": resource.NewObjectProperty(resource.PropertyMap{
					"approvedAmiIds": resource.NewArrayProperty([]resource.PropertyValue{
						resource.NewPropertyValue("ami-123"),
						resource.NewPropertyValue("ami-456"),
					}),
					"settings": resource.NewObjectProperty(resource.PropertyMap{
						"enabled": resource.NewPropertyValue(true),
						"count":   resource.NewPropertyValue(float64(5)),
					}),
				}),
			},
		},
		{
			name: "Array of objects",
			input: map[string]interface{}{
				"items": []interface{}{
					map[string]interface{}{"name": "item1", "value": float64(10)},
					map[string]interface{}{"name": "item2", "value": float64(20)},
				},
			},
			expectedOutput: resource.PropertyMap{
				"items": resource.NewArrayProperty([]resource.PropertyValue{
					resource.NewObjectProperty(resource.PropertyMap{
						"name":  resource.NewPropertyValue("item1"),
						"value": resource.NewPropertyValue(float64(10)),
					}),
					resource.NewObjectProperty(resource.PropertyMap{
						"name":  resource.NewPropertyValue("item2"),
						"value": resource.NewPropertyValue(float64(20)),
					}),
				}),
			},
		},
		{
			name:           "Empty map",
			input:          map[string]interface{}{},
			expectedOutput: resource.PropertyMap{},
		},
		{
			name: "Mixed types",
			input: map[string]interface{}{
				"stringVal": "hello",
				"numberVal": float64(42.5),
				"boolVal":   true,
				"arrayVal":  []interface{}{"a", "b", "c"},
			},
			expectedOutput: resource.PropertyMap{
				"stringVal": resource.NewPropertyValue("hello"),
				"numberVal": resource.NewPropertyValue(float64(42.5)),
				"boolVal":   resource.NewPropertyValue(true),
				"arrayVal": resource.NewArrayProperty([]resource.PropertyValue{
					resource.NewPropertyValue("a"),
					resource.NewPropertyValue("b"),
					resource.NewPropertyValue("c"),
				}),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := util.ConvertMapToPropertyMap(tt.input)
			assert.Equal(t, tt.expectedOutput, result)
		})
	}
}

func TestConvertInterfaceToPropertyValue(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected resource.PropertyValue
	}{
		{
			name:     "String value",
			input:    "test-string",
			expected: resource.NewPropertyValue("test-string"),
		},
		{
			name:     "Number value",
			input:    float64(123.45),
			expected: resource.NewPropertyValue(float64(123.45)),
		},
		{
			name:     "Boolean value",
			input:    true,
			expected: resource.NewPropertyValue(true),
		},
		{
			name:  "Array of strings",
			input: []interface{}{"a", "b", "c"},
			expected: resource.NewArrayProperty([]resource.PropertyValue{
				resource.NewPropertyValue("a"),
				resource.NewPropertyValue("b"),
				resource.NewPropertyValue("c"),
			}),
		},
		{
			name:     "Empty array",
			input:    []interface{}{},
			expected: resource.NewArrayProperty([]resource.PropertyValue{}),
		},
		{
			name: "Map",
			input: map[string]interface{}{
				"key": "value",
			},
			expected: resource.NewObjectProperty(resource.PropertyMap{
				"key": resource.NewPropertyValue("value"),
			}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := util.ConvertInterfaceToPropertyValue(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test the specific use case from the PolicyGroup issue
func TestPolicyPackConfigConversion(t *testing.T) {
	// Simulate the API response structure for policy pack config
	apiConfig := map[string]interface{}{
		"approvedAmiIds": []interface{}{"ami-0abcdef1234567890"},
		"allowedRegions": []interface{}{"us-east-1", "us-west-2"},
		"maxInstances":   float64(10),
		"requireTags":    true,
	}

	result := util.ConvertMapToPropertyMap(apiConfig)

	// Verify approvedAmiIds is properly converted as an array of strings
	assert.True(t, result["approvedAmiIds"].IsArray())
	arrayValues := result["approvedAmiIds"].ArrayValue()
	assert.Len(t, arrayValues, 1)
	assert.True(t, arrayValues[0].IsString())
	assert.Equal(t, "ami-0abcdef1234567890", arrayValues[0].StringValue())

	// Verify other fields are correct
	assert.True(t, result["allowedRegions"].IsArray())
	regions := result["allowedRegions"].ArrayValue()
	assert.Len(t, regions, 2)
	assert.Equal(t, "us-east-1", regions[0].StringValue())
	assert.Equal(t, "us-west-2", regions[1].StringValue())

	assert.True(t, result["maxInstances"].IsNumber())
	assert.Equal(t, float64(10), result["maxInstances"].NumberValue())

	assert.True(t, result["requireTags"].IsBool())
	assert.Equal(t, true, result["requireTags"].BoolValue())
}

// Test deserialization (the reverse direction)
func TestConvertPropertyMapToMap(t *testing.T) {
	tests := []struct {
		name     string
		input    resource.PropertyMap
		expected map[string]interface{}
	}{
		{
			name: "Simple string values",
			input: resource.PropertyMap{
				"key1": resource.NewPropertyValue("value1"),
				"key2": resource.NewPropertyValue("value2"),
			},
			expected: map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
			},
		},
		{
			name: "Array of strings (the main deserialization fix)",
			input: resource.PropertyMap{
				"approvedAmiIds": resource.NewArrayProperty([]resource.PropertyValue{
					resource.NewPropertyValue("ami-0abcdef1234567890"),
					resource.NewPropertyValue("ami-1234567890abcdef"),
				}),
				"regions": resource.NewArrayProperty([]resource.PropertyValue{
					resource.NewPropertyValue("us-east-1"),
					resource.NewPropertyValue("us-west-2"),
				}),
			},
			expected: map[string]interface{}{
				"approvedAmiIds": []interface{}{"ami-0abcdef1234567890", "ami-1234567890abcdef"},
				"regions":        []interface{}{"us-east-1", "us-west-2"},
			},
		},
		{
			name: "Nested object",
			input: resource.PropertyMap{
				"nestedObj": resource.NewObjectProperty(resource.PropertyMap{
					"innerKey": resource.NewPropertyValue("innerValue"),
					"innerNum": resource.NewPropertyValue(float64(42)),
				}),
			},
			expected: map[string]interface{}{
				"nestedObj": map[string]interface{}{
					"innerKey": "innerValue",
					"innerNum": float64(42),
				},
			},
		},
		{
			name: "Complex nested structure with arrays",
			input: resource.PropertyMap{
				"config": resource.NewObjectProperty(resource.PropertyMap{
					"approvedAmiIds": resource.NewArrayProperty([]resource.PropertyValue{
						resource.NewPropertyValue("ami-123"),
						resource.NewPropertyValue("ami-456"),
					}),
					"settings": resource.NewObjectProperty(resource.PropertyMap{
						"enabled": resource.NewPropertyValue(true),
						"count":   resource.NewPropertyValue(float64(5)),
					}),
				}),
			},
			expected: map[string]interface{}{
				"config": map[string]interface{}{
					"approvedAmiIds": []interface{}{"ami-123", "ami-456"},
					"settings": map[string]interface{}{
						"enabled": true,
						"count":   float64(5),
					},
				},
			},
		},
		{
			name: "Array of objects",
			input: resource.PropertyMap{
				"items": resource.NewArrayProperty([]resource.PropertyValue{
					resource.NewObjectProperty(resource.PropertyMap{
						"name":  resource.NewPropertyValue("item1"),
						"value": resource.NewPropertyValue(float64(10)),
					}),
					resource.NewObjectProperty(resource.PropertyMap{
						"name":  resource.NewPropertyValue("item2"),
						"value": resource.NewPropertyValue(float64(20)),
					}),
				}),
			},
			expected: map[string]interface{}{
				"items": []interface{}{
					map[string]interface{}{"name": "item1", "value": float64(10)},
					map[string]interface{}{"name": "item2", "value": float64(20)},
				},
			},
		},
		{
			name:     "Empty map",
			input:    resource.PropertyMap{},
			expected: map[string]interface{}{},
		},
		{
			name: "Mixed types",
			input: resource.PropertyMap{
				"stringVal": resource.NewPropertyValue("hello"),
				"numberVal": resource.NewPropertyValue(float64(42.5)),
				"boolVal":   resource.NewPropertyValue(true),
				"arrayVal": resource.NewArrayProperty([]resource.PropertyValue{
					resource.NewPropertyValue("a"),
					resource.NewPropertyValue("b"),
					resource.NewPropertyValue("c"),
				}),
			},
			expected: map[string]interface{}{
				"stringVal": "hello",
				"numberVal": float64(42.5),
				"boolVal":   true,
				"arrayVal":  []interface{}{"a", "b", "c"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := util.ConvertPropertyMapToMap(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConvertPropertyValueToInterface(t *testing.T) {
	tests := []struct {
		name     string
		input    resource.PropertyValue
		expected interface{}
	}{
		{
			name:     "String value",
			input:    resource.NewPropertyValue("test-string"),
			expected: "test-string",
		},
		{
			name:     "Number value",
			input:    resource.NewPropertyValue(float64(123.45)),
			expected: float64(123.45),
		},
		{
			name:     "Boolean value",
			input:    resource.NewPropertyValue(true),
			expected: true,
		},
		{
			name: "Array of strings",
			input: resource.NewArrayProperty([]resource.PropertyValue{
				resource.NewPropertyValue("a"),
				resource.NewPropertyValue("b"),
				resource.NewPropertyValue("c"),
			}),
			expected: []interface{}{"a", "b", "c"},
		},
		{
			name:     "Empty array",
			input:    resource.NewArrayProperty([]resource.PropertyValue{}),
			expected: []interface{}{},
		},
		{
			name: "Object",
			input: resource.NewObjectProperty(resource.PropertyMap{
				"key": resource.NewPropertyValue("value"),
			}),
			expected: map[string]interface{}{
				"key": "value",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := util.ConvertPropertyValueToInterface(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test roundtrip conversion (API response -> PropertyMap -> Map for API input)
func TestPolicyPackConfigRoundtrip(t *testing.T) {
	// Original config as it would come from the API
	originalConfig := map[string]interface{}{
		"approvedAmiIds": []interface{}{"ami-0abcdef1234567890"},
		"allowedRegions": []interface{}{"us-east-1", "us-west-2"},
		"maxInstances":   float64(10),
		"requireTags":    true,
		"nestedSettings": map[string]interface{}{
			"timeout": float64(300),
			"retries": float64(3),
		},
	}

	// Convert to PropertyMap (as would happen when reading from API)
	propertyMap := util.ConvertMapToPropertyMap(originalConfig)

	// Convert back to map (as would happen when sending to API)
	finalConfig := util.ConvertPropertyMapToMap(propertyMap)

	// Should match the original
	assert.Equal(t, originalConfig, finalConfig)
}

// Test the refactored PolicyGroup serialization with complex policy pack configs
func TestPolicyGroupSerializationConsistency(t *testing.T) {
	// Create a policy group input with complex config
	input := PulumiServicePolicyGroupInput{
		Name:             "test-group",
		OrganizationName: "test-org",
		EntityType:       "stacks",
		Mode:             "audit",
		Stacks: []pulumiapi.StackReference{
			{Name: "stack1", RoutingProject: "project1"},
			{Name: "stack2", RoutingProject: "project2"},
		},
		PolicyPacks: []pulumiapi.PolicyPackMetadata{
			{
				Name:        "aws-compliance",
				DisplayName: "AWS Compliance Pack",
				Version:     1,
				VersionTag:  "v1.0.0",
				Config: map[string]interface{}{
					"approvedAmiIds": []interface{}{"ami-123", "ami-456"},
					"maxInstances":   float64(10),
					"nestedConfig": map[string]interface{}{
						"regions": []interface{}{"us-east-1", "us-west-2"},
						"enabled": true,
					},
				},
			},
		},
	}

	// Convert to PropertyMap using refactored method
	propertyMap := input.ToPropertyMap()

	// Convert back using refactored method
	roundtripInput := ToPulumiServicePolicyGroupInput(propertyMap)

	// Verify all fields are preserved correctly
	assert.Equal(t, input.Name, roundtripInput.Name)
	assert.Equal(t, input.OrganizationName, roundtripInput.OrganizationName)
	assert.Equal(t, input.EntityType, roundtripInput.EntityType)
	assert.Equal(t, input.Mode, roundtripInput.Mode)
	assert.Equal(t, input.Stacks, roundtripInput.Stacks)
	
	// Verify policy pack details including complex config
	assert.Len(t, roundtripInput.PolicyPacks, 1)
	pp := roundtripInput.PolicyPacks[0]
	assert.Equal(t, input.PolicyPacks[0].Name, pp.Name)
	assert.Equal(t, input.PolicyPacks[0].DisplayName, pp.DisplayName)
	assert.Equal(t, input.PolicyPacks[0].Version, pp.Version)
	assert.Equal(t, input.PolicyPacks[0].VersionTag, pp.VersionTag)
	
	// Verify complex config is preserved
	originalConfig := input.PolicyPacks[0].Config
	roundtripConfig := pp.Config
	assert.Equal(t, originalConfig, roundtripConfig)
}