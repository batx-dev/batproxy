package http

import (
	"fmt"
	"net/http"
)

// Error prints & optionally logs an error message.
func Error(w http.ResponseWriter, req *http.Request, err error) {
	fmt.Fprintf(w, "%s", err)
}
