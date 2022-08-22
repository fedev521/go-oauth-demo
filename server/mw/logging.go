package mw

import (
	"io"
	"net/http"

	"github.com/gorilla/handlers"
)

// loggingMiddleware returns a middleware that logs information about HTTP
// request and response.
func loggingMiddleware(out io.Writer) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return handlers.LoggingHandler(out, next)
	}
}
