package bbb

import (
	"crypto/sha1"
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
)

// Params for the BBB API
type Params map[string]interface{}

// String of the query parameters.
// The order of the parameters is deterministic.
func (p Params) String() string {
	keys := make([]string, 0, len(p))
	for key := range p {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	// Encode query string
	var q []string
	for _, k := range keys {
		v := p[k]
		vStr := url.QueryEscape(fmt.Sprintf("%v", v))
		q = append(q, fmt.Sprintf("%s=%s", k, vStr))
	}
	return strings.Join(q, "&")
}

// GetMeetingID retrievs the well known
// parameter from the set of params.
func (p Params) GetMeetingID() (string, bool) {
	iID, ok := p[ParamMeetingID]
	if !ok {
		return "", false
	}
	id, ok := iID.(string)
	if !ok {
		return "", false
	}
	return id, true
}

// Request is a bbb request as decoded from the
// incoming url - but can be directly passed on to a
// BigBlueButton server.
type Request struct {
	Resource    string
	Method      string
	ContentType string
	Params      Params
	Body        []byte
	Checksum    []byte
}

// Request Builders:

// JoinRequest creates a new join request
func JoinRequest(params Params) *Request {
	return &Request{
		Method:   http.MethodGet,
		Resource: ResourceJoin,
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
func (req *Request) calculateChecksum(secret string) []byte {
	qry := req.Params.String()
	// Calculate checksum with server secret
	// Basically sign the endpoint + params
	mac := []byte(req.Resource + qry + secret)
	shasum := sha1.New()
	shasum.Write(mac)
	return []byte(hex.EncodeToString(shasum.Sum(nil)))
}

// Validate request coming from a frontend.
// Compare checksum with the checksum calculated from the params
// and the frontend secret
func (req *Request) Validate(secret string) error {
	expected := req.calculateChecksum(secret)
	if subtle.ConstantTimeCompare(expected, req.Checksum) != 1 {
		return fmt.Errorf("invalid checksum")
	}
	return nil
}

// Sign a request, with the backend secret.
func (req *Request) Sign(secret string) string {
	return string(req.calculateChecksum(secret))
}

// URL builds the URL representation of the
// request, directed at a backend.
func (req *Request) URL(apiBase, secret string) string {
	// In case the configuration does not end in a trailing slash,
	// append it when needed.
	if !strings.HasSuffix(apiBase, "/") {
		apiBase += "/"
	}

	// Sign the request and encode params
	qry := req.Params.String()
	chksum := req.Sign(secret)

	// Build request url
	reqURL := apiBase + req.Resource
	if qry == "" {
		reqURL += "?checksum=" + chksum
	} else {
		reqURL += "?" + qry + "&checksum=" + chksum
	}
	return reqURL
}
