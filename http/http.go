package http

import (
	"fmt"
	"net/http"

	"github.com/gorilla/schema"
)

var (
	decoder = schema.NewDecoder()
	encoder = schema.NewEncoder()
)

// Error prints & optionally logs an error message.
func Error(w http.ResponseWriter, req *http.Request, err error) {
	fmt.Fprintf(w, "%s", err)
}
