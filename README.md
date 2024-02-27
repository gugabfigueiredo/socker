# Socker 
a stupid simple golang mock server

## Usage

1. Create a new server
```go
server := socker.NewServer()
defer server.Close()
```
2. Setup your mock responses
```go
server.On("/something").RespondStatus(http.StatusOK)
```

3. Run your tests
```go
res, err := http.Get(server.URL + "/something")
assert.NoError(t, err)
assert.Equal(t, http.StatusOK, res.StatusCode)
```

### Setup

`On` will match any request to `/something`
```go
server.On("/something").RespondStatus(http.StatusOK)
```

`OnMethod` will match a requests to `/something` with the given method
```go
server.OnMethod(http.MethodGet, "/something").RespondStatus(http.StatusOK)
```

`OnRequest` will match strictly the request with same method, path, query and given header configuration.
```go
server.OnRequest(&http.Request{
    Method: http.MethodPost,
    URL: &url.URL{
        Path: "/something",
        RawQuery: "param=value",
    },
    Header: http.Header{
        "X-Header": []string{"value"}, // this will match only if the request has this header
        "-Unwanted-Header": []string{"value"}, // this will match only if the request does not have this header
    },
}).RespondStatus(http.StatusOK)
```
#### Wildcards
You can use wildcards to match any path that starts with a specific string.
This will match to any request to an endpoint that starts with `/something`
```go
server.On("/something/*").RespondStatus(http.StatusOK)
```
Works with `OnMethods` and `OnRequests` as well
```go
server.OnMethod(http.MethodGet, "/something/*").RespondStatus(http.StatusOK)
server.OnRequest(&http.Request{
    Method: http.MethodPost,
    URL: &url.URL{
        Path: "/something/*",
    },
}).RespondStatus(http.StatusOK)
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