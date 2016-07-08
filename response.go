package gonews

import (
	"fmt"
	"net/http"
)

type ResponseWriterExtraInterface interface {
	http.ResponseWriter
	IsResponseWritten() bool
	SetSession(SessionInterface)
	Session() SessionInterface
	HasSession() bool
	GetCurrentSize() int
}

// ResponseWriterExtra can notify if a response has been written
type ResponseWriterExtra struct {
	http.ResponseWriter
	Request            *http.Request
	sessionInterface   SessionInterface
	hasWrittenResponse bool
	currentSize        int
}

func (rw *ResponseWriterExtra) Session() SessionInterface {
	return rw.sessionInterface
}

func (rw *ResponseWriterExtra) SetSession(session SessionInterface) {
	rw.sessionInterface = session
}

func (rw *ResponseWriterExtra) HasSession() bool {
	return rw.sessionInterface != nil
}

func (rw *ResponseWriterExtra) SetWrittenResponse() {
	rw.hasWrittenResponse = true
}

// Write writes in the response stream
func (rw ResponseWriterExtra) Write(b []byte) (size int, err error) {
	if rw.HasSession() {
		rw.Session().Save(rw.Request, rw.ResponseWriter)
	}
	size, err = rw.ResponseWriter.Write(b)
	rw.currentSize += size
	return
}

func (rw ResponseWriterExtra) GetCurrentSize() int {
	return rw.currentSize
}

// IsResponseWritten returns true if Write has been called
func (rw *ResponseWriterExtra) IsResponseWritten() bool {
	return rw.hasWrittenResponse
}

// WriteHeader writes the status code
func (rw ResponseWriterExtra) WriteHeader(status int) {
	rw.Header().Set("Status-Code", fmt.Sprintf("%d", status))
	rw.ResponseWriter.WriteHeader(status)
}
