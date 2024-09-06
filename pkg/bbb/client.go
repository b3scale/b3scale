package bbb

/*
 Big Blue Button Client
*/

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

// A Client for communicating with a big blue button
// instance. Requests are signed and encoded.
// Responses are decoded.
type Client struct {
	conn *http.Client
}

// NewClient creates and configures a new http client
// and creates the big blue client object.
func NewClient() *Client {
	conn := &http.Client{
		Transport: &http.Transport{
			ForceAttemptHTTP2:     true,
			MaxIdleConnsPerHost:   20,
			IdleConnTimeout:       300 * time.Second,
			ResponseHeaderTimeout: 60 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // Thou shalt not follow redirects
		},
	}

	c := &Client{
		conn: conn,
	}

	return c
}

// Internal response decoding
func unmarshalRequestResponse(req *Request, data []byte) (Response, error) {
	switch req.Resource {
	case ResourceJoin:
		return UnmarshalJoinResponse(data)
	case ResourceCreate:
		return UnmarshalCreateResponse(data)
	case ResourceIsMeetingRunning:
		return UnmarshalIsMeetingRunningResponse(data)
	case ResourceEnd:
		return UnmarshalEndResponse(data)
	case ResourceGetMeetingInfo:
		return UnmarshalGetMeetingInfoResponse(data)
	case ResourceGetMeetings:
		return UnmarshalGetMeetingsResponse(data)
	case ResourceGetRecordings:
		return UnmarshalGetRecordingsResponse(data)
	case ResourcePublishRecordings:
		return UnmarshalPublishRecordingsResponse(data)
	case ResourceDeleteRecordings:
		return UnmarshalDeleteRecordingsResponse(data)
	case ResourceUpdateRecordings:
		return UnmarshalUpdateRecordingsResponse(data)
	case ResourceGetDefaultConfigXML:
		return UnmarshalGetDefaultConfigXMLResponse(data)
	case ResourceSetConfigXML:
		return UnmarshalSetConfigXMLResponse(data)
	case ResourceGetRecordingTextTracks:
		return UnmarshalGetRecordingTextTracksResponse(data)
	case ResourcePutRecordingTextTrack:
		return UnmarshalPutRecordingTextTrackResponse(data)
	}

	return nil, fmt.Errorf(
		"no response decoder for resource: %s", req.Resource)
}

// Do sends the request to the backend.
// The request is signed.
// The response is decoded into a BBB response.
func (c *Client) Do(ctx context.Context, req *Request) (Response, error) {
	log.Debug().
		Str("method", req.Request.Method).
		Str("url", req.URL()).
		Msg("client request")

	httpReqHeader := req.Request.Header.Clone()
	httpReqHeader.Del("content-length")

	var bodyReader io.Reader
	if req.Body != nil {
		bodyReader = bytes.NewReader(req.Body)
	}
	httpReq, err := http.NewRequestWithContext(
		ctx,
		req.Request.Method,
		req.URL(),
		bodyReader)
	if err != nil {
		return nil, err
	}

	// Set content type and other request headers
	httpReq.Header = httpReqHeader

	// Perform request
	httpRes, err := c.conn.Do(httpReq)
	if err != nil {
		return nil, err
	}

	// Read body
	defer httpRes.Body.Close()
	data, err := io.ReadAll(httpRes.Body)
	if err != nil {
		return nil, err
	}

	res, err := unmarshalRequestResponse(req, data)
	if err != nil {
		return nil, err
	}

	// Set response header and status
	res.SetHeader(httpRes.Header)
	res.SetStatus(httpRes.StatusCode)

	return res, nil
}
