package socker

import (
	"bytes"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/url"
	"testing"
)

func TestMocli(t *testing.T) {

	socker := NewServer()
	defer socker.Stop()

	socker.On(http.MethodGet, "/a/b/c").RespondStatus(http.StatusOK)
	socker.On(http.MethodPost, "/a/b/c").RespondStatus(http.StatusBadRequest)
	socker.OnAny("/a/b/c/d").RespondStatus(http.StatusMultiStatus)
	socker.OnRequest(&http.Request{
		Method: http.MethodGet,
		URL:    &url.URL{Path: "/a/request"},
	}).RespondStatus(http.StatusOK)

	res, err := http.Get(socker.URL + "/a/b/c")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)

	res, err = http.Post(socker.URL+"/a/b/c", "application/json", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)

	res, err = http.Get(socker.URL + "/a/b/c/d")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusMultiStatus, res.StatusCode)

	res, err = http.Post(socker.URL+"/a/b/c/d", "application/json", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusMultiStatus, res.StatusCode)

	res, err = http.Get(socker.URL + "/a/request")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)

	res, err = http.Post(socker.URL+"/a/request", "application/json", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, res.StatusCode)

	socker.On(http.MethodGet, "/in/run/time").RespondJSON(http.StatusOK, map[string]string{"time": "now"})
	res, err = http.Get(socker.URL + "/in/run/time")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
	defer res.Body.Close()
	jsonBody := &json.RawMessage{}
	_ = json.NewDecoder(res.Body).Decode(jsonBody)
	jsonString, _ := jsonBody.MarshalJSON()
	assert.Equal(t, `{"time":"now"}`, string(jsonString))

	socker.OnRequestStrict(&http.Request{
		Method: http.MethodGet,
		URL:    &url.URL{Path: "/a/request"},
		Header: http.Header{
			"Content-Type":  []string{"application/json"},
			"X-Request-ID":  []string{"123"},
			"Authorization": []string{"Bearer token"},
		},
		Body: io.NopCloser(bytes.NewBufferString(`{"foo":"bar"}`)),
	}).RespondStatus(http.StatusNotImplemented)
	req, _ := http.NewRequest(http.MethodGet, socker.URL+"/a/request", bytes.NewBufferString(`{"foo":"bar"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Request-ID", "123")
	req.Header.Set("Authorization", "Bearer token")
	res, err = http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotImplemented, res.StatusCode)
}
