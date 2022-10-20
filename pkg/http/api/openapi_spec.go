package api

import (
	"github.com/b3scale/b3scale/pkg/bbb"
	oa "github.com/b3scale/b3scale/pkg/openapi"
	"github.com/b3scale/b3scale/pkg/store"
	"github.com/b3scale/b3scale/pkg/store/schema"
)

// NewFrontendsAPISchema generates the endpoints for the frontend
func NewFrontendsAPISchema() map[string]oa.Path {
	return map[string]oa.Path{
		"/v1/frontends": oa.Path{
			"get": oa.Operation{
				Description: "Fetch all frontends",
				OperationID: "frontendsList",
				Summary:     "List",
				Tags:        []string{"Frontends"},
				Responses: oa.ResponseRefs{
					"200": oa.ResponseRef("Frontends"),
					"400": oa.ResponseRef("BadRequest"),
					"401": oa.ResponseRef("InvalidJWTError"),
				},
				Parameters: []oa.Schema{
					oa.ParamQuery(
						"account_ref",
						"Filter by account_ref"),
					oa.ParamQuery(
						"key",
						"Show only frontends matching the exact key"),
					oa.ParamQuery(
						"key__like",
						"Show only frontends matching parts of a key"),
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
					"400": oa.ResponseRef("BadRequest"),
					"401": oa.ResponseRef("InvalidJWTError"),
				},
			},
		},
		"/v1/frontends/{id}": oa.Path{
			"parameters": []oa.Schema{
				oa.ParamID(),
			},
			"get": oa.Operation{
				Description: "Fetch a single frontend identified by ID.",
				OperationID: "frontendsRead",
				Summary:     "Read",
				Tags:        []string{"Frontends"},
				Responses: oa.ResponseRefs{
					"200": oa.ResponseRef("Frontend"),
					"400": oa.ResponseRef("BadRequest"),
					"401": oa.ResponseRef("InvalidJWTError"),
					"404": oa.ResponseRef("NotFoundError"),
				},
			},
			"patch": oa.Operation{
				Description: "Update parts of a frontend.",
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
					"400": oa.ResponseRef("BadRequest"),
					"401": oa.ResponseRef("InvalidJWTError"),
					"404": oa.ResponseRef("NotFoundError"),
				},
			},
			"delete": oa.Operation{
				Description: "Remove a frontend.\n\nAll stored recordings will also be removed.",
				OperationID: "frontendsDestroy",
				Summary:     "Delete",
				Tags:        []string{"Frontends"},
				Responses: oa.ResponseRefs{
					"200": oa.ResponseRef("Frontend"),
					"400": oa.ResponseRef("BadRequest"),
					"401": oa.ResponseRef("InvalidJWTError"),
					"404": oa.ResponseRef("NotFoundError"),
				},
			},
		},
	}
}

// NewBackendsAPISchema generates the schema for backend
// related endpoint.
func NewBackendsAPISchema() map[string]oa.Path {
	return map[string]oa.Path{
		"/v1/backends": oa.Path{
			"get": oa.Operation{
				Description: "Fetch all backends",
				OperationID: "backendsList",
				Summary:     "List",
				Tags:        []string{"Backends"},
				Responses: oa.ResponseRefs{
					"200": oa.ResponseRef("Backends"),
					"400": oa.ResponseRef("BadRequest"),
					"401": oa.ResponseRef("InvalidJWTError"),
				},
				Parameters: []oa.Schema{
					oa.ParamQuery(
						"host",
						"List backends matching this exact host."),
					oa.ParamQuery(
						"host__like",
						"List backend partially matching the host."),
					oa.ParamQuery(
						"host__ilike",
						"List backends partially matching the host, case insensitive."),
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
					"400": oa.ResponseRef("BadRequest"),
					"401": oa.ResponseRef("InvalidJWTError"),
				},
			},
		},
		"/v1/backends/{id}": oa.Path{
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
					"400": oa.ResponseRef("BadRequest"),
					"401": oa.ResponseRef("InvalidJWTError"),
					"404": oa.ResponseRef("NotFoundError"),
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
					"400": oa.ResponseRef("BadRequest"),
					"401": oa.ResponseRef("InvalidJWTError"),
					"404": oa.ResponseRef("NotFoundError"),
				},
			},
			"delete": oa.Operation{
				Description: "Remove a backend.\n\nWhen not forced, decommissioning the backend will be requested.",
				OperationID: "backendsDestroy",
				Summary:     "Delete",
				Tags:        []string{"Backends"},
				Responses: oa.ResponseRefs{
					"200": oa.ResponseRef("Backend"),
					"400": oa.ResponseRef("BadRequest"),
					"401": oa.ResponseRef("InvalidJWTError"),
					"404": oa.ResponseRef("NotFoundError"),
				},
				Parameters: []oa.Schema{
					oa.ParamQuery(
						"force",
						"When `true`, the backend will be forcefully removed from the cluster."),
				},
			},
		},
	}
}

