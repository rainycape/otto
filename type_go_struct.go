package otto

import (
	"encoding/json"
	"reflect"
	"unicode"
	"unicode/utf8"
)

// FIXME Make a note about not being able to modify a struct unless it was
// passed as a pointer-to: &struct{ ... }
// This seems to be a limitation of the reflect package.
// This goes for the other Go constructs too.
// I guess we could get around it by either:
// 1. Creating a new struct every time
// 2. Creating an addressable? struct in the constructor

func (runtime *_runtime) newGoStructObject(value reflect.Value) *_object {
	self := runtime.newObject()
	self.class = "Object" // TODO Should this be something else?
	self.objectClass = _classGoStruct
	self.value = _newGoStructObject(value)
	return self
}

type _goStructObject struct {
	value reflect.Value
}

func _newGoStructObject(value reflect.Value) *_goStructObject {
	if reflect.Indirect(value).Kind() != reflect.Struct {
		dbgf("%/panic//%@: %v != reflect.Struct", value.Kind())
	}
	self := &_goStructObject{
		value: value,
	}
	return self
}

func (self _goStructObject) getValue(name string) reflect.Value {
	name = toGoName(name)
	if field := reflect.Indirect(self.value).FieldByName(name); field.IsValid() {
		return field
	}

	if method := self.value.MethodByName(name); method.IsValid() {
		return method
	}

	return reflect.Value{}
}

func (self _goStructObject) field(name string) (reflect.StructField, bool) {
	return reflect.Indirect(self.value).Type().FieldByName(toGoName(name))
}

func (self _goStructObject) method(name string) (reflect.Method, bool) {
	return reflect.Indirect(self.value).Type().MethodByName(toGoName(name))
}

func (self _goStructObject) setValue(name string, value Value) bool {
	field, exists := self.field(name)
	if !exists {
		return false
	}
	fieldValue := self.getValue(name)
	reflectValue, err := value.toReflectValue(field.Type.Kind())
	if err != nil {
		panic(err)
	}
	fieldValue.Set(reflectValue)

	return true
}

func goStructGetOwnProperty(self *_object, name string) *_property {
	object := self.value.(*_goStructObject)
	value := object.getValue(name)
	if value.IsValid() {
		return &_property{self.runtime.toValue(value.Interface()), 0110}
	}

	return objectGetOwnProperty(self, name)
}

func toJavascriptName(name string) string {
	if name != "" {
		runes := []rune(name)
		runes[0] = unicode.ToLower(runes[0])
		name = string(runes)
	}
	return name
}

func toGoName(name string) string {
	if name != "" {
		runes := []rune(name)
		runes[0] = unicode.ToUpper(runes[0])
		name = string(runes)
	}
	return name
}

func isExported(name string) bool {
	if name != "" {
		r, _ := utf8.DecodeRuneInString(name)
		return r != utf8.RuneError && unicode.IsUpper(r)
	}
	return false
}

func goStructEnumerate(self *_object, all bool, each func(string) bool) {
	object := self.value.(*_goStructObject)

	// Enumerate fields
	for index := 0; index < reflect.Indirect(object.value).NumField(); index++ {
		name := reflect.Indirect(object.value).Type().Field(index).Name
		if isExported(name) {
			if !each(toJavascriptName(name)) {
				return
			}
		}
	}

	// Enumerate methods
	for index := 0; index < object.value.NumMethod(); index++ {
		name := object.value.Type().Method(index).Name
		if isExported(name) {
			if !each(toJavascriptName(name)) {
				return
			}
		}
	}

	objectEnumerate(self, all, each)
}

func goStructCanPut(self *_object, name string) bool {
	object := self.value.(*_goStructObject)
	value := object.getValue(name)
	if value.IsValid() {
		return true
	}

	return objectCanPut(self, name)
}

func goStructPut(self *_object, name string, value Value, throw bool) {
	object := self.value.(*_goStructObject)
	if object.setValue(toGoName(name), value) {
		return
	}

	objectPut(self, name, value, throw)
}

func goStructMarshalJSON(self *_object) json.Marshaler {
	object := self.value.(*_goStructObject)
	goValue := reflect.Indirect(object.value).Interface()
	switch marshaler := goValue.(type) {
	case json.Marshaler:
		return marshaler
	}
	return nil
}
