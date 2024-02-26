package socker

import (
	"bytes"
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/url"
	"testing"
	"time"
)

func TestSocker(t *testing.T) {

	tt := []struct {
		name             string
		setup            func(server *MockServer)
		check            func(url string) (*http.Response, error)
		expectedResponse *http.Response
		expectedError    error
	}{
		{
			name: "on any - get and post - response status ok",
			setup: func(server *MockServer) {
				server.OnAny("/any").RespondStatus(http.StatusOK)
			},
			check: func(serverURL string) (*http.Response, error) {
				response, err := http.Get(serverURL + "/any")
				if err != nil || response.StatusCode != http.StatusOK {
					return response, fmt.Errorf("expected status %d, got %d", http.StatusOK, response.StatusCode)
				}
				return http.Post(serverURL+"/any", "application/json", nil)
			},
			expectedResponse: &http.Response{
				StatusCode: http.StatusOK,
			},
		},
		{
			name: "on get - get - response status ok",
			setup: func(server *MockServer) {
				server.On(http.MethodGet, "/get").RespondStatus(http.StatusOK)
			},
			check: func(serverURL string) (*http.Response, error) {
				return http.Get(serverURL + "/get")
			},
			expectedResponse: &http.Response{
				StatusCode: http.StatusOK,
			},
		},
		{
			name: "on get - post - response status not found",
			setup: func(server *MockServer) {
				server.On(http.MethodGet, "/get").RespondStatus(http.StatusOK)
			},
			check: func(serverURL string) (*http.Response, error) {
				return http.Post(serverURL+"/get", "application/json", nil)
			},
			expectedResponse: &http.Response{
				StatusCode: http.StatusNotFound,
			},
		},
		{
			name: "on get request - get matching request - response status ok",
			setup: func(server *MockServer) {
				server.OnRequest(&http.Request{
					Method: http.MethodGet,
					URL:    &url.URL{Path: "/a/request"},
					Header: http.Header{
						"Content-Type": []string{"application/json"},
						"X-Request-Id": []string{"123"},
					},
				}).RespondStatus(http.StatusOK)
			},
			check: func(serverURL string) (*http.Response, error) {
				req, _ := http.NewRequest(http.MethodGet, serverURL+"/a/request", bytes.NewBufferString(`{"foo":"bar"}`))
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("X-Request-Id", "123")
				return http.DefaultClient.Do(req)
			},
			expectedResponse: &http.Response{
				StatusCode: http.StatusOK,
			},
		},
		{
			name: "on get request - get request missing data - response status bad request",
			setup: func(server *MockServer) {
				server.OnRequest(&http.Request{
					Method: http.MethodGet,
					URL:    &url.URL{Path: "/a/request"},
					Header: http.Header{
						"Content-Type": []string{"application/json"},
						"X-Request-Id": []string{"123"},
					},
				}).RespondStatus(http.StatusOK)
			},
			check: func(serverURL string) (*http.Response, error) {
				req, _ := http.NewRequest(http.MethodGet, serverURL+"/a/request", bytes.NewBufferString(`{"foo":"bar"}`))
				req.Header.Set("Content-Type", "application/json")
				return http.DefaultClient.Do(req)
			},
			expectedResponse: &http.Response{
				StatusCode: http.StatusBadRequest,
			},
		},
		{
			name: "on get request - post request - response status not found",
			setup: func(server *MockServer) {
				server.OnRequest(&http.Request{
					Method: http.MethodGet,
					URL:    &url.URL{Path: "/a/request"},
					Header: http.Header{
						"Content-Type": []string{"application/json"},
						"X-Request-Id": []string{"123"},
					},
				}).RespondStatus(http.StatusOK)
			},
			check: func(serverURL string) (*http.Response, error) {
				req, _ := http.NewRequest(http.MethodPost, serverURL+"/a/request", bytes.NewBufferString(`{"foo":"bar"}`))
				req.Header.Set("Content-Type", "application/json")
				return http.DefaultClient.Do(req)
			},
			expectedResponse: &http.Response{
				StatusCode: http.StatusNotFound,
			},
		},
		{
			name: "on any - get - response json body",
			setup: func(server *MockServer) {
				server.OnAny("/any").RespondJSON(http.StatusOK, map[string]string{"foo": "bar"})
			},
			check: func(serverURL string) (*http.Response, error) {
				return http.Get(serverURL + "/any")
			},
			expectedResponse: &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "on any - get - response error",
			setup: func(server *MockServer) {
				server.OnAny("/any").RespondError(http.StatusBadRequest, "bad request")
			},
			check: func(serverURL string) (*http.Response, error) {
				return http.Get(serverURL + "/any")
			},
			expectedResponse: &http.Response{
				StatusCode: http.StatusBadRequest,
			},
		},
		{
			name: "on any - get - timeout",
			setup: func(server *MockServer) {
				server.OnAny("/any").Timeout(200 * time.Millisecond)
			},
			check: func(serverURL string) (*http.Response, error) {
				ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
				defer cancel()
				req, _ := http.NewRequestWithContext(
					ctx,
					http.MethodGet,
					serverURL+"/any",
					nil,
				)
				return http.DefaultClient.Do(req)
			},
			expectedError: context.DeadlineExceeded,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			socker := NewServer()
			defer socker.Close()
			tc.setup(socker)
			res, err := tc.check(socker.URL)
			if tc.expectedError != nil {
				assert.ErrorIs(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
			}
			if tc.expectedResponse != nil {
				assert.Equal(t, tc.expectedResponse.StatusCode, res.StatusCode)
			}
		})
	}
}