// NewMeetingsAPISchema create the endpoint schema for meetings
func NewMeetingsAPISchema() map[string]oa.Path {
	backendIDParam := oa.ParamQuery(
		"backend_id",
		"The ID of the backend where the meetings are located.\n\n*Either this or `backend_host` is required.*")
	backendHostParam := oa.ParamQuery(
		"backend_host",
		"The full host of the backend where the meetings are located. *Either this or `backend_id` is required.*")
	return map[string]oa.Path{
		"/v1/meetings": oa.Path{
			"get": oa.Operation{
				Description: "Fetch all meetings",
				OperationID: "meetingsList",
				Summary:     "List",
				Tags:        []string{"Meetings"},
				Responses: oa.ResponseRefs{
					"200": oa.ResponseRef("Meetings"),
					"400": oa.ResponseRef("BadRequest"),
					"401": oa.ResponseRef("InvalidJWTError"),
				},
				Parameters: []oa.Schema{
					backendIDParam, backendHostParam,
				},
			},
		},
		"/v1/meetings/{id}": oa.Path{
			"parameters": []oa.Schema{
				oa.ParamID(),
			},
			"get": oa.Operation{
				Description: "Fetch a single meeting identified by ID.",
				OperationID: "meetingsRead",
				Summary:     "Read",
				Tags:        []string{"Meetings"},
				Responses: oa.ResponseRefs{
					"200": oa.ResponseRef("Meeting"),
					"400": oa.ResponseRef("BadRequest"),
					"401": oa.ResponseRef("InvalidJWTError"),
					"404": oa.ResponseRef("NotFoundError"),
				},
			},
			"patch": oa.Operation{
				Description: "Update parts of a meeting.",
				OperationID: "meetingsPatch",
				Summary:     "Update",
				Tags:        []string{"Meetings"},
				RequestBody: &oa.Request{
					Content: map[string]oa.MediaType{
						oa.ApplicationJSON: oa.MediaType{
							Schema: oa.SchemaRef("MeetingPatch"),
						},
					},
				},
				Responses: oa.ResponseRefs{
					"200": oa.ResponseRef("Meeting"),
					"400": oa.ResponseRef("BadRequest"),
					"401": oa.ResponseRef("InvalidJWTError"),
					"404": oa.ResponseRef("NotFoundError"),
				},
			},
			"delete": oa.Operation{
				Description: "Remove a meeting.",
				OperationID: "meetingsDestroy",
				Summary:     "Delete",
				Tags:        []string{"Meetings"},
				Responses: oa.ResponseRefs{
					"200": oa.ResponseRef("Meeting"),
					"400": oa.ResponseRef("BadRequest"),
					"401": oa.ResponseRef("InvalidJWTError"),
					"404": oa.ResponseRef("NotFoundError"),
				},
			},
		},
	}
}

