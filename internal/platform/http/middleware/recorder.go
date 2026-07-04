package middleware

import (
	"bytes"
	"net/http"
)

type responseRecorder struct {
	http.ResponseWriter
	status int
	err    error
	body   bytes.Buffer
}

func (r *responseRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *responseRecorder) Write(p []byte) (int, error) {
	_, _ = r.body.Write(p)
	return r.ResponseWriter.Write(p)
}

func (r *responseRecorder) setError(err error) { r.err = err }
