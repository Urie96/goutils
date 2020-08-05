package goutils

import (
	"fmt"
	"reflect"
	"strings"
)

// Preload can load the fields with the preload tag of the in structure,
// just like the Preload method of gorm, using only one SQL statement.
// But it can be related by any one-to-one/many-to-one field, not necessarily a foreign key.
// Second param can be slice or struct.
// foreignkey specifies the field to be associated with this structure,
// and primarykey specifies the field to be associated with the structure pointed to by the field with the preload tag.
// For example: Asset *Asset `preload:"foreignkey:SerialNumber;primarykey:SerialNumber"`.
func Preload(find func(out interface{}, where ...interface{}), in interface{}) {
	slice := createAnyTypeSlice(in)
	if len(slice) < 1 {
		return
	}
	typ := getRealType(reflect.TypeOf(slice[0]))
	for i := 0; i < typ.NumField(); i++ {
		preloadTag := typ.Field(i).Tag.Get("preload")
		if preloadTag == "" {
			continue
		}
		foreignkey := getFieldInString(preloadTag, "foreignkey")
		if _, exists := typ.FieldByName(foreignkey); !exists {
			panic(fmt.Sprintf("There is no field named %s in the structure %s", foreignkey, typ.Name()))
		}
		primarykey := getFieldInString(preloadTag, "primarykey")
		if _, exists := getRealType(typ.Field(i).Type).FieldByName(primarykey); !exists {
			panic(fmt.Sprintf("There is no field named %s in the structure %s", primarykey, getRealType(typ.Field(i).Type).Name()))
		}
		if foreignkey == "" || primarykey == "" {
			panic("preload tag should contain both foreignkey and primarykey like 'foreignkey:SerialNumber;primarykey:SerialNumber'")
		}
		keyValue := make(map[interface{}]reflect.Value)
		valuesToBeSearch := make([]interface{}, len(slice))
		for index, item := range slice {
			val := reflect.ValueOf(item).Elem()
			key := val.FieldByName(foreignkey).Interface()
			valuesToBeSearch[index] = key
			keyValue[key] = val.Field(i)
		}
		fieldToBeAssociation, _ := typ.Field(i).Type.Elem().FieldByName(primarykey)
		dbres := reflect.New(reflect.SliceOf(typ.Field(i).Type)).Interface()
		find(dbres, getColumnName(&fieldToBeAssociation)+" in (?)", valuesToBeSearch)
		dbresSlice := createAnyTypeSlice(dbres)
		for _, item := range dbresSlice {
			val := reflect.ValueOf(item)
			key := val.Elem().FieldByName(foreignkey).Interface()
			keyValue[key].Set(val)
		}
	}
}

func getColumnName(field *reflect.StructField) string {
	if columnName := getFieldInString(field.Tag.Get("gorm"), "column"); columnName != "" {
		return columnName
	}
	return CamelCase(field.Name)
}

func getFieldInString(s, substr string) string {
	substr += ":"
	ss := strings.Split(s, ";")
	for _, v := range ss {
		if strings.HasPrefix(v, substr) {
			return strings.Trim(v[len(substr):], " ")
		}
	}
	return ""
}

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
