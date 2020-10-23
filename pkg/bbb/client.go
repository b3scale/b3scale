package bbb

/*
 Big Blue Button Client
*/

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"
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
			MaxIdleConnsPerHost:   10,
			IdleConnTimeout:       300 * time.Second,
			ResponseHeaderTimeout: 60 * time.Second,
		},
	}

	c := &Client{
		conn: conn,
	}

	return c
}

// Internal http request processing: Make net/http request
// from bbb request. Read and return body.
func (c *Client) httpDo(req *Request) ([]byte, error) {
	var bodyReader io.Reader
	if req.Body != nil {
		bodyReader = bytes.NewReader(req.Body)
	}
	httpReq, err := http.NewRequest(
		req.Method,
		req.URL(c.Host, c.Secret),
		bodyReader)
	if err != nil {
		return nil, err
	}

	// Set content type and other request headers
	if req.ContentType != "" {
		httpReq.Header.Set("Content-Type", req.ContentType)
	}

	// Perform request and read response body
	res, err := c.conn.Do(httpReq)
	if err != nil {
		return nil, err
	}

	// Read body
	defer res.Body.Close()
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return data, nil
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
func (c *Client) Do(req *Request) (Response, error) {
	// DEBUG LOG:
	// TODO: Use some advanced logger or remove in production
	// because this might expose sensitive data in logfiles
	log.Println("[HTTP]", req.Method, req.URL(c.Host, c.Secret))
	data, err := c.httpDo(req)
	if err != nil {
		return nil, err
	}
	return unmarshalRequestResponse(req, data)
}
