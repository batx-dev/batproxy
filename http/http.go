package http

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/batx-dev/batproxy"
	"github.com/gorilla/schema"
)

var (
	decoder = schema.NewDecoder()
	encoder = schema.NewEncoder()
)

// Error prints & optionally logs an error message.
func Error(w http.ResponseWriter, req *http.Request, err error) {
	code, message := batproxy.ErrorCode(err), batproxy.ErrorMessage(err)
	w.WriteHeader(ErrorStatusCode(code))
	_, _ = w.Write([]byte(message))
}

// ErrorResponse represents a JSON structure for error output.
type ErrorResponse struct {
	Error string `json:"error"`
}

// parseResponseError parses an JSON-formatted error response.
func parseResponseError(res *http.Response) error {
	defer res.Body.Close()

	buf, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	var errRes ErrorResponse
	if err := json.Unmarshal(buf, &errRes); err != nil {
		message := strings.TrimSpace(string(buf))
		if message == "" {
			message = "Empty response from server."
		}
		return batproxy.Errorf(FromErrorStatusCode(res.StatusCode), message)
	}
	return batproxy.Errorf(FromErrorStatusCode(res.StatusCode), errRes.Error)
}

// lookup of application error codes to HTTP status codes.
var codes = map[string]int{
	batproxy.ECONFLICT:       http.StatusConflict,
	batproxy.EINVALID:        http.StatusBadRequest,
	batproxy.ENOTFOUND:       http.StatusNotFound,
	batproxy.ENOTIMPLEMENTED: http.StatusNotImplemented,
	batproxy.EUNAUTHORIZED:   http.StatusUnauthorized,
	batproxy.EINTERNAL:       http.StatusInternalServerError,
	batproxy.EFORBIDDEN:      http.StatusForbidden,
}

// ErrorStatusCode returns the associated HTTP status code for a BatProxy error code.
func ErrorStatusCode(code string) int {
	if v, ok := codes[code]; ok {
		return v
	}
	return http.StatusInternalServerError
}

// FromErrorStatusCode returns the associated BatProxy code for an HTTP status code.
func FromErrorStatusCode(code int) string {
	for k, v := range codes {
		if v == code {
			return k
		}
	}
	return batproxy.EINTERNAL
}
