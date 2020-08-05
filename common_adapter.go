package goutils

import (
	"reflect"
	"strings"
)

// MapTo can map the structure or slice or map to another different type
func MapTo(in, out interface{}) interface{} {
	rType := reflect.TypeOf(in).Elem()
	switch rType.Kind() {
	case reflect.Struct:
		mapToStruct(in, out)
	case reflect.Slice:
		mapToSlice(in, out)
	case reflect.Map:
		mapToMap(in, out)
	default:
		panic("interface{} type should be *struct or *[]*strcut or *map[string]*struct")
	}
	return out
}

func mapToMap(in, out interface{}) {
	iv := reflect.ValueOf(in).Elem()
	ov := reflect.ValueOf(out).Elem()
	ov.Set(reflect.MakeMap(reflect.TypeOf(out).Elem()))
	for _, key := range iv.MapKeys() {
		mapValue := iv.MapIndex(key)
		elem := reflect.New(reflect.TypeOf(out).Elem().Elem().Elem())
		mapToStruct(mapValue.Interface(), elem.Interface())
		ov.SetMapIndex(key, elem)
	}
}

func mapToStruct(in, out interface{}) {
	v := reflect.ValueOf(in).Elem()
	fields := make(map[string]reflect.Value)
	for i := 0; i < v.NumField(); i++ {
		fieldInfo := v.Type().Field(i)
		fieldNameLower := strings.ToLower(fieldInfo.Name)
		fields[fieldNameLower] = v.Field(i)
	}

	o := reflect.ValueOf(out).Elem()
	for i := 0; i < o.NumField(); i++ {
		outField := o.Field(i)
		fieldInfo := o.Type().Field(i)
		fieldNameLower := strings.ToLower(fieldInfo.Name)
		if inField, ok := fields[fieldNameLower]; ok {
			populate(outField, inField)
		}
	}
}

func populate(field, data reflect.Value) {
	if field.Kind() == reflect.Ptr && data.Kind() == reflect.Ptr {
		field.Set(reflect.New(field.Type().Elem()))
		mapToStruct(data.Interface(), field.Interface())
	} else {
		field.Set(data)
	}
}

func mapToSlice(in, out interface{}) {
	inList := createAnyTypeSlice(in)
	o := reflect.ValueOf(out).Elem()
	for _, inItem := range inList {
		elem := reflect.New(o.Type().Elem().Elem())
		mapToStruct(inItem, elem.Interface())
		o.Set(reflect.Append(o, elem))
	}
}
