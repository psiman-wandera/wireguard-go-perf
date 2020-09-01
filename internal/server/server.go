package server

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"

)

// ListenAndServe starts server based on provided configuration and registers request handlers.
func ListenAndServe() error {
	return http.ListenAndServe(fmt.Sprintf(":%d", 8079), nil)
}
