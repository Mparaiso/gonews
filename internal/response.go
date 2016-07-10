package gonews

import (
	"fmt"
	"net/http"
	"sync"
)

// ResponseWriterExtra can notify if a response has been written
type ResponseWriterExtra struct {
	http.ResponseWriter
	Request            *http.Request
	sessionInterface   SessionInterface
	hasWrittenResponse bool
	currentSize        int
	logger             LoggerInterface
	sync.Once
}

// Session returns a SessionInterface
func (rw *ResponseWriterExtra) Session() SessionInterface {
	return rw.sessionInterface
}

// SetLogger sets the logger
func (rw *ResponseWriterExtra) SetLogger(logger LoggerInterface) {
	rw.logger = logger
}

// SetSession sets the session
func (rw *ResponseWriterExtra) SetSession(session SessionInterface) {
	rw.sessionInterface = session
}

// HasSession returns true if rw has a session
func (rw *ResponseWriterExtra) HasSession() bool {
	return rw.sessionInterface != nil
}

// SetWrittenResponse returns true if a response has been written
func (rw *ResponseWriterExtra) SetWrittenResponse() {
	rw.hasWrittenResponse = true
}

func (rw *ResponseWriterExtra) error(messages ...interface{}) {
	if rw.logger != nil {
		rw.logger.Error(append([]interface{}{"ResponseWithExtra.Write"}, messages...)...)
	}
}
func (rw *ResponseWriterExtra) debug(messages ...interface{}) {
	if rw.logger != nil {
		rw.logger.Debug(append([]interface{}{"ResponseWithExtra.Write"}, messages...)...)
	}
}

// Write writes in the response stream
func (rw *ResponseWriterExtra) Write(b []byte) (size int, err error) {
	rw.Once.Do(func() {
		if rw.HasSession() {
			rw.debug("Trying to save the current session")
			err := rw.Session().Save(rw.Request, rw.ResponseWriter)
			if err != nil {
				rw.error("Error saving the session ", err)
			}
		} else {
			rw.error("Session not found, can't save... ")
		}
	})

	size, err = rw.ResponseWriter.Write(b)
	if err != nil {
		rw.error(err)
	}
	rw.currentSize += size
	return
}

// GetCurrentSize get size written in response
func (rw *ResponseWriterExtra) GetCurrentSize() int {
	return rw.currentSize
}

// IsResponseWritten returns true if Write has been called
func (rw *ResponseWriterExtra) IsResponseWritten() bool {
	return rw.hasWrittenResponse
}

// WriteHeader writes the status code
func (rw *ResponseWriterExtra) WriteHeader(status int) {
	rw.Header().Set("Status-Code", fmt.Sprintf("%d", status))
	rw.ResponseWriter.WriteHeader(status)
}
