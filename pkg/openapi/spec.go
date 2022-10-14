package openapi

// ApplicationJSON is a media type
const ApplicationJSON = "application/json"

// Schema defines an api object
type Schema map[string]interface{}

// Patch rewrites a reference for a Patch operation
func (s Schema) Patch(prop string) Schema {
	current := s["properties"].(Properties)
	schema := current[prop].(FieldProperty)
	ref := schema["$ref"].(string)
	schema["$ref"] = ref + "Patch"
	current[prop] = schema
	return s
}

// Only creates a new object schema with properties filtered
func (s Schema) Only(props ...string) Schema {
	current := s["properties"].(Properties)
	next := Properties{}
	for _, name := range props {
		next[name] = current[name]
	}
	s["properties"] = next

	return s
}

// Require will mark properties as required
func (s Schema) Require(props ...string) Schema {
	s["required"] = props
	return s
}

// RequireFrom sets required field from an object
func (s Schema) RequireFrom(obj interface{}) Schema {
	s["required"] = RequiredFrom(obj)
	return s
}

// ArraySchema creates a schema for an array type
func ArraySchema(description string, items interface{}) Schema {
	return Schema{
		"type":        "array",
		"description": description,
		"items":       items,
	}
}

// ObjectSchema creates a new schema from an object
func ObjectSchema(description string, obj interface{}) Schema {
	props := PropertiesFrom(obj)
	return Schema{
		"type":        "object",
		"description": description,
		"properties":  props,
	}
}

// Properties is a key value mapping from string to property
type Properties map[string]interface{}

// FieldProperty is a free form property
type FieldProperty map[string]interface{}

// Property of an object
type Property struct {
	Type        string `json:"type"`
	Format      string `json:"format,omitempty"`
	Description string `json:"description"`
}

// Ref is a reference  within the openapi document
type Ref struct {
	Ref string `json:"$ref"`
}

// SchemaRef creates a new ref to a schema
func SchemaRef(schema string) Ref {
	return Ref{
		Ref: "#/components/schemas/" + schema,
	}
}

// MediaType is the body of a respones object, e.g. a JSON payload
// with a schema. All schemas here are referenced.
type MediaType struct {
	Schema interface{} `json:"schema"`
}

// Response encode a mapping between a status code and a
// media type object
type Response struct {
	Description string               `json:"description"`
	Content     map[string]MediaType `json:"content,omitempty"`
}

// ResponseRef is a reference to a response defined
// in components
func ResponseRef(response string) Ref {
	return Ref{
		Ref: "#/components/responses/" + response,
	}
}

// ResponseRefs is a mapping from status code to
// a response reference
type ResponseRefs map[string]Ref

// Request describes a request body
type Request struct {
	Content map[string]MediaType `json:"content"`
}

// An Operation describes an api endpoint in a mapping
// of HTTP verb to sdescription
type Operation struct {
	Description string       `json:"description"`
	Responses   ResponseRefs `json:"responses"`
	OperationID string       `json:"operationId"`
	Parameters  []Schema     `json:"parameters,omitempty"`
	RequestBody *Request     `json:"requestBody,omitempty"`
	Summary     string       `json:"summary,omitempty"`
	Tags        []string     `json:"tags,omitempty"`
}

// SecurityScheme describes a security scheme
type SecurityScheme struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	Name        string `json:"name"`
	In          string `json:"in"` // query, header cookie
	// Scheme       string `json:"scheme"` // RFC7235
	BearerFormat string `json:"bearerFormat"`
}

// Components describe the components object
type Components struct {
	Schemas   map[string]Schema   `json:"schemas"`
	Responses map[string]Response `json:"responses"`
	// Parameters      map[string]Parameter      `json:"parameters"`
	// Examples        map[string]Example        `json:"examples"`
	// RequestBodies   map[string]RequestBody    `json:"requestBodies"`
	// Headers         map[string]Headers        `json:"headers"`
	SecuritySchemes map[string]SecurityScheme `json:"securitySchemes"`
}

// Tag object for metadata
type Tag struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// Server is a description of an API server
type Server struct {
	URL         string `json:"url"`
	Description string `json:"description,omitempty"`
}

// SecuritySpec is a polymorphic type
type SecuritySpec map[string][]interface{}

// Security describes the auth methods of the API
type Security []SecuritySpec

// Path is a mapping of http verb to response
type Path map[string]interface{}

/*
type Path map[string]struct {
	Description string                  `json:"description"`
	Responses   map[string]ResponseRefs `json:"responses"`
}
*/

// License information
type License struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

// Info about the API
type Info struct {
	Title       string  `json:"title"`
	Version     string  `json:"version"`
	License     License `json:"license"`
	Description string  `json:"description"`
}

// Spec is the OpenAPI document describing the b3scale API
type Spec struct {
	OpenAPI    string          `json:"openapi"` // Version
	Info       Info            `json:"info"`
	Paths      map[string]Path `json:"paths"`
	Components Components      `json:"components"`
	Tags       []Tag           `json:"tags"`
	Servers    []Server        `json:"servers,omitempty"`
	Security   Security        `json:"security"`
}

// ParamID creates an 'id' path parameter
func ParamID() Schema {
	return Schema{
		"name":        "id",
		"in":          "path",
		"description": "the identifier of the object",
		"required":    true,
		"schema": Schema{
			"type": "string",
		},
	}
}