// NewCommandsAPISchema create the endpoint schema for commands
func NewCommandsAPISchema() map[string]oa.Path {
	return map[string]oa.Path{
		"/v1/commands": oa.Path{
			"get": oa.Operation{
				Description: "Fetch current command queue.",
				OperationID: "commandsList",
				Summary:     "List",
				Tags:        []string{"Commands"},
				Responses: oa.ResponseRefs{
					"200": oa.ResponseRef("Commands"),
					"400": oa.ResponseRef("BadRequest"),
					"401": oa.ResponseRef("InvalidJWTError"),
				},
			},
			"post": oa.Operation{
				Description: "Insert a new command into the queue.\n\nCurrently only `end_all_meetings` for a given backend is supported.\n\nExample: `{\"action\": \"end_all_meetings\", \"params\": {\"BackendID\": \"b056bc5e-372e-4562-b23a-bd6a92634e7b\"}}`",
				OperationID: "commandsCreate",
				Summary:     "Create",
				Tags:        []string{"Commands"},
				RequestBody: &oa.Request{
					Content: map[string]oa.MediaType{
						oa.ApplicationJSON: oa.MediaType{
							Schema: oa.SchemaRef("CommandRequest"),
						},
					},
				},
				Responses: oa.ResponseRefs{
					"202": oa.ResponseRef("Command"),
					"400": oa.ResponseRef("BadRequest"),
					"401": oa.ResponseRef("InvalidJWTError"),
				},
			},
		},
		"/v1/commands/{id}": oa.Path{
			"parameters": []oa.Schema{
				oa.ParamID(),
			},
			"get": oa.Operation{
				Description: "Fetch a single command identified by ID.",
				OperationID: "commandsRead",
				Summary:     "Read",
				Tags:        []string{"Commands"},
				Responses: oa.ResponseRefs{
					"200": oa.ResponseRef("Command"),
					"400": oa.ResponseRef("BadRequest"),
					"401": oa.ResponseRef("InvalidJWTError"),
					"404": oa.ResponseRef("NotFoundError"),
				},
			},
		},
	}
}

// NewRecordingsImportAPISchema creates the api schema for
// accepting a BBB recodrings metadata document
func NewRecordingsImportAPISchema() map[string]oa.Path {
	return map[string]oa.Path{
		"/v1/recordings-import": oa.Path{
			"post": oa.Operation{
				Summary:     "Import Recording Meta",
				Description: "Upload an recordings metadata XML document. The recording will be imported.\n\nThese are typically read from `/var/bigbluebutton/published/presentation/...meetingID.../metadata.xml`, see `post_publish_b3scale_import.rb` script.",
				OperationID: "recordingsImport",
				RequestBody: &oa.Request{
					Content: map[string]oa.MediaType{
						"application/xml": oa.MediaType{
							Schema: oa.Schema{
								"type": "object",
							},
						},
					},
				},
				Tags: []string{"Recordings"},
				Responses: oa.ResponseRefs{
					"200": oa.ResponseRef("Recording"),
					"400": oa.ResponseRef("BadRequest"),
					"401": oa.ResponseRef("InvalidJWTError"),
				},
			},
		},
	}
}

// NewAgentAPISchema creates the API schema for the node agent
func NewAgentAPISchema() map[string]oa.Path {
	return map[string]oa.Path{
		"/v1/agent/backend": oa.Path{
			"get": oa.Operation{
				OperationID: "agentBackendRead",
				Summary:     "Read Backend",
				Description: "Get the backend associated with the agent.",
				Tags:        []string{"Agent"},
				Responses: oa.ResponseRefs{
					"200": oa.ResponseRef("Backend"),
					"400": oa.ResponseRef("BadRequest"),
					"401": oa.ResponseRef("InvalidJWTError"),
					"404": oa.ResponseRef("NotFoundError"),
				},
			},
		},
		"/v1/agent/heartbeat": oa.Path{
			"post": oa.Operation{
				OperationID: "agentHeartbeatCreate",
				Summary:     "Create Heartbeat",
				Description: "Notify b3scale, that the agent is still alive.",
				Tags:        []string{"Agent"},
				Responses: oa.ResponseRefs{
					"200": oa.ResponseRef("Heartbeat"),
					"400": oa.ResponseRef("BadRequest"),
					"401": oa.ResponseRef("InvalidJWTError"),
					"404": oa.ResponseRef("NotFoundError"),
				},
			},
		},
		"/v1/agent/rpc": oa.Path{
			"post": oa.Operation{
				OperationID: "agentRpc",
				Summary:     "RPC",
				Description: "Perform a remote procedure call.\n\nThis API allows for remote procedures that involve complex that can not be expressed sufficiently through resource manipulation.\n\n**Warning:** this API is only meant to be used by the agent.\n\nOnly the envelope format is described here. For details, please check the source in `http/api/rpc.go`.",
				Tags:        []string{"Agent"},
				RequestBody: &oa.Request{
					Content: map[string]oa.MediaType{
						oa.ApplicationJSON: oa.MediaType{
							Schema: oa.SchemaRef("RPCRequest"),
						},
					},
				},
				Responses: oa.ResponseRefs{
					"200": oa.ResponseRef("RPCResponse"),
					"400": oa.ResponseRef("BadRequest"),
					"401": oa.ResponseRef("InvalidJWTError"),
					"404": oa.ResponseRef("NotFoundError"),
				},
			},
		},
	}
}

