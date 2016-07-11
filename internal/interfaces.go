package gonews

import (
	"net/http"

	"github.com/gorilla/sessions"
)

// LoggerInterface defines a logger
type LoggerInterface interface {
	Debug(messages ...interface{})
	Info(messages ...interface{})
	Error(messages ...interface{})
	ErrorWithStack(messages ...interface{})
}

// ResponseWriterExtraInterface is a response writer
// enhanced with various apis
type ResponseWriterExtra interface {
	http.ResponseWriter
	IsResponseWritten() bool
	SetSession(SessionWrapper)
	Session() SessionWrapper
	HasSession() bool
	GetCurrentSize() int
	SetLogger(LoggerInterface)
}

// SessionInterface was extracted from Session
type SessionWrapper interface {
	AddFlash(value interface{}, vars ...string)
	Flashes(vars ...string) []interface{}
	Name() string
	Save(r *http.Request, w http.ResponseWriter) error
	Store() sessions.Store
	Get(interface{}) interface{}
	Set(interface{}, interface{})
	Has(interface{}) bool
	Options() *sessions.Options
	SetOptions(*sessions.Options)
	Values() map[interface{}]interface{}
	ValuesString() map[string]interface{}
	Delete(interface{})
}

// Form interface is a form
type Form interface {
	// HandleRequest deserialize the request body into a form struct
	HandleRequest(r *http.Request) error
}

// CSRFGenerator generates and validate csrf tokens
type CSRFGenerator interface {
	Generate(userID, actionID string) string
	Valid(token, userID, actionID string) bool
}

// UserFinder can find users from a datasource
type UserFinder interface {
	GetOneByEmail(string) (*User, error)
	GetOneByUsername(string) (*User, error)
}

// ValidationError is a validation error
type ValidationError interface {
	HasErrors() bool
	Append(key, value string)
	Error() string
}
