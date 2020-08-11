package goutils

import (
	"fmt"
	"reflect"
	"strings"
)

// Params records various parameters required by preload
type Params struct {
	params
	in interface{}
}

type params struct {
	find       func(out interface{}, where ...interface{})
	slice      []interface{}
	elemType   reflect.Type
	primaryKey string
	foreignKey string
}

// New returns a Params containing a structure or slice that needs to be preloaded
func New(in interface{}) *Params {
	params := &Params{}
	params.in = in
	params.find = DefaultFindFunc
	return params
}

// FindFunc specifies the function of preload to obtain the result set
func (p *Params) FindFunc(find func(out interface{}, where ...interface{})) *Params {
	p.find = find
	return p
}

// PrimaryKey specifies the associated field of the parent structure
func (p *Params) PrimaryKey(name string) *Params {
	p.primaryKey = name
	return p
}

// ForeignKey specifies the associated field of the structure that needs to be preloaded
func (p *Params) ForeignKey(name string) *Params {
	p.foreignKey = name
	return p
}

func (p *Params) parseSlice() {
	p.slice = createAnyTypeSlice(p.in)
	if len(p.slice) > 0 {
		p.elemType = getRealType(reflect.TypeOf(p.slice[0]))
	}
}

func (p *params) getFieldIndex(fieldName string) int {
	typ := p.elemType
	for i := 0; i < typ.NumField(); i++ {
		if typ.Field(i).Name == fieldName {
			return i
		}
	}
	panic("this struct has no specify field")
}

func (p *Params) parseKeyByTag(fieldName string) {
	field, _ := p.elemType.FieldByName(fieldName)
	preloadTag := field.Tag.Get("preload")
	if preloadTag == "" {
		panic(fmt.Sprintf("field %s needs to have a tag called preload", fieldName))
	}
	foreignkey := getFieldInString(preloadTag, "foreignkey")
	primarykey := getFieldInString(preloadTag, "primarykey")
	if foreignkey == "" || primarykey == "" {
		panic("preload tag should contain both foreignkey and primarykey like 'foreignkey:SerialNumber;primarykey:SerialNumber'")
	}
	if _, exists := p.elemType.FieldByName(foreignkey); !exists {
		panic(fmt.Sprintf("there is no field named %s in the structure %s", foreignkey, p.elemType.Name()))
	}
	if _, exists := getRealType(field.Type).FieldByName(primarykey); !exists {
		panic(fmt.Sprintf("there is no field named %s in the structure %s", primarykey, getRealType(field.Type).Name()))
	}
	p.foreignKey = foreignkey
	p.primaryKey = primarykey
}

// DefaultFindFunc specifies the load data function used by default globally
var DefaultFindFunc = func(out interface{}, where ...interface{}) {}

// Preload can load the fields with the preload tag of the in structure,
// just like the Preload method of gorm, using only one SQL statement.
// But it can be related by any one-to-one/many-to-one field, not necessarily a foreign key.
// Second param can be slice or struct.
// foreignkey specifies the field to be associated with this structure,
// and primarykey specifies the field to be associated with the structure pointed to by the field with the preload tag.
// For example: Asset *Asset `preload:"foreignkey:SerialNumber;primarykey:SerialNumber"`.
func (p *Params) Preload(fieldName string) {
	p.parseSlice()
	if len(p.slice) == 0 {
		return
	}
	fieldIndex := p.getFieldIndex(fieldName)
	if p.foreignKey == "" || p.primaryKey == "" {
		p.parseKeyByTag(fieldName)
	}
	valuesToBeSearch := make([]interface{}, 0, len(p.slice))
	valuesToBeFilled := make(map[reflect.Value]interface{}, len(p.slice))
	foreignKeyIndex := p.getFieldIndex(p.foreignKey)
	for _, item := range p.slice {
		val := reflect.ValueOf(item).Elem()
		key := val.Field(foreignKeyIndex).Interface()
		field := val.Field(fieldIndex)
		if isZero(val.Field(foreignKeyIndex)) || !isZero(field) {
			continue
		}
		valuesToBeSearch = append(valuesToBeSearch, key)
		valuesToBeFilled[field] = key
	}
	if len(valuesToBeSearch) == 0 {
		return
	}
	fieldType := p.elemType.Field(fieldIndex).Type
	fieldToBeAssociation, _ := getRealType(fieldType).FieldByName(p.primaryKey)
	dbres := reflect.New(reflect.SliceOf(fieldType)).Interface()
	p.find(dbres, getColumnName(&fieldToBeAssociation)+" IN (?)", valuesToBeSearch)
	dbresMap := reflect.ValueOf(ConvertSliceToMap(dbres, p.primaryKey))
	for field, key := range valuesToBeFilled {
		res := dbresMap.MapIndex(reflect.ValueOf(key))
		if res.IsValid() {
			field.Set(res)
		}
	}
}

func isZero(val reflect.Value) bool {
	val = getRealValue(val)
	switch val.Kind() {
	case reflect.Struct:
		return val.IsNil()
	case reflect.String:
		return val.String() == ""
	default:
		return !val.IsValid()
	}
}

func getColumnName(field *reflect.StructField) string {
	if columnName := getFieldInString(field.Tag.Get("gorm"), "column"); columnName != "" {
		return columnName
	}
	return CamelCase(strings.ReplaceAll(field.Name, "ID", "Id"))
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
