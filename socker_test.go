package socker

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/url"
	"testing"
)

func TestMocli(t *testing.T) {

	socker := NewMockServer()
	socker.On(http.MethodGet, "/a/b/c").RespondStatus(http.StatusOK)
	socker.On(http.MethodPost, "/a/b/c").RespondStatus(http.StatusBadRequest)
	socker.OnAny("/a/b/c/d").RespondStatus(http.StatusMultiStatus)
	socker.OnRequest(&http.Request{
		Method: http.MethodGet,
		URL:    &url.URL{Path: "/a/request"},
	}).RespondStatus(http.StatusOK)
	socker.Start()
	defer socker.Stop()

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
}
