package openapi

import (
	"fmt"
	"reflect"
	"strings"
)

// Use the json: tag to retrieve the name of the prop
func propNameFromField(field reflect.StructField) string {
	name := strings.Split(field.Tag.Get("json"), ",")[0]
	if name == "" {
		name = field.Name
	}
	return name
}

// Bool
func propFromBoolType() FieldProperty {
	return FieldProperty{
		"type": "boolean",
	}
}

// String
func propFromStringType() FieldProperty {
	return FieldProperty{
		"type": "string",
	}
}

// Struct
func propFromStructType(field reflect.StructField, ftype reflect.Type) FieldProperty {
	// This might be a bit unstable but we need to deal with datetime
	name := ftype.Name()
	apiName, ok := field.Tag.Lookup("api")
	if ok {
		name = apiName
	}

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
	elem := ftype.Elem()
	if elem.Kind() == reflect.Pointer {
		elem = elem.Elem()
	}

	var itemProps FieldProperty
	switch elem.Kind() {
	case reflect.String:
		itemProps = FieldProperty{
			"type": "string",
		}
	case reflect.Struct:
		itemProps = FieldProperty{
			"$ref": "#/components/schemas/" + elem.Name(),
		}
	default:
		panic(fmt.Sprintf("unsupported type: %s, %s, %s", ftype, elem, elem.Kind()))
	}

	return FieldProperty{
		"type":  "array",
		"items": itemProps,
	}
}

// Int
func propFromIntType(field reflect.StructField, unsigned bool) FieldProperty {
	p := FieldProperty{
		"type": "integer",
	}
	if unsigned {
		p["minimum"] = 0
	}
	return p
}

func propFromFloatType() FieldProperty {
	return FieldProperty{
		"type": "number",
	}
}

// Map
func propFromMapType(ftype reflect.Type) FieldProperty {
	return FieldProperty{
		"type": "object",
		"additionalProperties": map[string]string{
			"type": "string",
		},
	}
}

// Make property from field. For pointers to other
// objects use a $ref.
func propFromField(field reflect.StructField) FieldProperty {
	ftype := field.Type
	nullable := false
	if field.Type.Kind() == reflect.Pointer {
		ftype = field.Type.Elem()
		nullable = true
	}

	var prop FieldProperty
	switch ftype.Kind() {
	case reflect.Struct:
		prop = propFromStructType(field, ftype)
	case reflect.Int64:
		prop = propFromIntType(field, false)
	case reflect.Uint:
		prop = propFromIntType(field, false)
	case reflect.Uint64:
		prop = propFromIntType(field, false)
	case reflect.Int:
		prop = propFromIntType(field, false)
	case reflect.Float64:
		prop = propFromFloatType()
	case reflect.Bool:
		prop = propFromBoolType()
	case reflect.String:
		prop = propFromStringType()
		if nullable {
			prop["nullable"] = true
		}
	case reflect.Slice:
		prop = propFromSliceType(ftype)
	case reflect.Map:
		prop = propFromMapType(ftype)
	case reflect.Interface:
		prop = propFromMapType(ftype)
	default:
		panic(fmt.Sprintf("unsupported field type: %s, %s", ftype, ftype.Kind()))
	}

	// Add example if present
	example, hasExample := field.Tag.Lookup("example")
	if hasExample {
		prop["example"] = example
	}

	// Add enum
	enum, hasEnum := field.Tag.Lookup("enum")
	if hasEnum {
		prop["enum"] = strings.Split(enum, ",")
	}

	// Add description
	desc, hasDescription := field.Tag.Lookup("doc")
	if hasDescription {
		if hasExample {
			desc += "\n\n**Example**: `" + example + "`"
		}
		prop["description"] = desc
	}

	return prop
}

// generate properties from struct type
func propsFromStruct(obj interface{}) Properties {
	objType := reflect.TypeOf(obj)
	fields := reflect.VisibleFields(objType)
	props := Properties{}
	// Iterate over fields
	for _, field := range fields {
		if !field.IsExported() {
			continue
		}
		if field.Name == "XMLName" {
			continue
		}
		pname := propNameFromField(field)
		if pname == "-" {
			continue
		}
		prop := propFromField(field)
		props[pname] = prop
	}
	return props
}

// PropertiesFrom produces schema properties
func PropertiesFrom(obj interface{}) Properties {
	props := propsFromStruct(obj)
	return props
}

// RequiredFrom retrievs every property
// where omitempty is not set
func RequiredFrom(obj interface{}) []string {
	objType := reflect.TypeOf(obj)
	fields := reflect.VisibleFields(objType)
	// Iterate over fields
	required := []string{}
	for _, field := range fields {
		if !field.IsExported() {
			continue
		}
		if field.Name == "XMLName" {
			continue
		}
		pname := propNameFromField(field)
		if pname == "-" {
			continue
		}
		flags := strings.Split(field.Tag.Get("json"), ",")
		if len(flags) > 1 {
			if flags[1] == "omitempty" {
				continue
			}
		}
		required = append(required, pname)
	}
	return required
}
