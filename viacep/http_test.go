package viacep

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestViaCep_HttpClient_Get(t *testing.T) {
	t.Run("successful GET request", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
			assert.Equal(t, "application/json", r.Header.Get("Accept"))

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"key": "value"}`))
		}))
		defer srv.Close()

		client := NewHTTPClient(1)

		dest := map[string]string{}
		err := client.Get(context.Background(), srv.URL, &dest)
		assert.NoError(t, err)
		assert.Equal(t, map[string]string{"key": "value"}, dest)
	})

	t.Run("invalid dest type", func(t *testing.T) {
		client := NewHTTPClient(1)

		invalidDest := "string_instead_of_pointer"
		err := client.Get(context.Background(), "http://", invalidDest)
		assert.EqualError(t, err, "expected a pointer for 'dest', but got string")
	})

	t.Run("non-ok status code", func(t *testing.T) {
		errorServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))

		defer errorServer.Close()

		client := NewHTTPClient(1)

		dest := map[string]string{}
		err := client.Get(context.Background(), errorServer.URL, &dest)
		assert.EqualError(t, err, fmt.Sprintf("API request to %s returned status code 500; expected 200 (OK)", errorServer.URL))
	})

	t.Run("HTTP request error", func(t *testing.T) {
		client := NewHTTPClient(0)
		url := "httpdd://invalid-url"
		dest := map[string]string{}

		err := client.Get(context.Background(), url, &dest)
		assert.EqualError(t, err, fmt.Sprintf("failed to send GET request to %s: Get \"httpdd://invalid-url\": unsupported protocol scheme \"httpdd\"", url))
	})

	t.Run("timeout", func(t *testing.T) {
		errorServer := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
			time.Sleep(10 * time.Millisecond)
		}))
		defer errorServer.Close()

		client := NewHTTPClient(0)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
		defer cancel()

		dest := map[string]string{}
		err := client.Get(ctx, errorServer.URL, &dest)
		assert.EqualError(t, err, fmt.Sprintf("failed to send GET request to %s: Get %q: context deadline exceeded", errorServer.URL, errorServer.URL))
	})
}
