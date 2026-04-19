package util

import (
	"fmt"
	"reflect"

	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
)

const structTagName = "pulumi"

// ToPropertyMap marshals a struct into a resource.PropertyMap. It obtains the
// resource.PropertyKey() values for each struct field by grabbing value of the
// structTagName.
//
// Map fields are only supported when typed as map[string]string; fields typed
// with any other key or value kind are emitted as a null property. Use
// FromPropertyMap to decode a map[string]string field; unsupported map types
// return an error there.
func ToPropertyMap(obj any) resource.PropertyMap {
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
func FromPropertyMap(properties resource.PropertyMap, out any) error {
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
	olds, err := plugin.UnmarshalProperties(
		req.GetOldInputs(),
		plugin.MarshalOptions{KeepUnknowns: false, SkipNulls: true},
	)
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
	case reflect.Map:
		// Handle map[string]string
		if v.Type().Key().Kind() == reflect.String && v.Type().Elem().Kind() == reflect.String {
			result := make(map[string]string, v.Len())
			for _, key := range v.MapKeys() {
				result[key.String()] = v.MapIndex(key).String()
			}
			return result
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
		// Allow map-to-map assignment even when concrete element types differ
		// (e.g. map[string]string target vs. map[string]interface{} source).
		if v.Kind() != reflect.Map || valueKind != reflect.Map {
			return fmt.Errorf("field type %q does not match property %q", v.Kind(), valueKind)
		}
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
	case reflect.Map:
		if v.Type().Key().Kind() != reflect.String || v.Type().Elem().Kind() != reflect.String {
			return fmt.Errorf("unsupported map type %s: only map[string]string is supported", v.Type())
		}
		pm, ok := value.(resource.PropertyMap)
		if !ok {
			return fmt.Errorf("expected resource.PropertyMap, got %T", value)
		}
		// Always start from a fresh map so keys that are absent from `value`
		// don't linger from a prior decode into the same struct.
		v.Set(reflect.MakeMapWithSize(v.Type(), len(pm)))
		for k, pv := range pm {
			if !pv.IsString() {
				return fmt.Errorf("expected string value for key %q, got %T", string(k), pv.V)
			}
			v.SetMapIndex(reflect.ValueOf(string(k)), reflect.ValueOf(pv.StringValue()))
		}
	}
	return nil
}

func getTagValue(tag reflect.StructTag, structTagName string) (string, bool) {
	return tag.Lookup(structTagName)
}
