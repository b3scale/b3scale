package api

import (
	"github.com/b3scale/b3scale/pkg/bbb"
	oa "github.com/b3scale/b3scale/pkg/openapi"
	"github.com/b3scale/b3scale/pkg/store"
)

// NewAPIEndpointsSchema generates all the endpoints for the schema
func NewAPIEndpointsSchema() map[string]oa.Path {
	return map[string]oa.Path{
		"/frontends": oa.Path{
			"get": oa.Operation{
				Description: "Fetch all frontends",
				OperationID: "frontendsList",
				Summary:     "List",
				Tags:        []string{"Frontends"},
				Responses: oa.ResponseRefs{
					"200": oa.ResponseRef("Frontends"),
				},
			},
			"post": oa.Operation{
				Description: "Register a new frontend",
				OperationID: "frontendsCreate",
				Summary:     "Create",
				Tags:        []string{"Frontends"},
				RequestBody: &oa.Request{
					Content: map[string]oa.MediaType{
						oa.ApplicationJSON: oa.MediaType{
							Schema: oa.SchemaRef("FrontendRequest"),
						},
					},
				},
				Responses: oa.ResponseRefs{
					"200": oa.ResponseRef("Frontend"),
				},
			},
		},
		"/frontends/{id}": oa.Path{
			"parameters": []oa.Schema{
				oa.ParamID(),
			},
			"get": oa.Operation{
				Description: "Fetch a single frontend identified by ID",
				OperationID: "frontendsRead",
				Summary:     "Read",
				Tags:        []string{"Frontends"},
				Responses: oa.ResponseRefs{
					"200": oa.ResponseRef("Frontend"),
				},
			},
			"patch": oa.Operation{
				Description: "Update parts of a frontend",
				OperationID: "frontendsPatch",
				Summary:     "Update",
				Tags:        []string{"Frontends"},
				RequestBody: &oa.Request{
					Content: map[string]oa.MediaType{
						oa.ApplicationJSON: oa.MediaType{
							Schema: oa.SchemaRef("FrontendPatch"),
						},
					},
				},
				Responses: oa.ResponseRefs{
					"200": oa.ResponseRef("Frontend"),
				},
			},
			"delete": oa.Operation{
				Description: "Remove a frontend",
				OperationID: "frontendsDestroy",
				Summary:     "Delete",
				Tags:        []string{"Frontends"},
				Responses: oa.ResponseRefs{
					"200": oa.ResponseRef("Frontend"),
				},
			},
		},

		"/backends": oa.Path{
			"get": oa.Operation{
				Description: "Fetch all backends",
				OperationID: "frontendsList",
				Summary:     "List",
				Tags:        []string{"Backends"},
				Responses: oa.ResponseRefs{
					"200": oa.ResponseRef("Backends"),
				},
			},
			"post": oa.Operation{
				Description: "Register a new backend",
				OperationID: "backendsCreate",
				Summary:     "Create",
				Tags:        []string{"Backends"},
				RequestBody: &oa.Request{
					Content: map[string]oa.MediaType{
						oa.ApplicationJSON: oa.MediaType{
							Schema: oa.SchemaRef("BackendRequest"),
						},
					},
				},
				Responses: oa.ResponseRefs{
					"200": oa.ResponseRef("Backend"),
				},
			},
		},
		"/backends/{id}": oa.Path{
			"parameters": []oa.Schema{
				oa.ParamID(),
			},
			"get": oa.Operation{
				Description: "Fetch a single backend identified by ID",
				OperationID: "backendsRead",
				Summary:     "Read",
				Tags:        []string{"Backends"},
				Responses: oa.ResponseRefs{
					"200": oa.ResponseRef("Backend"),
				},
			},
			"patch": oa.Operation{
				Description: "Update parts of a backend",
				OperationID: "backendsPatch",
				Summary:     "Update",
				Tags:        []string{"Backends"},
				RequestBody: &oa.Request{
					Content: map[string]oa.MediaType{
						oa.ApplicationJSON: oa.MediaType{
							Schema: oa.SchemaRef("BackendPatch"),
						},
					},
				},
				Responses: oa.ResponseRefs{
					"200": oa.ResponseRef("Backend"),
				},
			},
			"delete": oa.Operation{
				Description: "Remove a backend",
				OperationID: "backendsDestroy",
				Summary:     "Delete",
				Tags:        []string{"Backends"},
				Responses: oa.ResponseRefs{
					"200": oa.ResponseRef("Backend"),
				},
			},
		},
	}
}

