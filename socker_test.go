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
				server.On("/any").RespondStatus(http.StatusOK)
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
				server.OnMethod(http.MethodGet, "/get").RespondStatus(http.StatusOK)
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
				server.OnMethod(http.MethodGet, "/get").RespondStatus(http.StatusOK)
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
				StatusCode: http.StatusNotFound,
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
				server.On("/any").RespondJSON(http.StatusOK, map[string]string{"foo": "bar"})
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
				server.On("/any").RespondError(http.StatusBadRequest, "bad request")
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
				server.On("/any").Timeout(200 * time.Millisecond)
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
		{
			name: "on get wildcard - get - response status ok",
			setup: func(server *MockServer) {
				server.OnMethod(http.MethodGet, "/wild/*").RespondStatus(http.StatusOK)
			},
			check: func(serverURL string) (*http.Response, error) {
				return http.Get(serverURL + "/wild/anything")
			},
			expectedResponse: &http.Response{
				StatusCode: http.StatusOK,
			},
		},
		{
			name: "on get wildcard multiple choice - get most specific - response status ok",
			setup: func(server *MockServer) {
				server.OnMethod(http.MethodGet, "/wild/*").RespondStatus(http.StatusMultipleChoices)
				server.OnMethod(http.MethodGet, "/wild/anything").RespondStatus(http.StatusOK)
			},
			check: func(serverURL string) (*http.Response, error) {
				return http.Get(serverURL + "/wild/anything")
			},
			expectedResponse: &http.Response{
				StatusCode: http.StatusOK,
			},
		},
		{
			name: "on multiple choice - get - respond with most specific",
			setup: func(server *MockServer) {
				server.On("/multiple").RespondStatus(http.StatusOK)
				server.OnMethod(http.MethodGet, "/multiple").RespondStatus(http.StatusOK)
				server.OnRequest(&http.Request{
					Method: http.MethodGet,
					URL:    &url.URL{Path: "/multiple", RawQuery: "foo=bar"},
				}).RespondStatus(http.StatusAccepted)
			},
			check: func(serverURL string) (*http.Response, error) {
				return http.Get(serverURL + "/multiple?foo=bar")
			},
			expectedResponse: &http.Response{
				StatusCode: http.StatusAccepted,
			},
		},
		{
			name: "on multiple choice - get - respond with least specific",
			setup: func(server *MockServer) {
				server.On("/multiple").RespondStatus(http.StatusAccepted)
				server.OnMethod(http.MethodGet, "/multiple").RespondStatus(http.StatusOK)
				server.OnRequest(&http.Request{
					Method: http.MethodGet,
					URL:    &url.URL{Path: "/multiple", RawQuery: "foo=bar"},
				}).RespondStatus(http.StatusOK)
			},
			check: func(serverURL string) (*http.Response, error) {
				return http.Post(serverURL+"/multiple", "application/json", nil)
			},
			expectedResponse: &http.Response{
				StatusCode: http.StatusAccepted,
			},
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
			} else {
				assert.Nil(t, res)
			}
		})
	}
}

func TestSockerOnPort(t *testing.T) {
	socker := NewServerOnPort("8080")
	defer socker.Close()
	socker.On("/foo").RespondStatus(http.StatusOK)

	res, err := http.Get("http://127.0.0.1:8080" + "/foo")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)
}

func TestSocker_Router(t *testing.T) {

	tt := []struct {
		name             string
		setup            func(server *MockServer)
		requestPath      string
		expectedRoutes   []string
		expectedResponse *http.Response
		expectedError    error
	}{
		{
			name: "on shallow route - get - response status ok",
			setup: func(server *MockServer) {
				server.OnRoute("/shallow", func(m *MockServer) {
					m.On("/").RespondStatus(http.StatusOK)
				})
			},
			requestPath:    "/shallow",
			expectedRoutes: []string{"/shallow"},
			expectedResponse: &http.Response{
				StatusCode: http.StatusOK,
			},
		},
		{
			name: "on deep route - get - response status ok",
			setup: func(server *MockServer) {
				server.OnRoute("/deep", func(m *MockServer) {
					m.OnRoute("/nested", func(m *MockServer) {
						m.On("/path").RespondStatus(http.StatusOK)
					})
				})
			},
			requestPath:    "/deep/nested/path",
			expectedRoutes: []string{"/deep/nested/path"},
			expectedResponse: &http.Response{
				StatusCode: http.StatusOK,
			},
		},
		{
			name: "on multiple routes - get - response status ok",
			setup: func(server *MockServer) {
				server.OnRoute("/multiple", func(m *MockServer) {
					m.On("/one").RespondStatus(http.StatusOK)
					m.On("/two").RespondStatus(http.StatusOK)
				})
			},
			requestPath:    "/multiple/two",
			expectedRoutes: []string{"/multiple/two", "/multiple/one"},
			expectedResponse: &http.Response{
				StatusCode: http.StatusOK,
			},
		},
		{
			name: "on mixed routes - get - response status ok",
			setup: func(server *MockServer) {
				server.On("/mixed").RespondStatus(http.StatusOK)
				server.OnRoute("/mixed", func(m *MockServer) {
					m.On("/one").RespondStatus(http.StatusOK)
					m.OnRoute("/two", func(m *MockServer) {
						m.On("/three").RespondStatus(http.StatusOK)
					})
				})
			},
			requestPath:    "/mixed/two/three",
			expectedRoutes: []string{"/mixed/two/three", "/mixed/one", "/mixed"},
			expectedResponse: &http.Response{
				StatusCode: http.StatusOK,
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			socker := NewServer()
			defer socker.Close()
			tc.setup(socker)
			keys := make([]string, 0, len(socker.handlers))
			for k := range socker.handlers {
				keys = append(keys, k)
			}
			assert.ElementsMatch(t, tc.expectedRoutes, keys)
			resp, err := http.Get(socker.URL + tc.requestPath)
			if tc.expectedError != nil {
				assert.EqualError(t, err, tc.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			if tc.expectedResponse != nil {
				assert.Equal(t, tc.expectedResponse.StatusCode, resp.StatusCode)
			} else {
				assert.Nil(t, resp)
			}
		})
	}
}
