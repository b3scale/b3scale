package callbacks

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

// Configuration
const (
	RetryCount   = 5
	RetryWaitMin = 1 * time.Second
	RetryWaitMax = 30 * time.Second

	RequestTimeout = 10 * time.Second
)

// Request describes a callback invocation.
type Request struct {
	// The callback URL
	URL string

	// Callback is any type of payload
	Callback Callback
}

// Dispatch creates a new http client instance and makes a
// request to the callback URL.
//
// For now I assume that the method is always POST. In case
// this assumption does not hold, we need to move the actual
// invocation to the callback object.
//
// Each dispatch will spawn a new goroutine. This is a naive
// implementation and might need to be moved to a worker pool.
func Dispatch(req *Request) {
	ctx := context.Background()
	go func() {
		if err := runCallback(ctx, req); err != nil {
			log.Error().
				Err(err).
				Str("url", req.URL).
				Msg("callback invocation failed")
		}
	}()
}

// Make the request to the callback URL.
// Retry on failure with backoff.
func runCallback(ctx context.Context, req *Request) error {
	wait := RetryWaitMin
	tTotal := time.Now()

	// Encode request body.
	body := req.Callback.Encode()

	for i := 1; i <= RetryCount; i++ {
		// Make request
		tReq := time.Now()
		log.Info().
			Int("try", i).
			Str("url", req.URL).
			Msg("dispatching callback")

		err := doCallbackRequest(ctx, req.URL, body)
		dReq := time.Now().Sub(tReq)
		dTotal := time.Now().Sub(tTotal)

		if err != nil {
			log.Error().
				Int("try", i).
				Dur("duration_request", dReq).
				Dur("duration_total", dTotal).
				Str("url", req.URL).
				Err(err).Msg("callback request failed")
		} else {
			log.Info().
				Int("try", i).
				Dur("duration_request", dReq).
				Dur("duration_total", dTotal).
				Str("url", req.URL).
				Msg("callback request successful")
			return nil
		}

		// Calculate backoff: Always double the wait time,
		// until we reach the maximum allowed.
		wait = wait * time.Duration(2.0)
		if wait > RetryWaitMax {
			wait = RetryWaitMax
		}
		time.Sleep(wait)
	}

	// Time to give up
	return fmt.Errorf("max retries exceeded")
}

// Do the actual request to the callback URL.
func doCallbackRequest(ctx context.Context, url, body string) error {
	client := http.DefaultClient
	ctx, cancel := context.WithTimeout(ctx, RequestTimeout)
	defer cancel()

	// Encode body: The BBB API expects 'multipart/form-data'.
	cbReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		url,
		strings.NewReader(body))
	if err != nil {
		return err
	}
	// Set content type
	cbReq.Header.Set("Content-Type", "multipart/form-data")

	res, err := client.Do(cbReq)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	return nil
}
