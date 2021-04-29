package bbb

import (
	"crypto/sha1"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strings"
)

// Well known params
const (
	ParamMeetingID = "meetingID"
	ParamChecksum  = "checksum"
)

var (
	// ReQueryChecksum is used for removing the checksum
	// from a querystring in the incoming HTTP request
	ReQueryChecksum = regexp.MustCompile("&checksum(=[^&]*)?|^checksum(=[^&]*)?&?")
)

// Params for the BBB API (we opt for stringly typed.)
type Params map[string]string

// String of the query parameters.
// The order of the parameters is made deterministic.
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
	q := make([]string, 0, len(keys))
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
	*http.Request

	Resource string
	Params   Params
	Body     []byte
	Checksum string

	Backend  *Backend
	Frontend *Frontend
}

// HasBody checks for the presence of a request body
func (req *Request) HasBody() bool {
	if req.Body != nil && len(req.Body) > 0 {
		return true
	}
	return false
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

// MarshalURLSafe will encode the request in an urlsafe way
// using json+base64.
func (req *Request) MarshalURLSafe() []byte {
	// We can not directly mashal the http.Request, so we create
	// a temporary map with all relevant data
	repr := map[string]interface{}{
		"mth": req.Request.Method,
		"hdr": req.Request.Header,
		"res": req.Resource,
		"prm": req.Params,
	}
	data, err := json.Marshal(repr)
	if err != nil {
		panic(err)
	}
	// Encode as urlsafe base64
	buf := make([]byte, base64.RawURLEncoding.EncodedLen(len(data)))
	base64.RawURLEncoding.Encode(buf, data)
	return buf
}

// UnmarshalURLSafeRequest will decode an encoded request. Remember not
// to trust any user data! THIS IS NOT SIGNED.
func UnmarshalURLSafeRequest(data []byte) (req *Request, err error) {
	payload := make([]byte, base64.RawURLEncoding.DecodedLen(len(data)))
	if _, err := base64.RawURLEncoding.Decode(payload, data); err != nil {
		return nil, err
	}

	var r interface{}
	if err := json.Unmarshal(payload, &r); err != nil {
		return nil, err
	}

	// Recover if the the following decoding of the representation
	// into a Request fails.
	defer func() {
		if recv := recover(); recv != nil {
			err = fmt.Errorf("decoding error: %v", recv)
		}
	}()

	req = decodeURLSafeRequest(r)
	return req, nil
}

func decodeURLSafeRequest(enc interface{}) *Request {
	repr := enc.(map[string]interface{})
	params := Params{}
	if repr["prm"] != nil {
		for k, v := range repr["prm"].(map[string]interface{}) {
			params[k] = v.(string)
		}
	}

	header := http.Header{}
	if repr["hdr"] != nil {
		for k, v := range repr["hdr"].(map[string]interface{}) {
			values := make([]string, len(v.([]interface{})))
			for j, hv := range v.([]interface{}) {
				values[j] = hv.(string)
			}
			header[k] = values
		}
	}

	return &Request{
		Resource: repr["res"].(string),
		Params:   params,
		Request: &http.Request{
			Method: repr["mth"].(string),
			Header: header,
		},
	}
}

// JoinRequest creates a new join request
func JoinRequest(params Params) *Request {
	return &Request{
		Request: &http.Request{
			Method: http.MethodGet,
		},
		Resource: ResourceJoin,
		Params:   params,
	}
}

// EndRequest creates a meeting end request
func EndRequest(params Params) *Request {
	return &Request{
		Request: &http.Request{
			Method: http.MethodGet,
		},
		Resource: ResourceEnd,
		Params:   params,
	}
}

// CreateRequest creates a new create request
func CreateRequest(params Params, body []byte) *Request {
	return &Request{
		Request: &http.Request{
			Method: http.MethodPost,
			Header: http.Header{
				"Content-Type": []string{"application/xml"},
			},
		},
		Resource: ResourceCreate,
		Params:   params,
		Body:     body,
	}
}

// GetMeetingsRequest builds a new getMeetings request
func GetMeetingsRequest(params Params) *Request {
	return &Request{
		Request: &http.Request{
			Method: http.MethodGet,
		},
		Resource: ResourceGetMeetings,
		Params:   params,
	}
}

// GetMeetingInfoRequest creates a new getMeetingInfo request
func GetMeetingInfoRequest(params Params) *Request {
	return &Request{
		Request: &http.Request{
			Method: http.MethodGet,
		},
		Resource: ResourceGetMeetingInfo,
		Params:   params,
	}
}

// IsMeetingRunningRequest makes a new isMeetingRunning request
func IsMeetingRunningRequest(params Params) *Request {
	return &Request{
		Request: &http.Request{
			Method: http.MethodGet,
		},
		Resource: ResourceIsMeetingRunning,
		Params:   params,
	}
}

// Internal calculate checksum with a given secret.
func (req *Request) calculateChecksumSHA1(query, secret string) []byte {
	// Calculate checksum with server secret
	// Basically sign the endpoint + params
	mac := []byte(req.Resource + query + secret)
	shasum := sha1.New()
	shasum.Write(mac)
	return []byte(hex.EncodeToString(shasum.Sum(nil)))
}

// Internal calculate checksum with a given secret.
func (req *Request) calculateChecksumSHA256(query, secret string) []byte {
	// Calculate checksum with server secret
	// Basically sign the endpoint + params
	mac := []byte(req.Resource + query + secret)
	shasum := sha256.New()
	shasum.Write(mac)
	return []byte(hex.EncodeToString(shasum.Sum(nil)))
}

// Verify request coming from a frontend:
// Compare checksum with the checksum calculated from the
// incoming raw query string and the frontend secret
func (req *Request) Verify() error {
	// Use request querystring and remove checksum
	query := ReQueryChecksum.ReplaceAllString(req.Request.URL.RawQuery, "")
	secret := req.Frontend.Secret

	var expected []byte
	if len(req.Checksum) > 40 {
		expected = req.calculateChecksumSHA256(query, secret)
	} else {
		expected = req.calculateChecksumSHA1(query, secret)
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
	query := req.Params.String()
	return string(req.calculateChecksumSHA256(query, secret))
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
