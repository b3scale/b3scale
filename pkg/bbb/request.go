package bbb

import (
	"crypto/sha1"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
)

// Well known params
const (
	ParamMeetingID = "meetingID"
	ParamChecksum  = "checksum"
)

// Params for the BBB API (we opt for stringly typed.)
type Params map[string]string

// String of the query parameters.
// The order of the parameters is deterministic.
func (p Params) String() string {
	keys := make([]string, 0, len(p))
	for key := range p {
		// We omit the checksum.
		if key == "checksum" {
			continue
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)

	// Encode query string
	var q []string
	for _, k := range keys {
		v := p[k]
		vStr := url.QueryEscape(v)
		q = append(q, fmt.Sprintf("%s=%s", k, vStr))
	}
	return strings.Join(q, "&")
}

// MeetingID retrievs the well known meeting id
// value from the set of params.
func (p Params) MeetingID() (string, bool) {
	id, ok := p[ParamMeetingID]
	if !ok {
		return "", false
	}
	return id, true
}

// Checksum retrievs the well known checksum param
func (p Params) Checksum() (string, bool) {
	checksum, ok := p[ParamChecksum]
	if !ok {
		return "", false
	}
	return checksum, true
}

// Request is a bbb request as decoded from the
// incoming url - but can be directly passed on to a
// BigBlueButton server.
//
// It is associated with a backend and a frontend.
type Request struct {
	Resource    string
	Method      string
	ContentType string
	Params      Params
	Body        []byte
	Checksum    string

	Backend  *Backend
	Frontend *Frontend
}

// Request Builders:

// WithBackend adds a backend to the request
func (req *Request) WithBackend(b *Backend) *Request {
	req.Backend = b
	return req
}

// WithFrontend adds a frontend to the request
func (req *Request) WithFrontend(f *Frontend) *Request {
	req.Frontend = f
	return req
}

// JoinRequest creates a new join request
func JoinRequest(params Params) *Request {
	return &Request{
		Method:   http.MethodGet,
		Resource: ResourceJoin,
		Params:   params,
	}
}

// EndRequest creates a meeting end request
func EndRequest(params Params) *Request {
	return &Request{
		Method:   http.MethodGet,
		Resource: ResourceEnd,
		Params:   params,
	}
}

// CreateRequest creates a new create request
func CreateRequest(params Params, body []byte) *Request {
	return &Request{
		Method:      http.MethodPost,
		Resource:    ResourceCreate,
		ContentType: "application/xml",
		Params:      params,
		Body:        body,
	}
}

// GetMeetingsRequest builds a new getMeetings request
func GetMeetingsRequest(params Params) *Request {
	return &Request{
		Method:   http.MethodGet,
		Resource: ResourceGetMeetings,
		Params:   params,
	}
}

// GetMeetingInfoRequest creates a new getMeetingInfo request
func GetMeetingInfoRequest(params Params) *Request {
	return &Request{
		Method:   http.MethodGet,
		Resource: ResourceGetMeetingInfo,
		Params:   params,
	}
}

// IsMeetingRunningRequest makes a new isMeetingRunning request
func IsMeetingRunningRequest(params Params) *Request {
	return &Request{
		Method:   http.MethodGet,
		Resource: ResourceIsMeetingRunning,
		Params:   params,
	}
}

// Internal calculate checksum with a given secret.
func (req *Request) calculateChecksumSHA1(secret string) []byte {
	qry := req.Params.String()
	// Calculate checksum with server secret
	// Basically sign the endpoint + params
	mac := []byte(req.Resource + qry + secret)
	shasum := sha1.New()
	shasum.Write(mac)
	return []byte(hex.EncodeToString(shasum.Sum(nil)))
}

// Internal calculate checksum with a given secret.
func (req *Request) calculateChecksumSHA256(secret string) []byte {
	qry := req.Params.String()
	// Calculate checksum with server secret
	// Basically sign the endpoint + params
	mac := []byte(req.Resource + qry + secret)
	shasum := sha256.New()
	shasum.Write(mac)
	return []byte(hex.EncodeToString(shasum.Sum(nil)))
}

// Verify request coming from a frontend.
// Compare checksum with the checksum calculated from the params
// and the frontend secret
func (req *Request) Verify() error {
	secret := req.Frontend.Secret
	var expected []byte
	if len(req.Checksum) > 40 {
		expected = req.calculateChecksumSHA256(secret)
	} else {
		expected = req.calculateChecksumSHA1(secret)
	}
	if subtle.ConstantTimeCompare(
		expected,
		[]byte(req.Checksum)) != 1 {
		return fmt.Errorf("invalid checksum")
	}
	return nil
}

// Sign a request, with the backend secret.
func (req *Request) Sign() string {
	secret := req.Backend.Secret
	return string(req.calculateChecksumSHA256(secret))
}

// URL builds the URL representation of the
// request, directed at a backend.
func (req *Request) URL() string {
	// In case the configuration does not end in a trailing slash,
	// append it when needed.
	apiBase := req.Backend.Host
	if !strings.HasSuffix(apiBase, "/") {
		apiBase += "/"
	}

	// Sign the request and encode params
	qry := req.Params.String()
	chksum := req.Sign()

	// Build request url
	reqURL := apiBase + req.Resource
	if qry == "" {
		reqURL += "?checksum=" + chksum
	} else {
		reqURL += "?" + qry + "&checksum=" + chksum
	}
	return reqURL
}
