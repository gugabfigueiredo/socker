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