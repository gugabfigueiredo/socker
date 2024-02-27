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
> server.On("/something").RespondStatus(http.StatusAccepted) // this will override the previous On("/something")
> server.OnMethod(http.MethodGet, "/something").RespondStatus(http.StatusOK)
> server.OnMethod(http.MethodGet, "/something").RespondStatus(http.StatusAccepted) // this will override the previous OnMethod(http.MethodGet, "/something")
> server.OnRequest(&http.Request{Method: http.MethodPost, URL: &url.URL{Path: "/something"}}).RespondStatus(http.StatusOK)
> server.OnRequest(&http.Request{Method: http.MethodPost, URL: &url.URL{Path: "/something"}}).RespondStatus(http.StatusAccepted) // this will override the previous OnRequest
> ...
>```

### Request Matching
socker will try to match any incoming request to the most specific setting in this order:

`OnRequest` + wildcards > `OnMethod` + wildcards > `On` + wildcards

The first match will be used to respond to the request.

This means that given the following setup:
```go
server.On("/something/*").RespondStatus(http.StatusOK)
server.On("/*").RespondStatus(http.StatusAccepted)
```
A request to `/something/else` will respond with `http.StatusOK` and a request to `/anything` will respond with `http.StatusAccepted`

### Responding
Use `Respond` with a `Responder` to respond with a custom response
```go
server.On("/something").Respond(socker.Responder{
    Status: http.StatusOK,
    Headers: http.Header{
        "X-Header": []string{"value"},
        "Content-Type": []string{"text/plain"},
    },
    Body: "Hello, World!",
})
```

You can also provide your own `http.HandlerFunc` to respond to the request
```go
server.On("/something").RespondWith(func(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("Hello, World!"))
})
```
But there are some helper methods to respond with
```go
server.On("/something").RespondStatus(http.StatusOK)
server.On("/something").RespondJSON(map[string]interface{}{"key": "value"})
```
