package viacep

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"time"

	"github.com/go-resty/resty/v2"
)

type httpClient interface {
	// get sends an HTTP GET request to the specified URL and stores the response.
	//
	// Parameters:
	//   - ctx: The context to manage the request lifecycle, such as timeouts or cancellations.
	//   - url: The URL to which the GET request will be sent.
	//   - dest:  A pointer to the variable where the response data will be stored. The type of dest should be a pointer to an object that matches the expected data structure.
	//
	// Returns:
	//   - error: If an error occurs during the request, it will be returned. Otherwise, nil is returned.
	get(ctx context.Context, url string, dest any) error
}

type restyHttpClient struct {
	httpClient *resty.Client
}

func newRestyHttpClient() *restyHttpClient {
	httpClient := resty.New()
	httpClient.
		SetRetryCount(3).
		SetRetryWaitTime(500 * time.Millisecond).
		SetRetryAfter(func(client *resty.Client, resp *resty.Response) (time.Duration, error) {
			return 0, fmt.Errorf("API call limit exceeded: you have reached the maximum number of allowed attempts(%d)", client.RetryCount)
		})

	return &restyHttpClient{httpClient: httpClient}
}

func (r *restyHttpClient) get(ctx context.Context, url string, dest any) error {
	if reflect.ValueOf(dest).Kind() != reflect.Ptr {
		return fmt.Errorf("expected a pointer for 'dest', but got %s", reflect.TypeOf(dest))
	}

	req := r.httpClient.R()
	resp, err := req.SetResult(dest).Get(url)
	if err != nil {
		return fmt.Errorf("failed to send GET request to %s: %w", url, err)
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("API request to %s returned status code %d; expected %d (OK)", resp.Request.URL, resp.StatusCode(), http.StatusOK)
	}

	return nil
}
