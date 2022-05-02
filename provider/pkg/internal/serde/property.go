package serde

import (
	structpb "github.com/golang/protobuf/ptypes/struct"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
)

func FromProperties(props *structpb.Struct, structTagName string, out interface{}) error {
	inputs, err := plugin.UnmarshalProperties(props, plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
	if err != nil {
		return err
	}
	return FromPropertyMap(inputs, structTagName, out)
}

func ToProperties(obj interface{}, structTagName string) (*structpb.Struct, error) {
	propertyMap := ToPropertyMap(obj, structTagName)
	return plugin.MarshalProperties(propertyMap, plugin.MarshalOptions{})
}
