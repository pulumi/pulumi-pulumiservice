package provider

import (
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/internal/pulumiapi"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
)

func getSecretOrStringValue(prop resource.PropertyValue) string {
	switch prop.V.(type) {
	case *resource.Secret:
		return prop.SecretValue().Element.StringValue()
	case nil:
		return ""
	default:
		return prop.StringValue()
	}
}

func getSecretOrStringNullableValue(prop resource.PropertyValue) *string {
	var resultString string
	switch prop.V.(type) {
	case *resource.Secret:
		resultString = prop.SecretValue().Element.StringValue()
	case nil:
		return nil
	default:
		resultString = prop.StringValue()
	}
	return &resultString
}

// All imported inputs will have a dummy value, asking to be replaced in real code
// All imported properties are just set to ciphertext read from Pulumi Service
func importSecretValue(propertyMap resource.PropertyMap, propertyName string, cipherValue pulumiapi.SecretValue, isInput bool) {
	if isInput {
		propertyMap[resource.PropertyKey(propertyName)] = resource.MakeSecret(resource.NewPropertyValue(replaceMe))
	} else {
		propertyMap[resource.PropertyKey(propertyName)] = resource.NewPropertyValue(cipherValue.Value)
	}
}

// On Create or Update, inputs already have a plaintext value, just set it
// Properties are just set to ciphertext returned from Pulumi Service
func createSecretValue(propertyMap resource.PropertyMap, propertyName string, cipherValue pulumiapi.SecretValue, plaintextValue pulumiapi.SecretValue, isInput bool) {
	if isInput {
		propertyMap[resource.PropertyKey(propertyName)] = resource.MakeSecret(resource.NewPropertyValue(plaintextValue.Value))
	} else {
		propertyMap[resource.PropertyKey(propertyName)] = resource.NewPropertyValue(cipherValue.Value)
	}
}

// Merge happens when existing resource is refreshed from Pulumi Service
// Output properties are just replaced with ciphertext retrieved from Pulumi Service
// Inputs are more complicated :
// If ciphertext never changed, keep existing plaintext value
// If ciphertext is different, set plaintext to empty string
// If retrieved state has a value that current state does not have, pass in nil, which will fill plaintext with empty string
func mergeSecretValue(propertyMap resource.PropertyMap, propertyName string, cipherValue pulumiapi.SecretValue, plaintextValue *pulumiapi.SecretValue, oldCipherValue *pulumiapi.SecretValue, isInput bool) {
	if isInput {
		if oldCipherValue != nil && cipherValue.Value == oldCipherValue.Value {
			propertyMap[resource.PropertyKey(propertyName)] = resource.MakeSecret(resource.NewPropertyValue(plaintextValue.Value))
		} else {
			propertyMap[resource.PropertyKey(propertyName)] = resource.MakeSecret(resource.NewPropertyValue(""))
		}
	} else {
		propertyMap[resource.PropertyKey(propertyName)] = resource.NewPropertyValue(cipherValue.Value)
	}
}
