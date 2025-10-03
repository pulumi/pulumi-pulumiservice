package util

import (
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
)

var StandardUnmarshal plugin.MarshalOptions = plugin.MarshalOptions{
	KeepUnknowns: false,
	SkipNulls:    true,
	KeepSecrets:  false,
}

var StandardMarshal plugin.MarshalOptions = plugin.MarshalOptions{
	KeepUnknowns: false,
	SkipNulls:    true,
	KeepSecrets:  true,
}

// These options should be used when we need to know whether a field was a secret or not.
// These should also be always used in Check() method, otherwise secrets leak on preview
// If you do use it, make sure all the methods use getSecretOrBlankValue() methods found below
var KeepSecretsUnmarshal plugin.MarshalOptions = StandardMarshal

func GetSecretOrStringValue(prop resource.PropertyValue) string {
	switch prop.V.(type) {
	case *resource.Secret:
		return prop.SecretValue().Element.StringValue()
	case nil:
		return ""
	default:
		return prop.StringValue()
	}
}

func GetSecretOrStringNullableValue(prop resource.PropertyValue) *string {
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

func GetSecretOrBoolValue(prop resource.PropertyValue) bool {
	switch prop.V.(type) {
	case *resource.Secret:
		return prop.SecretValue().Element.BoolValue()
	default:
		return prop.BoolValue()
	}
}

func GetSecretOrArrayValue(prop resource.PropertyValue) []resource.PropertyValue {
	switch prop.V.(type) {
	case *resource.Secret:
		return prop.SecretValue().Element.ArrayValue()
	default:
		return prop.ArrayValue()
	}
}

func GetSecretOrObjectValue(prop resource.PropertyValue) resource.PropertyMap {
	switch prop.V.(type) {
	case *resource.Secret:
		return prop.SecretValue().Element.ObjectValue()
	default:
		return prop.ObjectValue()
	}
}

// This is a value for imported secrets, to hint that value needs to be replaced
// in generated code
const replaceMe = "<REPLACE WITH ACTUAL SECRET VALUE>"

// ImportSecretValue sets secret values for imported resources.
// All imported inputs will have a dummy value, asking to be replaced in real code.
// All imported properties are just set to ciphertext read from Pulumi Service.
func ImportSecretValue(
	propertyMap resource.PropertyMap,
	propertyName string,
	cipherValue pulumiapi.SecretValue,
	isInput bool,
) {
	if isInput {
		propertyMap[resource.PropertyKey(propertyName)] = resource.MakeSecret(resource.NewPropertyValue(replaceMe))
	} else {
		propertyMap[resource.PropertyKey(propertyName)] = resource.NewPropertyValue(cipherValue.Value)
	}
}

// CreateSecretValue sets secret values for created or updated resources.
// On Create or Update, inputs already have a plaintext value, just set it.
// Properties are just set to ciphertext returned from Pulumi Service.
func CreateSecretValue(
	propertyMap resource.PropertyMap,
	propertyName string,
	cipherValue pulumiapi.SecretValue,
	plaintextValue pulumiapi.SecretValue,
	isInput bool,
) {
	if isInput {
		propertyMap[resource.PropertyKey(propertyName)] = resource.MakeSecret(
			resource.NewPropertyValue(plaintextValue.Value),
		)
	} else {
		propertyMap[resource.PropertyKey(propertyName)] = resource.NewPropertyValue(cipherValue.Value)
	}
}

// MergeSecretValue merges secret values when an existing resource is refreshed from Pulumi Service.
// Output properties are just replaced with ciphertext retrieved from Pulumi Service.
// Inputs are more complicated:
// If ciphertext never changed, keep existing plaintext value.
// If ciphertext is different, set plaintext to empty string.
// If retrieved state has a value that current state does not have, pass in nil,
// which will fill plaintext with empty string.
func MergeSecretValue(
	propertyMap resource.PropertyMap,
	propertyName string,
	cipherValue pulumiapi.SecretValue,
	plaintextValue *pulumiapi.SecretValue,
	oldCipherValue *pulumiapi.SecretValue,
	isInput bool,
) {
	if isInput {
		if oldCipherValue != nil && cipherValue.Value == oldCipherValue.Value {
			propertyMap[resource.PropertyKey(propertyName)] = resource.MakeSecret(
				resource.NewPropertyValue(plaintextValue.Value),
			)
		} else {
			propertyMap[resource.PropertyKey(propertyName)] = resource.MakeSecret(resource.NewPropertyValue(""))
		}
	} else {
		propertyMap[resource.PropertyKey(propertyName)] = resource.NewPropertyValue(cipherValue.Value)
	}
}
