package bbb

// The Frontend is a source for requests
type Frontend struct {
	Key    string `json:"key" doc:"The tenant is identified by the key, which is part of frontend specific API url." example:"greenlight01"`
	Secret string `json:"secret" doc:"The individual BBB API secrect for this frontend. API requests coming from this frontend, must be signed with this secret."`
}