// NewAPIResponses creates all default (error) responses
func NewAPIResponses() map[string]oa.Response {
	return map[string]oa.Response{
		"Frontends": oa.Response{
			Description: "List Of Frontends",
			Content: map[string]oa.MediaType{
				oa.ApplicationJSON: oa.MediaType{
					Schema: oa.SchemaRef("Frontends"),
				},
			},
		},
		"Frontend": oa.Response{
			Description: "Frontend",
			Content: map[string]oa.MediaType{
				oa.ApplicationJSON: oa.MediaType{
					Schema: oa.SchemaRef("Frontend"),
				},
			},
		},
		"Backends": oa.Response{
			Description: "List Of Backends",
			Content: map[string]oa.MediaType{
				oa.ApplicationJSON: oa.MediaType{
					Schema: oa.SchemaRef("Backends"),
				},
			},
		},
		"Backend": oa.Response{
			Description: "Backend",
			Content: map[string]oa.MediaType{
				oa.ApplicationJSON: oa.MediaType{
					Schema: oa.SchemaRef("Backend"),
				},
			},
		},
		"NotFoundError": oa.Response{
			Description: "The requested resource could not be found",
			Content: map[string]oa.MediaType{
				oa.ApplicationJSON: oa.MediaType{
					Schema: oa.SchemaRef("NotFoundError"),
				},
			},
		},
	}
}

// NewAPISchemas creates the schemas we use in the API
func NewAPISchemas() map[string]oa.Schema {
	return map[string]oa.Schema{
		"Frontends": oa.ArraySchema(
			"A list of frontends",
			oa.SchemaRef("Frontend")),
		"FrontendRequest": oa.ObjectSchema(
			"A frontend request",
			store.FrontendState{}).
			Only("active", "bbb", "settings", "account_ref").
			Require("bbb"),
		"FrontendPatch": oa.ObjectSchema(
			"A frontend update",
			store.FrontendState{}).
			Only("active", "bbb", "settings", "account_ref").
			Patch("bbb"),
		"Frontend": oa.ObjectSchema(
			"A frontend",
			store.FrontendState{}).
			RequireFrom(store.FrontendState{}),
		"FrontendConfig": oa.ObjectSchema(
			"A BBB frontend configuration",
			bbb.Frontend{}).
			RequireFrom(bbb.Frontend{}),
		"FrontendConfigPatch": oa.ObjectSchema(
			"A BBB frontend configuration",
			bbb.Frontend{}),
		"FrontendSettings": oa.ObjectSchema(
			"Frontend settings",
			store.FrontendSettings{}).
			RequireFrom(store.FrontendSettings{}),
		"DefaultPresentationSettings": oa.ObjectSchema(
			"Default Presentation",
			store.DefaultPresentationSettings{}).
			RequireFrom(store.DefaultPresentationSettings{}),

		"Backends": oa.ArraySchema(
			"List Of Backends",
			oa.SchemaRef("Backend")),
		"BackendRequest": oa.ObjectSchema(
			"Backend Request",
			store.BackendState{}).
			Require("bbb").
			Only("admin_state", "bbb", "settings", "load_factor"),
		"BackendPatch": oa.ObjectSchema(
			"Backend Update",
			store.BackendState{}),
		"Backend": oa.ObjectSchema(
			"Backend",
			store.BackendState{}).
			RequireFrom(store.BackendState{}),
		"BackendConfig": oa.ObjectSchema(
			"Backend Config",
			bbb.Backend{}).
			RequireFrom(bbb.Backend{}),
		"BackendSettings": oa.ObjectSchema(
			"Backend Settings ",
			store.BackendSettings{}).
			RequireFrom(store.BackendSettings{}),

		"Error":           NewErrorSchema(),
		"NotFoundError":   NewNotFoundErrorSchema(),
		"ValidationError": NewValidationErrorSchema(),
	}
}

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
		"required": []string{
			"error",
			"message",
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
			oa.SchemaRef("Error"),
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
			oa.SchemaRef("Error"),
		},
	}
}

// NewAPISecuritySchemes creates the default security schemes
func NewAPISecuritySchemes() map[string]oa.SecurityScheme {
	return map[string]oa.SecurityScheme{
		"jwt": oa.SecurityScheme{
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
			Description: "This document describes the specifications for the b3scale API v1",
			Version:     "1.0.0",
			License: oa.License{
				Name: "Apache 2.0",
				URL:  "https://www.apache.org/licenses/LICENSE-2.0.html",
			},
		},
		Servers: []oa.Server{
			{
				Description: "default server location",
				URL:         "/api/v1",
			},
		},
		Paths: NewAPIEndpointsSchema(),
		Components: oa.Components{
			Responses:       NewAPIResponses(),
			Schemas:         NewAPISchemas(),
			SecuritySchemes: NewAPISecuritySchemes(),
		},
		Tags: []oa.Tag{
			{
				Name:        "Frontends",
				Description: "B3scale tenants are called frontends. In the following section you can find all frontend related operations.",
			},
		},
		Security: []oa.SecuritySpec{
			{
				"jwt": []interface{}{},
			},
		},
	}
}