// NewMetaEndpointsSchema creates the api meta endpoints
func NewMetaEndpointsSchema() map[string]oa.Path {
	return map[string]oa.Path{
		"/v1": oa.Path{
			"get": oa.Operation{
				Description: "Retrieve an API status",
				OperationID: "statusRead",
				Summary:     "Read Status",
				Tags:        []string{"API"},
				Responses: oa.ResponseRefs{
					"200": oa.ResponseRef("Status"),
					"400": oa.ResponseRef("BadRequest"),
					"401": oa.ResponseRef("InvalidJWTError"),
				},
			},
		},
	}
}

// NewCtrlEndpointsSchema creates the api ctrl endpoints
func NewCtrlEndpointsSchema() map[string]oa.Path {
	return map[string]oa.Path{
		"/v1/ctrl/migrate": oa.Path{
			"post": oa.Operation{
				OperationID: "ctrlMigrate",
				Summary:     "Migrate Database",
				Tags:        []string{"CTRL"},
				Responses: oa.ResponseRefs{
					"200": oa.ResponseRef("MigrateStatus"),
					"400": oa.ResponseRef("BadRequest"),
					"401": oa.ResponseRef("InvalidJWTError"),
				},
			},
		},
	}

}

// NewAPIEndpointsSchema combines all the endpoints schemas
func NewAPIEndpointsSchema() map[string]oa.Path {
	return oa.Endpoints(
		NewMetaEndpointsSchema(),
		NewFrontendsAPISchema(),
		NewBackendsAPISchema(),
		NewMeetingsAPISchema(),
		NewCommandsAPISchema(),
		NewRecordingsImportAPISchema(),
		NewAgentAPISchema(),
		NewCtrlEndpointsSchema(),
	)
}

