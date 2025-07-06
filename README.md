# middest

![test status](https://github.com/fivethirty/middest/actions/workflows/test.yml/badge.svg)

Go middleware collection created with hypermedia applications in mind.

## Modules

- `ctxlog`: Structured request logging with automatic request ID generation
- `recoverer`: Graceful panic recovery with stack trace logging
- `requestsize`: Request body size limiting
- `contenttype`: Content type handling
- `errs`: Error handling utilities
- `handlers`: Helper for composing multiple middleware

## Usage

```go
package main

import (
    "net/http"
    "github.com/fivethirty/middest/handlers"
    "github.com/fivethirty/middest/ctxlog"
    "github.com/fivethirty/middest/recoverer"
)

func main() {
    logger := ctxlog.DefaultLogger
    
    middlewares := []func(http.Handler) http.Handler{
        recoverer.New(logger),
        ctxlog.New(logger),
    }
    
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("Hello, World!"))
    })
    
    http.HandleFunc("/", handlers.WithMiddleware(middlewares, handler))
    http.ListenAndServe(":8080", nil)
}
```
