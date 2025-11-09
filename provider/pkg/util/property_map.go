package util

import (
	"fmt"
	"reflect"

	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
)

// ToPropertyMap marshals a struct into a resource.PropertyMap. It obtains
// the resource.PropertyKey() values for each struct field by grabbing
// value of the structTagKey
func ToPropertyMap(obj interface{}, structTagName string) resource.PropertyMap {
	v := reflect.ValueOf(obj)
	kind := v.Kind()
	if kind != reflect.Struct {
		panic("type must be struct")
	}
	properties := resource.PropertyMap{}
	for i := 0; i < v.NumField(); i++ {
		fv := v.Field(i)
		fieldName, ok := getTagValue(v.Type().Field(i).Tag, structTagName)
		if !ok {
			continue
		}
		properties[resource.PropertyKey(fieldName)] = resource.NewPropertyValue(get(fv))
	}
	return properties
}

// FromPropertyMap unmarshals properties into out.
func FromPropertyMap(properties resource.PropertyMap, structTagName string, out interface{}) error {
	v := reflect.ValueOf(out)
	kind := v.Kind()
	if kind == reflect.Ptr {
		v = v.Elem()
		kind = v.Kind()
	}
	if kind != reflect.Struct {
		panic("out should be pointer to struct")
	}
	for i := 0; i < v.NumField(); i++ {
		fv := v.Field(i)
		fieldName, ok := getTagValue(v.Type().Field(i).Tag, structTagName)
		if !ok {
			continue
		}
		mapVal, ok := properties[resource.PropertyKey(fieldName)]
		if !ok {
			// skip fields that aren't in property map. callers can validate that fields on out are
			// set properly
			continue
		}
		err := set(fv, mapVal.V)
		if err != nil {
			return err
		}

	}
	return nil
}

// DiffOldsAndNews unmarshals a DiffRequest and runs a diff on them. It returns any keys changed
func DiffOldsAndNews(req *pulumirpc.DiffRequest) ([]string, error) {
	olds, err := plugin.UnmarshalProperties(req.GetOldInputs(), plugin.MarshalOptions{KeepUnknowns: false, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	news, err := plugin.UnmarshalProperties(req.GetNews(), plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: false})
	if err != nil {
		return nil, err
	}
	d := olds.Diff(news)
	var diffs []string
	for _, key := range d.ChangedKeys() {
		diffs = append(diffs, string(key))
	}
	return diffs, nil
}

// get returns the value depending on the kind
func get(v reflect.Value) interface{} {
	switch v.Kind() {
	case reflect.String:
		return v.String()
	case reflect.Int, reflect.Int64, reflect.Int32:
		return v.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint()
	case reflect.Float32, reflect.Float64:
		return v.Float()
	case reflect.Ptr:
		if v.CanAddr() {
			return get(v.Elem())
		}
		return nil
	default:
		return nil
	}
}

// set determines the kind, then sets the value
func set(v reflect.Value, value interface{}) error {
	valueValue := reflect.ValueOf(value)
	valueKind := valueValue.Kind()
	var floatValue *float64
	if valueKind == reflect.Invalid {
		return nil
	}
	if valueKind == reflect.Float64 {
		fv := valueValue.Float()
		floatValue = &fv
	} else if v.Kind() != valueKind {
		return fmt.Errorf("field type %q does not match property %q", v.Kind(), valueKind)
	}
	switch v.Kind() {
	case reflect.String:
		v.SetString(value.(string))
	case reflect.Int, reflect.Int64, reflect.Int32:
		if floatValue != nil {
			v.SetInt(int64(*floatValue))
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v.SetUint(uint64(*floatValue))
	case reflect.Float32, reflect.Float64:
		v.SetFloat(*floatValue)
	case reflect.Ptr:
		// create a new ptr to the type of this pointer. i.e. create string if *string
		v.Set(reflect.New(v.Type().Elem()))
		// call set again, but with deref'd value. note that this will recurse down for ptr to ptr's (and so on)
		return set(v.Elem(), value)
	}
	return nil
}

func getTagValue(tag reflect.StructTag, structTagName string) (string, bool) {
	return tag.Lookup(structTagName)
}

// ConvertMapToPropertyMap converts a map[string]interface{} to resource.PropertyMap
// handling nested structures and arrays properly
func ConvertMapToPropertyMap(m map[string]interface{}) resource.PropertyMap {
	pm := resource.PropertyMap{}
	for k, v := range m {
		pm[resource.PropertyKey(k)] = ConvertInterfaceToPropertyValue(v)
	}
	return pm
}

// ConvertInterfaceToPropertyValue converts interface{} to resource.PropertyValue
// handling arrays, objects, and primitive types recursively
func ConvertInterfaceToPropertyValue(v interface{}) resource.PropertyValue {
	switch val := v.(type) {
	case []interface{}:
		// Handle arrays by recursively converting each element
		arrayValues := make([]resource.PropertyValue, len(val))
		for i, elem := range val {
			arrayValues[i] = ConvertInterfaceToPropertyValue(elem)
		}
		return resource.NewArrayProperty(arrayValues)
	case map[string]interface{}:
		// Handle nested objects by recursively converting
		return resource.NewObjectProperty(ConvertMapToPropertyMap(val))
	default:
		// Handle primitive types (string, number, bool, etc.)
		return resource.NewPropertyValue(val)
	}
}

// ConvertPropertyMapToMap converts resource.PropertyMap to map[string]interface{}
// handling nested structures and arrays properly
func ConvertPropertyMapToMap(pm resource.PropertyMap) map[string]interface{} {
	m := make(map[string]interface{})
	for k, v := range pm {
		m[string(k)] = ConvertPropertyValueToInterface(v)
	}
	return m
}

// ConvertPropertyValueToInterface converts resource.PropertyValue to interface{}
// handling arrays, objects, and primitive types recursively
func ConvertPropertyValueToInterface(pv resource.PropertyValue) interface{} {
	if pv.IsArray() {
		// Handle arrays by recursively converting each element
		arrayValues := pv.ArrayValue()
		result := make([]interface{}, len(arrayValues))
		for i, elem := range arrayValues {
			result[i] = ConvertPropertyValueToInterface(elem)
		}
		return result
	} else if pv.IsObject() {
		// Handle nested objects by recursively converting
		return ConvertPropertyMapToMap(pv.ObjectValue())
	} else {
		// Handle primitive types (string, number, bool, etc.)
		return pv.V
	}
}
