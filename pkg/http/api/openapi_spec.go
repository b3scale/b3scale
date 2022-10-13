package api

import (
	oa "github.com/b3scale/b3scale/pkg/openapi"
)

// NewErrorSchema creates the error schema
func NewErrorSchema() oa.Schema {
	return oa.Schema{
		"description": "A general error response object.",
		"type":        "object",
		"properties": oa.Properties{
			"error": oa.Property{
				Type: "string",
				Description: "An error type tag. " +
					"See specific error for details.",
			},
			"message": oa.Property{
				Type: "string",
				Description: "A human readable message with details " +
					"about the error.",
			},
		},
	}
}

// NewNotFoundErrorSchema creates a not found error object
// schema
func NewNotFoundErrorSchema() oa.Schema {
	return oa.Schema{
		"description": "The requested resource could not be found",
		"type":        "object",
		"allOf": []interface{}{
			oa.SchemaRef{
				Ref: "#/components/schemas/Error",
			},
		},
	}
}

// NewValidationErrorSchema creates the validation error schema
func NewValidationErrorSchema() oa.Schema {
	return oa.Schema{
		"type": "object",
		"description": "Request validation failed." +
			"The error type is: `validation_error`",
		"allOf": []interface{}{
			oa.SchemaRef{
				Ref: "#/components/schemas/Error",
			},
		},
	}
}

// NewAPISchemas creates the schemas we use in the API
func NewAPISchemas() map[string]oa.Schema {
	return map[string]oa.Schema{
		"Error":           NewErrorSchema(),
		"NotFoundError":   NewNotFoundErrorSchema(),
		"ValidationError": NewValidationErrorSchema(),
	}
}

// NewAPIResponses creates all default (error) responses
func NewAPIResponses() map[string]oa.Response {
	return map[string]oa.Response{
		"NotFoundError": oa.Response{
			Description: "The requested resource could not be found",
			Content: map[string]oa.MediaType{
				oa.ApplicationJSON: oa.MediaType{
					Schema: oa.SchemaRef{
						Ref: "#/components/schemas/NotFoundError",
					},
				},
			},
		},
	}
}

// NewAPISecuritySchemes creates the default security schemes
func NewAPISecuritySchemes() map[string]oa.SecurityScheme {
	return map[string]oa.SecurityScheme{
		"jwt_bearer": oa.SecurityScheme{
			Type:         "bearer",
			Description:  "Authorization using a JWT bearer token",
			BearerFormat: "JWT",
			In:           "header",
			Name:         "JWT bearer token",
		},
	}
}

// NewAPISpec makes the api specs for b3scale
func NewAPISpec() *oa.Spec {
	return &oa.Spec{
		OpenAPI: "3.0.1",
		Info: oa.Info{
			Title:       "b3scale api v1",
			Description: "This document describes the OpenAPI specifications for the b3scale API v1",
			Version:     "1.0.0",
			License: oa.License{
				Name: "Apache 2.0",
				URL:  "https://www.apache.org/licenses/LICENSE-2.0.html",
			},
		},
		Components: oa.Components{
			Responses:       NewAPIResponses(),
			Schemas:         NewAPISchemas(),
			SecuritySchemes: NewAPISecuritySchemes(),
		},
		Security: []oa.SecuritySpec{
			{
				"jwt_bearer": []interface{}{},
			},
		},
	}
}
