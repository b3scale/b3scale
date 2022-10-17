package bbb

// The Backend is a bbb backend a request
// can be directed to like Client.Do(Backend, Req).
type Backend struct {
	Host   string `json:"host" doc:"The full qualified address of the host, including the API endpoint." example:"https://backendnode01/bigbluebutton/api/"`
	Secret string `json:"secret" doc:"The API secret for the BBB host."`
}
