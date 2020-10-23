package bbb

// The Backend is a bbb backend a request
// can be directed to like Client.Do(Backend, Req).
type Backend struct {
	Host   string
	Secret string
}
