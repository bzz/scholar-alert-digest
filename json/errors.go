package json

import (
	"encoding/json"
	"net/http"
)

// ErrResponse presents error information to JSON API user.
type ErrResponse struct {
	Err            string `json:"error,omitempty"` // low-level error
	HTTPStatusCode int    `json:"-"`               // http response status code

	StatusText string `json:"status"` // user-level status message
	Redirect   string `json:",omitempty"`
}

func jsonError(w http.ResponseWriter, status int, er ErrResponse) {
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": er,
	})
}

func ErrNotFound(w http.ResponseWriter, err error, msg string) {
	jsonError(w, http.StatusNotFound, ErrResponse{
		Err:        err.Error(),
		StatusText: msg,
	})
	return
}

func ErrUnprocessable(w http.ResponseWriter, err error, msg string) {
	jsonError(w, http.StatusUnprocessableEntity, ErrResponse{
		Err:        err.Error(),
		StatusText: msg,
	})
	return
}

func ErrUnauthorized(w http.ResponseWriter, url string) {
	status := http.StatusUnauthorized
	jsonError(w, status, ErrResponse{
		StatusText: http.StatusText(status),
		Redirect:   url,
	})
	return
}

func ErrFailedDependency(w http.ResponseWriter, err error, msg string) {
	jsonError(w, http.StatusFailedDependency, ErrResponse{
		Err:        err.Error(),
		StatusText: msg,
	})
	return
}
