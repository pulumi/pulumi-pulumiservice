package util

import (
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
)

func FromProperties(props *structpb.Struct, out interface{}) error {
	inputs, err := plugin.UnmarshalProperties(props, StandardUnmarshal)
	if err != nil {
		return err
	}
	return FromPropertyMap(inputs, out)
}

func ToProperties(obj interface{}) (*structpb.Struct, error) {
	propertyMap := ToPropertyMap(obj)
	return plugin.MarshalProperties(propertyMap, StandardMarshal)
}
