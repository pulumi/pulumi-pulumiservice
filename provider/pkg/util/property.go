package util // nolint:revive // util is a common and acceptable package name

import (
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
)

func FromProperties(props *structpb.Struct, structTagName string, out interface{}) error {
	inputs, err := plugin.UnmarshalProperties(props, StandardUnmarshal)
	if err != nil {
		return err
	}
	return FromPropertyMap(inputs, structTagName, out)
}

func ToProperties(obj interface{}, structTagName string) (*structpb.Struct, error) {
	propertyMap := ToPropertyMap(obj, structTagName)
	return plugin.MarshalProperties(propertyMap, StandardMarshal)
}
