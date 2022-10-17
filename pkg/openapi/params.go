package openapi

// ParamID creates an 'id' path parameter
func ParamID() Schema {
	return Schema{
		"name":        "id",
		"in":          "path",
		"description": "The identifier of the object.",
		"required":    true,
		"schema": Schema{
			"type": "string",
		},
	}
}

// ParamQuery creates a query parameter
func ParamQuery(name, description string) Schema {
	return Schema{
		"name":        name,
		"in":          "query",
		"description": description,
		"required":    false,
		"schema": Schema{
			"type": "string",
		},
	}
}
