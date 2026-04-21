package json

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

// Unmarshal parses the JSON-encoded data and stores the result
// in the value pointed to by v.
// AI-GENERATED
func Unmarshal(data []byte, v interface{}) error {
	tokenizer := newTokenizer(string(data))
	tokens, err := tokenizer.tokenize()
	if err != nil {
		return err
	}
	parser := newParser(tokens)
	parsedData, err := parser.parse()
	if err != nil {
		return err
	}

	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return errors.New("json: Unmarshal(non-pointer or nil)")
	}
	return decodeValue(parsedData, rv.Elem())
}

func decodeValue(data interface{}, v reflect.Value) error {
	if data == nil {
		return nil
	}

	if v.Kind() == reflect.Interface {
		v.Set(reflect.ValueOf(data))
		return nil
	}

	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		return decodeValue(data, v.Elem())
	}

	switch d := data.(type) {
	case map[string]interface{}:
		if v.Kind() == reflect.Struct {
			for key, val := range d {
				var field reflect.Value
				typ := v.Type()
				for i := 0; i < typ.NumField(); i++ {
					f := typ.Field(i)
					tag := f.Tag.Get("json")
					tagName := strings.Split(tag, ",")[0]
					if tagName == key || (tagName == "" && strings.EqualFold(f.Name, key)) {
						field = v.Field(i)
						break
					}
				}
				if field.IsValid() && field.CanSet() {
					if err := decodeValue(val, field); err != nil {
						return err
					}
				}
			}
		} else if v.Kind() == reflect.Map {
			if v.IsNil() {
				v.Set(reflect.MakeMap(v.Type()))
			}
			for key, val := range d {
				mapKey := reflect.ValueOf(key)
				mapVal := reflect.New(v.Type().Elem()).Elem()
				if err := decodeValue(val, mapVal); err != nil {
					return err
				}
				v.SetMapIndex(mapKey, mapVal)
			}
		} else {
			return fmt.Errorf("cannot decode map into %v", v.Type())
		}
	case []interface{}:
		if v.Kind() == reflect.Slice {
			slice := reflect.MakeSlice(v.Type(), len(d), len(d))
			for i, val := range d {
				if err := decodeValue(val, slice.Index(i)); err != nil {
					return err
				}
			}
			v.Set(slice)
		} else {
			return fmt.Errorf("cannot decode array into %v", v.Type())
		}
	case string:
		if v.Kind() == reflect.String {
			v.SetString(d)
		} else {
			return fmt.Errorf("cannot decode string into %v", v.Type())
		}
	case int64:
		if v.Kind() >= reflect.Int && v.Kind() <= reflect.Int64 {
			v.SetInt(d)
		} else if v.Kind() >= reflect.Uint && v.Kind() <= reflect.Uint64 {
			v.SetUint(uint64(d))
		} else if v.Kind() >= reflect.Float32 && v.Kind() <= reflect.Float64 {
			v.SetFloat(float64(d))
		} else {
			return fmt.Errorf("cannot decode int64 into %v", v.Type())
		}
	case float64:
		if v.Kind() >= reflect.Float32 && v.Kind() <= reflect.Float64 {
			v.SetFloat(d)
		} else if v.Kind() >= reflect.Int && v.Kind() <= reflect.Int64 {
			v.SetInt(int64(d))
		} else {
			return fmt.Errorf("cannot decode float64 into %v", v.Type())
		}
	case bool:
		if v.Kind() == reflect.Bool {
			v.SetBool(d)
		} else {
			return fmt.Errorf("cannot decode bool into %v", v.Type())
		}
	default:
		return fmt.Errorf("unknown type: %T", d)
	}
	return nil
}
