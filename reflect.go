package goutils

import "reflect"

func createAnyTypeSlice(in interface{}) []interface{} {
	val := getRealValue(reflect.ValueOf(in))
	var out []interface{}
	switch val.Type().Kind() {
	case reflect.Slice, reflect.Array:
		out = make([]interface{}, val.Len())
		for i := 0; i < val.Len(); i++ {
			out[i] = val.Index(i).Interface()
		}
		return out
	default:
		out = append(out, in)
	}
	return out
}

func getRealType(typ reflect.Type) reflect.Type {
	for typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	return typ
}

func getRealValue(val reflect.Value) reflect.Value {
	for val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	return val
}

func getFieldIndex(typ reflect.Type, fieldName string) int {
	for i := 0; i < typ.NumField(); i++ {
		if typ.Field(i).Name == fieldName {
			return i
		}
	}
	panic("this struct has no specify field")
}
