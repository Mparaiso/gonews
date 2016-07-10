package gonews

import (
	"github.com/gorilla/sessions"
	"net/http"
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
type ResponseWriterExtraInterface interface {
	http.ResponseWriter
	IsResponseWritten() bool
	SetSession(SessionInterface)
	Session() SessionInterface
	HasSession() bool
	GetCurrentSize() int
	SetLogger(LoggerInterface)
}

// SessionInterface was extracted from Session
type SessionInterface interface {
	AddFlash(value interface{}, vars ...string)
	Flashes(vars ...string) []interface{}
	Name() string
	Save(r *http.Request, w http.ResponseWriter) error
	Store() sessions.Store
	Get(Any) Any
	Set(Any, Any)
	Has(Any) bool
	Options() *sessions.Options
	SetOptions(*sessions.Options)
	Values() map[string]interface{}
}

// Form interface is a form
type Form interface {
	// HandleRequest deserialize the request body into a form struct
	HandleRequest(r *http.Request) error
}

// CSRFProvider provide csrf tokens
type CSRFProvider interface {
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
