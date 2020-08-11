package goutils

import "reflect"

func MakeSliceFilledWithStructFields(in interface{}, fieldName string) interface{} {
	slice := createAnyTypeSlice(in)
	typ := getRealType(reflect.TypeOf(slice[0]))
	fieldIndex := getFieldIndex(typ, fieldName)
	vals := make([]reflect.Value, len(slice))
	for i, item := range slice {
		vals[i] = getRealValue(reflect.ValueOf(item)).Field(fieldIndex)
	}
	resValue := reflect.New(reflect.SliceOf(typ.Field(fieldIndex).Type))
	return reflect.Append(resValue.Elem(), vals...).Interface()
}

func ConvertSliceToMap(in interface{}, fieldName string) interface{} {
	slice := createAnyTypeSlice(in)
	typ := getRealType(reflect.TypeOf(slice[0]))
	fieldIndex := getFieldIndex(typ, fieldName)
	keyType := typ.Field(fieldIndex).Type
	resValue := reflect.MakeMap(reflect.MapOf(keyType, reflect.TypeOf(slice[0])))
	for _, elem := range slice {
		val := reflect.ValueOf(elem)
		rval := getRealValue(val)
		resValue.SetMapIndex(rval.Field(fieldIndex), val)
	}
	return resValue.Interface()
}
