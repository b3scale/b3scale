package bbb

import (
	"crypto/sha1"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"net/url"
	"sort"
	"strings"

	"gitlab.com/infra.run/public/b3scale/pkg/config"
)

// Params for the BBB API
type Params map[string]interface{}

// Encode query parameters.
// The order of the parameters is deterministic.
func (p Params) Encode() string {
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

// Request is a bbb request as decoded from the
// incoming url - but can be directly passed on to a
// BigBlueButton server.
type Request struct {
	Frontend *config.Frontend
	Backend  *config.Backend
	Resource string
	Params   Params
	Checksum []byte
}

// Internal calculate checksum with a given secret.
func (req *Request) calculateChecksum(secret string) []byte {
	qry := req.Params.Encode()

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
func (req *Request) Validate() error {
	secret := req.Frontend.Secret
	expected := req.calculateChecksum(secret)
	if subtle.ConstantTimeCompare(expected, req.Checksum) != 1 {
		return fmt.Errorf("invalid checksum")
	}
	return nil
}

// Sign a request, with the backend secret.
func (req *Request) Sign() string {
	secret := req.Backend.Secret
	return string(req.calculateChecksum(secret))
}

// String builds the URL representation of the
// request, directed at a backend.
func (req *Request) String() string {
	// In case the configuration does not end in a trailing slash,
	// append it when needed.
	apiBase := req.Backend.Host
	if !strings.HasSuffix(apiBase, "/") {
		apiBase += "/"
	}

	// Sign the request and encode params
	qry := req.Params.Encode()
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
