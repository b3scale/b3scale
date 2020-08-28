package bbb

/*
 Big Blue Button Client
*/

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"time"

	"gitlab.com/infra.run/public/b3scale/pkg/config"
)

// A Client for communicating with a big blue button
// instance. Requests are signed and encoded.
// Responses are decoded.
type Client struct {
	cfg  *config.Backend
	conn *http.Client
}

// NewClient creates and configures a new http client
// and creates the big blue client object.
func NewClient(cfg *config.Backend) *Client {
	conn := &http.Client{
		Transport: &http.Transport{
			MaxIdleConnsPerHost:   10,
			IdleConnTimeout:       300 * time.Second,
			ResponseHeaderTimeout: 60 * time.Second,
		},
	}

	client := &Client{
		cfg:  cfg,
		conn: conn,
	}

	return client
}

// Internal http GET, makes the request to an URL and
// reads the entire response.
func (client *Client) get(url string) ([]byte, error) {
	// Make HTTP request
	res, err := client.conn.Get(url)
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

// Internal http POST. Reads the entire response and returns
// the data.
func (client *Client) post(
	url string,
	contentType string,
	data []byte,
) ([]byte, error) {
	// Make request to the server
	reader := bytes.NewReader(data)
	res, err := client.conn.Post(url, contentType, reader)
	if err != nil {
		return nil, err
	}

	// Read body
	defer res.Body.Close()
	rData, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return rData, nil
}
