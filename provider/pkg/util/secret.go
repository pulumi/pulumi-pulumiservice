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

// This is a value for imported secrets, to hint that value needs to be replaced
// in generated code
const replaceMe = "<REPLACE WITH ACTUAL SECRET VALUE>"

// All imported inputs will have a dummy value, asking to be replaced in real code
// All imported properties are just set to ciphertext read from Pulumi Service
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

// On Create or Update, inputs already have a plaintext value, just set it
// Properties are just set to ciphertext returned from Pulumi Service
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

// MergeSecretValue is the permissive merge: it cannot detect server-side
// rotation. As long as we have any prior cipher state for this secret, we
// trust the user's plaintext input. Use this when the server returns
// non-deterministic ciphertext per call (e.g. AES-GCM with a random IV, as
// the Pulumi Cloud deployment-settings endpoint does today).
//
// Output properties are just replaced with ciphertext retrieved from Pulumi Service.
// Inputs:
//   - have prior state → keep plaintext from user inputs
//   - no prior state   → empty plaintext (engine will see a diff)
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
