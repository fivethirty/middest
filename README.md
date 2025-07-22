# middest

![test status](https://github.com/fivethirty/middest/actions/workflows/test.yml/badge.svg)

Go middleware collection created with hypermedia applications in mind.

## Modules

- `ctxlog`: Structured request logging
- `recoverer`: Panic recovery with 
- `requestsize`: Request body size limiting
- `contenttype`: Content type filtering
- `errs`: Error handling utilities
- `handlers`: Helper for composing middleware

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
