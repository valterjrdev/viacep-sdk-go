package viacep

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"time"

	"github.com/go-resty/resty/v2"
)

var headersDefault = map[string]string{"Content-Type": "application/json", "Accept": "application/json"}

type Http interface {
	// get sends an HTTP GET request to the specified URL and stores the response.
	//
	// Parameters:
	//   - ctx: The context to manage the request lifecycle, such as timeouts or cancellations.
	//   - url: The URL to which the GET request will be sent.
	//   - dest:  A pointer to the variable where the response data will be stored. The type of dest should be a pointer to an object that matches the expected data structure.
	//
	// Returns:
	//   - error: If an error occurs during the request, it will be returned. Otherwise, nil is returned.
	Get(ctx context.Context, url string, dest any) error
}

type HttpClient struct {
	httpClient *resty.Client
}

func NewHttpClient(maxRetry int) *HttpClient {
	restyHttpClient := resty.New()
	restyHttpClient.SetRetryCount(maxRetry).SetRetryWaitTime(500 * time.Millisecond)

	return &HttpClient{
		httpClient: restyHttpClient,
	}
}

func (r *HttpClient) Get(ctx context.Context, url string, dest any) error {
	if reflect.ValueOf(dest).Kind() != reflect.Ptr {
		return fmt.Errorf("expected a pointer for 'dest', but got %s", reflect.TypeOf(dest))
	}

	req := r.httpClient.R().SetContext(ctx)
	resp, err := req.SetHeaders(headersDefault).SetResult(dest).Get(url)
	if err != nil {
		return fmt.Errorf("failed to send GET request to %s: %w", url, err)
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("API request to %s returned status code %d; expected %d (OK)", resp.Request.URL, resp.StatusCode(), http.StatusOK)
	}

	return nil
}