// NewAPIResponses creates all default (error) responses
func NewAPIResponses() map[string]oa.Response {
	return map[string]oa.Response{
		"Status": oa.Response{
			Description: "API and Server Status",
			Content: map[string]oa.MediaType{
				oa.ApplicationJSON: oa.MediaType{
					Schema: oa.SchemaRef("Status"),
				},
			},
		},

		"Frontends": oa.Response{
			Description: "List of Frontends",
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
			Description: "List of Backends",
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

		"Meetings": oa.Response{
			Description: "List of Meetings",
			Content: map[string]oa.MediaType{
				oa.ApplicationJSON: oa.MediaType{
					Schema: oa.SchemaRef("Meetings"),
				},
			},
		},
		"Meeting": oa.Response{
			Description: "Meeting",
			Content: map[string]oa.MediaType{
				oa.ApplicationJSON: oa.MediaType{
					Schema: oa.SchemaRef("Meeting"),
				},
			},
		},

		"Commands": oa.Response{
			Description: "List of Commands",
			Content: map[string]oa.MediaType{
				oa.ApplicationJSON: oa.MediaType{
					Schema: oa.SchemaRef("Commands"),
				},
			},
		},
		"Command": oa.Response{
			Description: "Command",
			Content: map[string]oa.MediaType{
				oa.ApplicationJSON: oa.MediaType{
					Schema: oa.SchemaRef("Command"),
				},
			},
		},

		"Recording": oa.Response{
			Description: "Recording",
			Content: map[string]oa.MediaType{
				oa.ApplicationJSON: oa.MediaType{
					Schema: oa.SchemaRef("Recording"),
				},
			},
		},

		"Heartbeat": oa.Response{
			Description: "Heartbeat",
			Content: map[string]oa.MediaType{
				oa.ApplicationJSON: oa.MediaType{
					Schema: oa.SchemaRef("Heartbeat"),
				},
			},
		},

		"RPCResponse": oa.Response{
			Description: "RPCResponse",
			Content: map[string]oa.MediaType{
				oa.ApplicationJSON: oa.MediaType{
					Schema: oa.SchemaRef("RPCResponse"),
				},
			},
		},

		"MigrateStatus": oa.Response{
			Description: "MigrateStatus",
			Content: map[string]oa.MediaType{
				oa.ApplicationJSON: oa.MediaType{
					Schema: oa.SchemaRef("SchemaStatus"),
				},
			},
		},

		"NotFoundError": oa.Response{
			Description: "The requested resource could not be found.",
			Content: map[string]oa.MediaType{
				oa.ApplicationJSON: oa.MediaType{
					Schema: oa.SchemaRef("NotFoundError"),
				},
			},
		},
		"InvalidJWTError": oa.Response{
			Description: "The JWT authorization token was invalid or expired.",
			Content: map[string]oa.MediaType{
				oa.ApplicationJSON: oa.MediaType{
					Schema: oa.SchemaRef("ServerError"),
				},
			},
		},
		"BadRequest": oa.Response{
			Description: "The request was invalid. Maybe a JWT authorization header was not provided or a validation failed.",
			Content: map[string]oa.MediaType{
				oa.ApplicationJSON: oa.MediaType{
					Schema: oa.SchemaRef("ServerError"),
				},
			},
		},
	}
}

// NewAPISchemas creates the schemas we use in the API
func NewAPISchemas() map[string]oa.Schema {
	return map[string]oa.Schema{
		"Status": oa.ObjectSchema(
			"Server Status",
			StatusResponse{}).
			RequireFrom(StatusResponse{}),

		"Frontends": oa.ArraySchema(
			"A list of frontends",
			oa.SchemaRef("Frontend")),
		"FrontendRequest": oa.ObjectSchema(
			"A Frontend Request",
			store.FrontendState{}).
			Only("active", "bbb", "settings", "account_ref").
			Require("bbb"),
		"FrontendPatch": oa.ObjectSchema(
			"A Frontend Update",
			store.FrontendState{}).
			Only("active", "bbb", "settings", "account_ref").
			Patch("bbb"),
		"Frontend": oa.ObjectSchema(
			"Frontend",
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
			"Frontend Settings",
			store.FrontendSettings{}).
			RequireFrom(store.FrontendSettings{}),
		"DefaultPresentationSettings": oa.ObjectSchema(
			"Default Presentation",
			store.DefaultPresentationSettings{}).
			RequireFrom(store.DefaultPresentationSettings{}),

		"Backends": oa.ArraySchema(
			"List of Backends",
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

		"Meetings": oa.ArraySchema(
			"List of Meetings",
			oa.SchemaRef("Meeting")),
		"Meeting": oa.ObjectSchema(
			"Meeting",
			store.MeetingState{}).
			RequireFrom(store.MeetingState{}),
		"MeetingPatch": oa.ObjectSchema(
			"Meeting Update",
			store.MeetingState{}).
			Only("meeting").
			Patch("meeting").
			Require("meeting"),
		"MeetingInfo": oa.ObjectSchema(
			"Meeting Info",
			bbb.Meeting{}).
			RequireFrom(bbb.Meeting{}),
		"MeetingInfoPatch": oa.ObjectSchema(
			"Meeting Info",
			bbb.Meeting{}),
		"Attendee": oa.ObjectSchema(
			"Meeting Attendee",
			bbb.Attendee{}).
			RequireFrom(bbb.Attendee{}),
		"Breakout": oa.ObjectSchema(
			"Meeting Breakout Room",
			bbb.Breakout{}).
			RequireFrom(bbb.Breakout{}),

		"Commands": oa.ArraySchema(
			"List of Commands",
			oa.SchemaRef("Command")),
		"Command": oa.ObjectSchema(
			"Command",
			store.Command{}).
			RequireFrom(store.Command{}).
			Nullable("result", "started_at", "stopped_at"),
		"CommandRequest": oa.ObjectSchema(
			"Command Request",
			store.Command{}).
			Only("action", "params").
			Require("action", "params"),

		"Recording": oa.ObjectSchema(
			"Recording",
			bbb.Recording{}).
			RequireFrom(bbb.Recording{}),
		"Format": oa.ObjectSchema(
			"Format",
			bbb.Format{}).
			RequireFrom(bbb.Format{}),
		"Preview": oa.ObjectSchema(
			"Preview",
			bbb.Preview{}).RequireFrom(bbb.Preview{}),
		"Images": oa.ObjectSchema(
			"Images",
			bbb.Images{}).RequireFrom(bbb.Images{}),
		"Image": oa.ObjectSchema(
			"Image",
			bbb.Images{}).RequireFrom(bbb.Image{}),

		"Heartbeat": oa.ObjectSchema(
			"Hearbeat", store.AgentHeartbeat{}).
			RequireFrom(store.AgentHeartbeat{}),

		"RPCRequest":  NewRPCRequestSchema(),
		"RPCResponse": NewRPCResponseSchema(),

		"SchemaStatus": oa.ObjectSchema(
			"SchemaStatus", schema.Status{}).
			RequireFrom(schema.Status{}),

		"MigrationState": oa.ObjectSchema(
			"MigrationState", schema.MigrationState{}).
			RequireFrom(schema.MigrationState{}),

		"Error":           NewErrorSchema(),
		"NotFoundError":   NewNotFoundErrorSchema(),
		"ValidationError": NewValidationErrorSchema(),
		"ServerError":     NewServerErrorSchema(),
	}
}

// RPC

// NewRPCRequestSchema creates the RPCRequest schema
func NewRPCRequestSchema() oa.Schema {
	return oa.Schema{
		"description": "RPCRequest Envelope",
		"type":        "object",
		"properties": oa.Properties{
			"action": oa.FieldProperty{
				"type":        "string",
				"description": "The name of the procedure to invoke.",
			},
			"payload": oa.FieldProperty{
				"type":                 "object",
				"additionalProperties": true,
				"description":          "An object with the RPC request parameters.",
			},
		},
		"required": []string{"action", "payload"},
	}
}

// NewRPCResponseSchema creates the RPC response schema
func NewRPCResponseSchema() oa.Schema {
	return oa.Schema{
		"description": "RPCResponse Envelope",
		"type":        "object",
		"properties": oa.Properties{
			"status": oa.FieldProperty{
				"type":        "string",
				"description": "The name of the procedure to invoke.",
				"enum":        []string{"ok", "error"},
			},
			"result": oa.FieldProperty{
				"type":                 "object",
				"description":          "An object with the encoded result of the call.",
				"additionalProperties": true,
			},
		},
		"required": []string{"status", "result"},
	}
}

// NewServerErrorSchema creates a fallback error schema with
// only contains the error message.
func NewServerErrorSchema() oa.Schema {
	return oa.Schema{
		"description": "A Server Error",
		"type":        "object",
		"properties": oa.Properties{
			"message": oa.Property{
				Type: "string",
				Description: "A human readable message with details " +
					"about the error.",
			},
		},
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
			Type:         "http",
			Scheme:       "bearer",
			Description:  "Authorization using a JWT bearer token",
			BearerFormat: "JWT",
			// In:           "header",
			// Name:         "JWT bearer token",
		},
	}
}

// NewAPISpec makes the api specs for b3scale
func NewAPISpec() *oa.Spec {
	return &oa.Spec{
		OpenAPI: "3.0.1",
		Info: oa.Info{
			Title:       "b3scale api v1",
			Description: "This document describes the specifications for the b3scale API v1.",
			Version:     "1.1.0",
			License: oa.License{
				Name: "Apache 2.0",
				URL:  "https://www.apache.org/licenses/LICENSE-2.0.html",
			},
			Contact: oa.Contact{
				Name:  "The B3Scale Developers",
				URL:   "https://b3scale.io",
				Email: "mail@infra.run",
			},
		},
		Servers: []oa.Server{
			{
				Description: "default server location",
				URL:         "/api",
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
				Name:        "API",
				Description: "Get meta information about the API.",
			},
			{
				Name:        "Frontends",
				Description: "B3scale tenants are called frontends. In the following section you can find all frontend related operations.",
			},
			{
				Name:        "Backends",
				Description: "Big Blue Button (BBB) servers are called backends. Each is a node in the cluster, having an agent running.\n\nThe following endpoints are for managing backends.",
			},
			{
				Name:        "Meetings",
				Description: "The meetings API can be used to update and query meetings. Creating new meetings is not supported at the time.",
			},
			{
				Name:        "Recordings",
				Description: "Currently only importing recording by uploading a `metadata.xml` is supported.",
			},
			{
				Name:        "Commands",
				Description: "The commands API is used queue asynchronous commands. Currently only `end_all_meetings` for a given backend is supported.",
			},
			{
				Name:        "Agent",
				Description: "This API is used by the agent, running on each node.",
			},
			{
				Name:        "CTRL",
				Description: "This api endpoint is for sending control commands to the server.",
			},
		},
		Security: []oa.SecuritySpec{
			{
				"jwt": []interface{}{},
			},
		},
	}
}
