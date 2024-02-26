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
server.On("/something").RespondStatus(http.StatusOK)
```

This will match to GET requests to `/something`
```go
server.OnMethod(http.MethodGet, "/something").RespondStatus(http.StatusOK)
```

This will match strictly the request with same method, path, query and present headers. anything else will be ignored and passed on to be consumed by another setting.
```go
server.OnRequest(&http.Request{
    Method: http.MethodPost,
    URL: &url.URL{
        Path: "/something", // wildcards are treated as literals here
        RawQuery: "param=value",
    },
    Header: http.Header{
        "X-Header": []string{"value"},
    },
}).RespondStatus(http.StatusOK)
```

You can use wildcards to match any path that starts with a specific string.
This will match to any request to an endpoint that starts with `/something`
```go
server.On("/something/*").RespondStatus(http.StatusOK)
```
###
> **Note**: Multiple calls of `On()`, `OnMethod()` or `OnRequest()` with same method and path will override existing settings created by the same methods.
> ```go
> server.On("/something").RespondStatus(http.StatusOK)
> server.On("/something").RespondStatus(http.StatusNotFound) // this will override the previous setting
> server.OnMethod(http.MethodGet, "/something").RespondStatus(http.StatusOK) // this will not
> server.OnMethod(http.MethodGet, "/something").RespondStatus(http.StatusNotFound) // this will!
> server.OnRequest(&http.Request{Method: http.MethodPost, URL: &url.URL{Path: "/something"}}).RespondStatus(http.StatusOK) // this will not
> ...
>```