# Socker - a stupid simple golang mock server

## Usage

1. Create a new server
```go
server := socker.NewServer()
defer server.Close()
```
2. Configure your mock responses
```go
server.OnAny("/something").RespondStatus(http.StatusOK)
```

3. Run your tests
```go
res, err := http.Get(server.URL + "/something")
assert.NoError(t, err)
assert.Equal(t, http.StatusOK, res.StatusCode)
```

### Pattern Matching

You can be as broad or specific as you want with the path matching. Socker will try to match the more specific first.

This will match to any request to `/something`
```go
server.OnAny("/something").RespondStatus(http.StatusOK)
```

This will match to GET requests to an endpoint that starts with `/something`
```go
server.On(http.MethodGet, "/something").RespondStatus(http.StatusOK)
```

This will match any request with same method and path, but will return bad request if query parameters and headers are missing
```go
server.OnRequest(&http.Request{
    Method: http.MethodPost,
    URL: &url.URL{
        Path: "/something",
        RawQuery: "param=value",
    },
    Header: http.Header{
        "X-Header": []string{"value"},
    },
}).RespondStatus(http.StatusOK)
```
> **Note**: Future calls with same method and path will override any of the previously discussed configurations.

### Strict Matching

This will match the request exactly as it is, including headers and query parameters.
```go
server.OnRequestStrict(&http.Request{
    Method: http.MethodPost,
    URL: &url.URL{
        Path: "/something",
        RawQuery: "param=value",
    },
    Header: http.Header{
        "X-Header": []string{"value"},
    },
}).RespondStatus(http.StatusOK)
```