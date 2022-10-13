package openapi

import (
	"fmt"
	"reflect"
	"strings"
)

// Use the json: tag to retrieve the name of the prop
func propNameFromField(field reflect.StructField) string {
	name := strings.Split(field.Tag.Get("json"), " ")[0]
	return name
}

// Bool
func propFromBoolType() FieldProperty {
	return FieldProperty{
		"type": "bool",
	}
}

// String
func propFromStringType() FieldProperty {
	return FieldProperty{
		"type": "string",
	}
}

// Struct
func propFromStructType(ftype reflect.Type) FieldProperty {
	// This might be a bit unstable but we need to deal with datetime
	name := ftype.Name()
	pkg := ftype.PkgPath()
	if pkg == "time" && name == "Time" {
		return FieldProperty{
			"type":   "string",
			"format": "date-time",
		}
	}

	// Ref
	return FieldProperty{
		"$ref": "#/components/schemas/" + name,
	}
}

// Slice
func propFromSliceType(ftype reflect.Type) FieldProperty {
	return FieldProperty{
		"type": "array",
		"items": map[string]string{
			"$ref": "#/components/schemas/" + "FIIIEL",
		},
	}
}

// Map
func propFromMapType(ftype reflect.Type) FieldProperty {
	return FieldProperty{
		"type": "hmm",
	}
}

// Make property from field. For pointers to other
// objects use a $ref.
func propFromField(field reflect.StructField) FieldProperty {
	ftype := field.Type
	if field.Type.Kind() == reflect.Pointer {
		ftype = field.Type.Elem()
	}

	var prop FieldProperty
	switch ftype.Kind() {
	case reflect.Struct:
		prop = propFromStructType(ftype)
	case reflect.Bool:
		prop = propFromBoolType()
	case reflect.String:
		prop = propFromStringType()
	case reflect.Slice:
		prop = propFromSliceType(ftype)
	case reflect.Map:
		prop = propFromMapType(ftype)
	default:
		panic(fmt.Sprintf("unsupported type: %s, %s", ftype, ftype.Kind()))
	}

	// Add description
	desc, ok := field.Tag.Lookup("doc")
	if ok {
		prop["description"] = desc
	}

	return prop
}

// PropertiesFromObject produces schema properties
func PropertiesFromObject(obj interface{}) Properties {
	objType := reflect.TypeOf(obj)
	fields := reflect.VisibleFields(objType)

	// Fields which need to be marked as required
	props := Properties{}

	// Iterate over fields
	for _, field := range fields {
		prop := propFromField(field)
		pname := propNameFromField(field)
		props[pname] = prop
	}

	return props
}
