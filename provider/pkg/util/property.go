package util

import (
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	"google.golang.org/protobuf/types/known/structpb"
)

func FromProperties(props *structpb.Struct, out interface{}) error {
	inputs, err := plugin.UnmarshalProperties(props, StandardUnmarshal)
	if err != nil {
		return err
	}
	FromPropertyMap(inputs, out)
	return nil
}

func ToProperties(obj interface{}) (*structpb.Struct, error) {
	propertyMap := ToPropertyMap(obj)
	properties, err := plugin.MarshalProperties(propertyMap, StandardMarshal)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal input to properties: %w", err)
	}
	return properties, nil
}
